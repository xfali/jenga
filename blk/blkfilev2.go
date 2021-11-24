// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xfali/jenga/compressor"
	"github.com/xfali/jenga/flags"
	"io"
	"os"
)

const (
	BlkFileV2Version uint16 = 0x0002
)

// File format:
// |MAGIC NUNMBER(2 Bytes)|VERSION(2 Bytes)|DATA FORMAT(2 Bytes)|REVERSE(2 Bytes)|ENTITY_1|ENTITY_2|...|ENTITY_N|
// Entity format:
// |VARINT(1-10 Bytes)|STRING(string length)|DATA SIZE(8 Bytes)|DATA(data size)|
type BlkFileV2 struct {
	file       BlockReadWriter
	opener     Opener
	cur        int64
	header     FileHeader
	compressor compressor.Compressor
}

func NewBlkFileV2(path string) *BlkFileV2 {
	return NewBlkFileV2WithOpener(BlkFileV2Openers.Local(path))
}

func NewBlkFileV2WithOpener(opener Opener) *BlkFileV2 {
	return &BlkFileV2{
		opener: opener,
		header: FileHeader{
			Version:    BlkFileV2Version,
			DataFormat: compressor.TypeNone,
		},
		cur:        0,
		compressor: nil,
	}
}

func (bf *BlkFileV2) WithCompressor(compressor compressor.Compressor) *BlkFileV2 {
	if compressor != nil {
		bf.compressor = compressor
	}
	return bf
}

func (bf *BlkFileV2) Open(flag flags.OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return errors.New("Tar format flag cannot contains both OpFlagReadOnly and OpFlagWriteOnly. ")
	}
	f, new, err := bf.opener(flag)
	if err != nil {
		return err
	}
	bf.file = f
	if !new {
		if flag.CanRead() {
			bf.cur = BlkFileHeadSize
			err = bf.readHeader()
			if err != nil {
				_ = f.Close()
				return err
			}
			return nil
		} else if flag.CanWrite() {
			err = bf.readHeader()
			if err != nil {
				_ = f.Close()
				return err
			}
			bf.cur, err = bf.file.Seek(0, io.SeekEnd)
			if err != nil {
				_ = f.Close()
			}
			return err
		}
	} else {
		if flag.NeedCreate() {
			bf.cur = BlkFileHeadSize
			if bf.compressor == nil {
				bf.compressor = compressor.NewBufferCompressor(BlkFileBufferSize)
			}
			bf.header.DataFormat = bf.compressor.Type().Value()
			err = bf.writeHeader(0)
			if err != nil {
				_ = f.Close()
			}
			return err
		}
	}
	_ = f.Close()
	return fmt.Errorf("Cannot open with flag %d. ", flag)
}

func (bf *BlkFileV2) Close() error {
	if bf.file != nil {
		return bf.file.Close()
	}
	return nil
}

// 12 Bytes
func (bf *BlkFileV2) writeHeader(size uint64) error {
	return WriteFileHeader(bf.header, bf.file)
}

func (bf *BlkFileV2) selectCompressor() error {
	if bf.compressor == nil {
		bf.compressor = compressor.NewBufferCompressor(BlkFileBufferSize)
	}
	cur := bf.compressor.Type().Value()
	switch bf.header.DataFormat {
	case compressor.TypeNone:
		if cur != bf.header.DataFormat {
			bf.compressor = compressor.NewBufferCompressor(BlkFileBufferSize)
		}
	case compressor.TypeGzip:
		if cur != bf.header.DataFormat {
			bf.compressor = compressor.NewGzipCompressor()
		}
	case compressor.TypeZlib:
		if cur != bf.header.DataFormat {
			bf.compressor = compressor.NewZlibCompressor()
		}
	default:
		return fmt.Errorf("Cannot support format type: %d. ", bf.header.DataFormat)
	}
	return nil
}

func (bf *BlkFileV2) readHeader() error {
	h, err := ReadFileHeader(bf.file)
	if err != nil {
		return err
	}
	if h.Version != BlkFileV2Version {
		return fmt.Errorf("Version: %d not support. ", h.Version)
	}
	bf.header = h
	return bf.selectCompressor()
}

