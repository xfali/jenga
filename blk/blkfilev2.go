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
	file       *os.File
	path       string
	magic      uint16
	version    uint16
	cur        int64
	dataFormat uint16
	reverse    uint16
	compressor compressor.Compressor
}

func NewBlkFileV2(path string) *BlkFileV2 {
	return &BlkFileV2{
		path:       path,
		magic:      BlkFileMagicCode,
		version:    BlkFileV2Version,
		cur:        0,
		compressor: nil,
		dataFormat: compressor.TypeNone,
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
	info, err := os.Stat(bf.path)
	if err == nil {
		if flag.CanRead() {
			f, err := os.Open(bf.path)
			if err != nil {
				return err
			}
			bf.file = f
			bf.cur = BlkFileHeadSize
			return bf.readHeader()
		} else if flag.CanWrite() {
			f, err := os.OpenFile(bf.path, os.O_RDWR, 0666)
			if err != nil {
				return err
			}
			bf.file = f
			err = bf.readHeader()
			if err != nil {
				return err
			}
			bf.cur = info.Size()
			_, err = bf.file.Seek(0, io.SeekEnd)
			return err
		}
	} else {
		if flag.NeedCreate() {
			f, err := os.OpenFile(bf.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return err
			}
			bf.file = f
			bf.cur = BlkFileHeadSize
			if bf.compressor == nil {
				bf.compressor = compressor.NewBufferCompressor(BlkFileBufferSize)
			}
			bf.dataFormat = bf.compressor.Type().Value()
			return bf.writeHeader(0)
		}
	}
	return fmt.Errorf("Cannot open file %s with flag %d. ", bf.path, flag)
}

func (bf *BlkFileV2) Close() error {
	if bf.file != nil {
		return bf.file.Close()
	}
	return nil
}

// 12 Bytes
func (bf *BlkFileV2) writeHeader(size uint64) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, bf.magic)
	_, err := bf.file.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, bf.version)
	_, err = bf.file.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, bf.dataFormat)
	_, err = bf.file.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, bf.reverse)
	_, err = bf.file.Write(buf)
	if err != nil {
		return err
	}
	return err
}

func (bf *BlkFileV2) selectCompressor() error {
	if bf.compressor == nil {
		bf.compressor = compressor.NewBufferCompressor(BlkFileBufferSize)
	}
	cur := bf.compressor.Type().Value()
	switch bf.dataFormat {
	case compressor.TypeNone:
		if cur != bf.dataFormat {
			bf.compressor = compressor.NewBufferCompressor(BlkFileBufferSize)
		}
	case compressor.TypeGzip:
		if cur != bf.dataFormat {
			bf.compressor = compressor.NewGzipCompressor()
		}
	case compressor.TypeZlib:
		if cur != bf.dataFormat {
			bf.compressor = compressor.NewZlibCompressor()
		}
	default:
		return fmt.Errorf("Cannot support format type: %d. ", bf.dataFormat)
	}
	return nil
}

func (bf *BlkFileV2) readHeader() error {
	buf := make([]byte, 2)
	_, err := bf.file.Read(buf)
	if err != nil {
		return err
	}
	bf.magic = binary.BigEndian.Uint16(buf)
	if bf.magic != BlkFileMagicCode {
		return errors.New("File format not match, maybe broken. ")
	}

	_, err = bf.file.Read(buf)
	if err != nil {
		return err
	}
	bf.version = binary.BigEndian.Uint16(buf)
	if bf.version != BlkFileV2Version {
		return fmt.Errorf("Version: %d not support. ", bf.version)
	}

	_, err = bf.file.Read(buf)
	if err != nil {
		return err
	}
	bf.dataFormat = binary.BigEndian.Uint16(buf)

	_, err = bf.file.Read(buf)
	if err != nil {
		return err
	}
	bf.reverse = binary.BigEndian.Uint16(buf)
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
	header := &BlkHeader{}

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
	header.Key = string(buf)
	buf = make([]byte, 8)
	rn, err = bf.file.Read(buf)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	header.Size = int64(binary.BigEndian.Uint64(buf))
	header.offset = bf.cur
	var n int64
	if w != nil {
		r := io.LimitReader(bf.file, header.Size)
		n, _, err = bf.compressor.Decompress(w, r)
	} else {
		n, err = bf.file.Seek(header.Size, io.SeekCurrent)
		n = n - bf.cur
	}
	bf.cur += n
	if err != nil {
		return nil, err
	}
	if n != header.Size {
		return nil, errors.New("Read size is not match the Header Size! ")
	}

	return header, nil
}

func (bf *BlkFileV2) Flush() error {
	return bf.file.Sync()
}
