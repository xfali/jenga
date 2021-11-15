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
	blk  jengablk.JengaBlocks
}

type Opt func(j *blkJenga, uri string)

func NewJenga(uri string, opts ...Opt) *blkJenga {
	ret := &blkJenga{
		blk: jengablk.NewV1BlockFile(uri),
	}
	for _, opt := range opts {
		opt(ret, uri)
	}
	return ret
}

func NewJengaWithOpts(opts ...Opt) *blkJenga {
	ret := &blkJenga{}
	for _, opt := range opts {
		opt(ret, "")
	}
	if ret.blk == nil {
		panic("Cannot Open! ")
	}
	return ret
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
	if jenga.blk.NeedSize() && (size <= 0 && r != nil) {
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

func V1(opts ...jengablk.BlocksV1Opt) Opt {
	return func(j *blkJenga, uri string) {
		if uri != "" {
			j.blk = jengablk.NewV1BlockFile(uri, opts...)
		} else {
			j.blk = jengablk.NewV1Blocks(opts...)
		}
	}
}

func V2(opts ...jengablk.BlocksV2Opt) Opt {
	return func(j *blkJenga, uri string) {
		if uri != "" {
			j.blk = jengablk.NewV2BlockFile(uri, opts...)
		} else {
			j.blk = jengablk.NewV2Blocks(opts...)
		}
	}
}

func V2Gzip(opts ...jengablk.BlocksV2Opt) Opt {
	return func(j *blkJenga, uri string) {
		var newOpt []jengablk.BlocksV2Opt
		if uri != "" {
			newOpt = append(newOpt, jengablk.BlockV2Opts.LocalFile(uri))
		}
		newOpt = append(newOpt, jengablk.BlockV2Opts.WithGzip())
		j.blk = jengablk.NewV2Blocks(append(newOpt, opts...)...)
	}
}

func V2Zlib(opts ...jengablk.BlocksV2Opt) Opt {
	return func(j *blkJenga, uri string) {
		var newOpt []jengablk.BlocksV2Opt
		if uri != "" {
			newOpt = append(newOpt, jengablk.BlockV2Opts.LocalFile(uri))
		}
		newOpt = append(newOpt, jengablk.BlockV2Opts.WithZlib())
		j.blk = jengablk.NewV2Blocks(append(newOpt, opts...)...)
	}
}

func WithBlocks(blk jengablk.JengaBlocks) Opt {
	return func(j *blkJenga, uri string) {
		j.blk = blk
	}
}

func WithFactory(factory func(uri string) jengablk.JengaBlocks) Opt {
	return func(j *blkJenga, uri string) {
		j.blk = factory(uri)
	}
}
