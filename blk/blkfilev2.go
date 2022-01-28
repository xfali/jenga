// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"encoding/binary"
	"github.com/xfali/jenga/compressor"
	"github.com/xfali/jenga/flags"
	"github.com/xfali/jenga/jengaerr"
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
		return jengaerr.OpenRWFlagError.Format("BlockV2")
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
	return jengaerr.OpenFlagError.Format(flag)
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
		return jengaerr.DataFormatNotSupportError.Format(bf.header.DataFormat)
	}
	return nil
}

func (bf *BlkFileV2) readHeader() error {
	h, err := ReadFileHeader(bf.file)
	if err != nil {
		return err
	}
	if h.Version != bf.header.Version {
		return jengaerr.VersionNotSupportError.Format(h.Version, bf.header.Version)
	}
	bf.header = h
	return bf.selectCompressor()
}

func (bf *BlkFileV2) WriteFile(path string) (int64, error) {
	_, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return bf.WriteBlock(path, f)
}

func (bf *BlkFileV2) ReadFile(path string) (*BlkHeader, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return bf.ReadBlock(f)
}

func (bf *BlkFileV2) WriteBlock(key string, reader io.Reader) (int64, error) {
	length := len(key)
	vi := VarInt{}
	vi.InitFromUInt64(uint64(length))
	wn, err := bf.file.Write(vi.Bytes())
	bf.cur += int64(wn)
	if err != nil {
		return 0, err
	}
	wn, err = bf.file.Write([]byte(key))
	bf.cur += int64(wn)
	if err != nil {
		return 0, err
	}
	cur := bf.cur
	// write data first,get data length
	_, err = bf.file.Seek(8, io.SeekCurrent)
	bf.cur += 8
	if err != nil {
		return 0, err
	}
	// write data
	originWn, n, err := bf.compressor.Compress(bf.file, reader)
	bf.cur += n
	if err != nil {
		return originWn, err
	}
	// seek to size record position
	_, err = bf.file.Seek(cur, io.SeekStart)
	if err != nil {
		return originWn, err
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	_, err = bf.file.Write(buf)
	if err != nil {
		return originWn, err
	}
	// seek to end
	_, err = bf.file.Seek(bf.cur, io.SeekStart)
	return originWn, err
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

func (bf *BlkFileV2) readKey() (string, error) {
	vi := VarInt{}
	b, rn, err := vi.LoadFromReader(bf.file)
	bf.cur += int64(rn)
	if err != nil {
		return "", err
	}
	if !b {
		return "", jengaerr.ReadBlockVarintFailedError
	}
	size := vi.ToUint()
	buf := make([]byte, size)
	rn, err = bf.file.Read(buf)
	bf.cur += int64(rn)
	if err != nil {
		return "", err
	}
	if rn != int(size) {
		return "", jengaerr.ReadKeySizeNotMatchError
	}
	return string(buf), nil
}

func (bf *BlkFileV2) readPayloadSize() (int64, error) {
	buf := make([]byte, 8)
	rn, err := bf.file.Read(buf)
	bf.cur += int64(rn)
	if err != nil {
		return -1, err
	}
	return int64(binary.BigEndian.Uint64(buf)), nil
}

func (bf *BlkFileV2) readPayload(w io.Writer, size int64) (int64, error) {
	var n, originSize int64
	var err error
	if w != nil {
		r := io.LimitReader(bf.file, size)
		n, originSize, err = bf.compressor.Decompress(w, r)
	} else {
		n, err = bf.file.Seek(size, io.SeekCurrent)
		n = n - bf.cur
	}
	bf.cur += n
	if err != nil {
		return -1, err
	}
	if n != size {
		return -1, jengaerr.ReadNodeSizeNotMatchError
	}
	return originSize, err
}

func (bf *BlkFileV2) readBlock(w io.Writer) (*blkNode, error) {
	node := &blkNode{}

	key, err := bf.readKey()
	if err != nil {
		return nil, err
	}
	node.key = key

	size, err := bf.readPayloadSize()
	if err != nil {
		return nil, err
	}

	node.size = size
	node.offset = bf.current()
	node.originSize, err = bf.readPayload(w, size)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (bf *BlkFileV2) current() int64 {
	return bf.cur
}

func (bf *BlkFileV2) Flush() error {
	return bf.file.Sync()
}

type blkFileV2Openers struct{}

var BlkFileV2Openers blkFileV2Openers

func (o blkFileV2Openers) Local(path string) Opener {
	return func(flag flags.OpenFlag) (BlockReadWriter, bool, error) {
		if flag.CanWrite() && flag.CanRead() {
			return nil, false, jengaerr.OpenRWFlagError.Format("File")
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
		return nil, false, jengaerr.OpenFileError.Format(path, flag)
	}
}
