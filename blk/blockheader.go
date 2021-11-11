// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import "fmt"

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

func (h *BlkHeader) String() string {
	return fmt.Sprintf("key: %s , size: %d", h.Key, h.Size)
}

func (h *BlkHeader) Invalid() bool {
	return h.Size == 0
}
