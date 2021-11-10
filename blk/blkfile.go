// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xfali/jenga/flags"
	"io"
	"os"
)

const (
	BlkFileMagicCode       uint16 = 0xB1EF
	BlkFileVersion         uint16 = 0x0001
	BlkFileHeadSize               = 4
	BlkFileBufferSize             = 32 * 1024
	BlkHeaderUnknownOffset        = -1
)

type BlkFile struct {
	file    *os.File
	path    string
	magic   uint16
	version uint16
	buf     []byte
	cur     int64
}

type BlkHeader struct {
	// block key(name)
	Key string

	// block size
	Size int64

	// block offset
	offset int64
}

func NewBlkHeader(key string, size int64) *BlkHeader {
	return &BlkHeader{
		Key:    key,
		Size:   size,
		offset: BlkHeaderUnknownOffset,
	}
}

func NewBlkFile(path string) *BlkFile {
	return &BlkFile{
		path:    path,
		magic:   BlkFileMagicCode,
		version: BlkFileVersion,
		buf:     make([]byte, BlkFileBufferSize),
		cur:     0,
	}
}

func (bf *BlkFile) Open(flag flags.OpenFlag) error {
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
			f, err := os.OpenFile(bf.path, os.O_RDWR|os.O_APPEND, 0666)
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
			f, err := os.OpenFile(bf.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return err
			}
			bf.file = f
			bf.cur = BlkFileHeadSize
			return bf.writeHeader(0)
		}
	}
	return fmt.Errorf("Cannot open file %s with flag %d. ", bf.path, flag)
}

func (bf *BlkFile) Close() error {
	if bf.file != nil {
		return bf.file.Close()
	}
	return nil
}

// 12 Bytes
func (bf *BlkFile) writeHeader(size uint64) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, bf.magic)
	_, err := bf.file.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, bf.version)
	_, err = bf.file.Write(buf)
	return err
}

func (bf *BlkFile) readHeader() error {
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
	if bf.version != BlkFileVersion {
		return fmt.Errorf("Version: %d not support. ", bf.version)
	}

	return nil
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

func (h *BlkHeader) String() string {
	return fmt.Sprintf("key: %s , size: %d", h.Key, h.Size)
}

func (h *BlkHeader) Invalid() bool {
	return h.Size == 0
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
		return nil, errors.New("Read size is not match then Header Size! ")
	}

	return header, nil
}

func (bf *BlkFile) Flush() error {
	return bf.file.Sync()
}
