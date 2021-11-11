// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"github.com/xfali/jenga/flags"
	"io"
)

type JengaBlocks interface {
	Open(flag flags.OpenFlag) error
	Keys() []string
	WriteBlock(header *BlkHeader, reader io.Reader) error
	ReadBlock(w io.Writer) (*BlkHeader, error)
	ReadBlockByKey(path string, writer io.Writer) (int64, error)
	Close() (err error)
	NeedSize() bool
}
