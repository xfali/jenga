// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jenga

import (
	"fmt"
	"github.com/xfali/jenga/blk"
	"io"
)

type blkJenga struct {
	flag OpenFlag
	blk  *jengablk.BlkMFile
}

func NewJenga(path string, opts ...jengablk.MFileOpt) *blkJenga {
	return &blkJenga{
		blk: jengablk.NewBlkMFile(path, opts...),
	}
}

func (jenga *blkJenga) Open(flag OpenFlag) error {
	jenga.flag = flag
	return jenga.blk.Open(flag)
}

func (jenga *blkJenga) KeyList() []string {
	return jenga.blk.Keys()
}

func (jenga *blkJenga) Write(path string, size int64, r io.Reader) error {
	if !jenga.flag.CanWrite() {
		return WriteFlagError
	}
	if size <= 0 && jenga.blk.NeedSize() {
		return fmt.Errorf("blkJenga param size %d is Illegal, it must be actual reader data size. ", size)
	}
	return jenga.blk.WriteBlock(jengablk.NewBlkHeader(path, size), r)
}

func (jenga *blkJenga) Read(path string, w io.Writer) (int64, error) {
	if !jenga.flag.CanRead() {
		return 0, ReadFlagError
	}
	return jenga.blk.ReadBlockByKey(path, w)
}

func (jenga *blkJenga) Close() (err error) {
	return jenga.blk.Close()
}
