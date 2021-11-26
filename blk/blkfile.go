// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"github.com/xfali/jenga/flags"
	"github.com/xfali/jenga/jengaerr"
	"io"
	"os"
)

const (
	BlkFileVersion uint16 = 0x0001
)

// File format:
// |MAGIC NUNMBER(2 Bytes)|VERSION(2 Bytes)|DATA FORMAT(2 Bytes)|REVERSE(2 Bytes)|ENTITY_1|ENTITY_2|...|ENTITY_N|
// Entity format:
// |VARINT(1-10 Bytes)|STRING(string length)|VARINT(1-10 Bytes)|DATA(data size)|
type BlkFile struct {
	file    BlockReadWriter
	opener  Opener
	version uint16
	buf     []byte
	cur     int64
}

func NewBlkFile(path string) *BlkFile {
	return NewBlkFileWithOpener(BlkFileOpeners.Local(path))
}

func NewBlkFileWithOpener(opener Opener) *BlkFile {
	return &BlkFile{
		opener:  opener,
		version: BlkFileVersion,
		buf:     make([]byte, BlkFileBufferSize),
		cur:     0,
	}
}

func (bf *BlkFile) Open(flag flags.OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return jengaerr.OpenRWFlagError.Format("BlockV1")
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
			}
			return err
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

func (bf *BlkFile) Close() error {
	if bf.file != nil {
		return bf.file.Close()
	}
	return nil
}

// 12 Bytes
func (bf *BlkFile) writeHeader(size uint64) error {
	return WriteFileHeader(FileHeader{
		Version: bf.version,
	}, bf.file)
}

func (bf *BlkFile) readHeader() error {
	h, err := ReadFileHeader(bf.file)
	if err != nil {
		return err
	}
	if h.Version != BlkFileVersion {
		return jengaerr.VersionNotSupportError.Format(h.Version, "BlkFile", BlkFileVersion)
	}
	bf.version = h.Version
	return err
}

func (bf *BlkFile) WriteFile(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return bf.WriteBlock(NewBlkHeader(path, info.Size()), f)
}

func (bf *BlkFile) ReadFile(path string) (*BlkHeader, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return bf.ReadBlock(f)
}

func (bf *BlkFile) WriteBlock(header *BlkHeader, reader io.Reader) (int64, error) {
	length := len(header.Key)
	vi := VarInt{}
	vi.InitFromUInt64(uint64(length))
	wn, err := bf.file.Write(vi.Bytes())
	bf.cur += int64(wn)
	if err != nil {
		return 0, err
	}
	wn, err = bf.file.Write([]byte(header.Key))
	bf.cur += int64(wn)
	if err != nil {
		return 0, err
	}
	vi.InitFromUInt64(uint64(header.Size))
	_, err = bf.file.Write(vi.Bytes())
	if err != nil {
		return 0, err
	}
	n, err := io.CopyBuffer(bf.file, reader, bf.buf)
	bf.cur += int64(n)
	if err != nil {
		return n, err
	}
	if n != header.Size {
		return n, jengaerr.WriteSizeNotMatchError
	}
	return n, nil
}

func (bf *BlkFile) seek(offset int64) error {
	cur, err := bf.file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	bf.cur = cur
	return nil
}

func (bf *BlkFile) ReadBlock(w io.Writer) (*BlkHeader, error) {
	n, err := bf.readBlock(w)
	if err == nil {
		return &BlkHeader{
			Key:  n.key,
			Size: n.size,
		}, nil
	}
	return nil, err
}

func (bf *BlkFile) readBlock(w io.Writer) (*blkNode, error) {
	header := &blkNode{}

	vi := VarInt{}
	b, rn, err := vi.LoadFromReader(bf.file)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, jengaerr.ReadBlockVarintFailedError
	}
	size := vi.ToUint()
	buf := make([]byte, size)
	rn, err = bf.file.Read(buf)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	if rn != int(size) {
		return nil, jengaerr.ReadKeySizeNotMatchError
	}
	header.key = string(buf)
	vi = VarInt{}
	b, rn, err = vi.LoadFromReader(bf.file)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	header.size = vi.ToInt()
	header.offset = bf.cur
	var n int64
	if w != nil {
		r := io.LimitReader(bf.file, header.size)
		n, err = io.CopyBuffer(w, r, bf.buf)
	} else {
		n, err = bf.file.Seek(header.size, io.SeekCurrent)
		n = n - bf.cur
	}
	bf.cur += n
	if err != nil {
		return nil, err
	}
	if n != header.size {
		return nil, jengaerr.ReadNodeSizeNotMatchError
	}

	return header, nil
}

func (bf *BlkFile) Flush() error {
	return bf.file.Sync()
}

type blkFileOpeners struct{}

var BlkFileOpeners blkFileOpeners

func (o blkFileOpeners) Local(path string) Opener {
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
				f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0666)
				return f, false, err
			}
		} else {
			if flag.NeedCreate() {
				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				return f, true, err
			}
		}
		return nil, false, jengaerr.OpenFileError.Format(path, flag)
	}
}
