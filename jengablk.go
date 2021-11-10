// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jenga

import (
	"github.com/xfali/jenga/blk"
	"io"
)

type blkJenga struct {
	flag OpenFlag
	blk  *jengablk.BlkMFile
}

func NewJenga(path string) *blkJenga {
	return &blkJenga{
		blk: jengablk.NewBlkMFile(path),
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
