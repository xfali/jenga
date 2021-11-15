// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"errors"
	"fmt"
	"github.com/xfali/jenga/flags"
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
	return fmt.Errorf("Cannot open with flag %d. ", flag)
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
		return fmt.Errorf("Version: %d not support. ", bf.version)
	}
	bf.version = h.Version
	return err
}

func (bf *BlkFile) WriteFile(path string) error {
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

func (bf *BlkFile) ReadFile(path string) (*BlkHeader, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return bf.ReadBlock(f)
}

func (bf *BlkFile) WriteBlock(header *BlkHeader, reader io.Reader) error {
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
	vi.InitFromUInt64(uint64(header.Size))
	_, err = bf.file.Write(vi.Bytes())
	if err != nil {
		return err
	}
	n, err := io.CopyBuffer(bf.file, reader, bf.buf)
	bf.cur += int64(n)
	if err != nil {
		return err
	}
	if n != header.Size {
		return errors.New("Write size is not match then Header Size! ")
	}
	return nil
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
	vi = VarInt{}
	b, rn, err = vi.LoadFromReader(bf.file)
	bf.cur += int64(rn)
	if err != nil {
		return nil, err
	}
	header.Size = vi.ToInt()
	header.offset = bf.cur
	var n int64
	if w != nil {
		r := io.LimitReader(bf.file, header.Size)
		n, err = io.CopyBuffer(w, r, bf.buf)
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

func (bf *BlkFile) Flush() error {
	return bf.file.Sync()
}

type blkFileOpeners struct{}

var BlkFileOpeners blkFileOpeners

func (o blkFileOpeners) Local(path string) Opener {
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
				f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0666)
				return f, false, err
			}
		} else {
			if flag.NeedCreate() {
				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				return f, true, err
			}
		}
		return nil, false, fmt.Errorf("Cannot open file %s with flag %d. ", path, flag)
	}
}
