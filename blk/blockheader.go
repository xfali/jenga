// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BlkFileMagicCode       uint16 = 0xB1EF
	BlkFileHeadSize               = 8
	BlkFileBufferSize             = 32 * 1024
	BlkHeaderUnknownOffset        = -1
)

type BlkHeader struct {
	// block key(name)
	Key string

	// block size
	Size int64
}

type blkNode struct {
	// node key(name)
	key string

	// node size
	size int64

	// blk origin size without compress
	originSize int64

	// node offset
	offset int64
}

func (h *blkNode) invalid() bool {
	return h.size == 0
}

func NewBlkHeader(key string, size int64) *BlkHeader {
	return &BlkHeader{
		Key:  key,
		Size: size,
	}
}

func (h *BlkHeader) String() string {
	return fmt.Sprintf("key: %s , size: %d", h.Key, h.Size)
}

func (h *BlkHeader) Invalid() bool {
	return h.Size == 0
}

type FileHeader struct {
	MagicCode  uint16
	Version    uint16
	DataFormat uint16
	Reverse    uint16
}

func ReadFileHeader(r io.Reader) (FileHeader, error) {
	h := FileHeader{}
	buf := make([]byte, 2)
	_, err := r.Read(buf)
	if err != nil {
		return h, err
	}
	h.MagicCode = binary.BigEndian.Uint16(buf)
	if h.MagicCode != BlkFileMagicCode {
		return h, errors.New("Jenga file format not match, maybe broken. ")
	}
	_, err = r.Read(buf)
	if err != nil {
		return h, err
	}
	h.Version = binary.BigEndian.Uint16(buf)

	_, err = r.Read(buf)
	if err != nil {
		return h, err
	}
	h.DataFormat = binary.BigEndian.Uint16(buf)

	_, err = r.Read(buf)
	if err != nil {
		return h, err
	}
	h.Reverse = binary.BigEndian.Uint16(buf)
	return h, nil
}

func WriteFileHeader(h FileHeader, w io.Writer) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, BlkFileMagicCode)
	_, err := w.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, h.Version)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, h.DataFormat)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, h.Reverse)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return err
}