func (bf *BlkFileV2) WriteFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return bf.WriteBlock(NewBlkHeader(path, info.Size()), f)
}

func (bf *BlkFileV2) ReadFile(path string) (*BlkHeader, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return bf.ReadBlock(f)
}

func (bf *BlkFileV2) WriteBlock(header *BlkHeader, reader io.Reader) error {
	length := len(header.Key)
	vi := VarInt{}
	vi.InitFromUInt64(uint64(length))
	wn, err := bf.file.Write(vi.Bytes())
	bf.cur += int64(wn)
	if err != nil {
		return err
	}
	wn, err = bf.file.Write([]byte(header.Key))
	bf.cur += int64(wn)
	if err != nil {
		return err
	}
	cur := bf.cur
	// write data first,get data length
	_, err = bf.file.Seek(8, io.SeekCurrent)
	bf.cur += 8
	if err != nil {
		return err
	}
	// write data
	_, n, err := bf.compressor.Compress(bf.file, reader)
	bf.cur += n
	if err != nil {
		return err
	}
	// seek to size record position
	_, err = bf.file.Seek(cur, io.SeekStart)
	if err != nil {
		return err
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	_, err = bf.file.Write(buf)
	if err != nil {
		return err
	}
	// seek to end
	_, err = bf.file.Seek(bf.cur, io.SeekStart)
	return err
}

func (bf *BlkFileV2) seek(offset int64) error {
	cur, err := bf.file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	bf.cur = cur
	return nil
}

func (bf *BlkFileV2) ReadBlock(w io.Writer) (*BlkHeader, error) {
	n, err := bf.readBlock(w)
	if err != nil {
		return nil, err
	} else {
		return &BlkHeader{
			Key:  n.key,
			Size: n.originSize,
		}, nil
	}
}

func (bf *BlkFileV2) readBlock(w io.Writer) (*blkNode, error) {
	node := &blkNode{}

	vi := VarInt{}
	b, rn, err := vi.LoadFromReader(bf.file)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New("Cannot parse varint. ")
	}
	size := vi.ToUint()
	buf := make([]byte, size)
	rn, err = bf.file.Read(buf)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	if rn != int(size) {
		return nil, errors.New("Read key length is not match record size! ")
	}
	node.key = string(buf)
	buf = make([]byte, 8)
	rn, err = bf.file.Read(buf)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	node.size = int64(binary.BigEndian.Uint64(buf))
	node.offset = bf.cur
	var n int64
	if w != nil {
		r := io.LimitReader(bf.file, node.size)
		n, node.originSize, err = bf.compressor.Decompress(w, r)
	} else {
		n, err = bf.file.Seek(node.size, io.SeekCurrent)
		n = n - bf.cur
	}
	bf.cur += n
	if err != nil {
		return nil, err
	}
	if n != node.size {
		return nil, errors.New("Read size is not match the Header Size! ")
	}

	return node, nil
}

func (bf *BlkFileV2) Flush() error {
	return bf.file.Sync()
}

type blkFileV2Openers struct{}

var BlkFileV2Openers blkFileV2Openers

func (o blkFileV2Openers) Local(path string) Opener {
	return func(flag flags.OpenFlag) (BlockReadWriter, bool, error) {
		if flag.CanWrite() && flag.CanRead() {
			return nil, false, errors.New("Tar format flag cannot contains both OpFlagReadOnly and OpFlagWriteOnly. ")
		}
		_, err := os.Stat(path)
		if err == nil {
			if flag.CanRead() {
				f, err := os.Open(path)
				return f, false, err
			} else if flag.CanWrite() {
				f, err := os.OpenFile(path, os.O_RDWR, 0666)
				return f, false, err
			}
		} else {
			if flag.NeedCreate() {
				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
				return f, true, err
			}
		}
		return nil, false, fmt.Errorf("Cannot open file %s with flag %d. ", path, flag)
	}
}
