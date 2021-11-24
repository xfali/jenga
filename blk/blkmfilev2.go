// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"errors"
	"fmt"
	"github.com/xfali/jenga/compressor"
	"github.com/xfali/jenga/flags"
	"io"
	"os"
	"sync"
)

type blockV2 struct {
	f    *BlkFileV2
	meta sync.Map
}

type BlocksV2Opt func(f *blockV2)

func NewV2BlockFile(path string, opts ...BlocksV2Opt) *blockV2 {
	newOpt := []BlocksV2Opt{BlockV2Opts.LocalFile(path)}
	return NewV2Blocks(append(newOpt, opts...)...)
}

func NewV2Blocks(opts ...BlocksV2Opt) *blockV2 {
	ret := &blockV2{}
	for _, opt := range opts {
		opt(ret)
	}
	if ret.f == nil {
		panic("Blocks cannot open!")
	}
	return ret
}

func (bf *blockV2) Open(flag flags.OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return errors.New("Tar format flag cannot contains both OpFlagReadOnly and OpFlagWriteOnly. ")
	}
	err := bf.f.Open(flag)
	if err != nil {
		return err
	}
	return bf.loadMeta(flag)
}

func (bf *blockV2) loadMeta(flag flags.OpenFlag) error {
	if bf.f.cur != BlkFileHeadSize {
		err := bf.f.seek(BlkFileHeadSize)
		if err != nil {
			return err
		}
	}
	for {
		n, err := bf.f.readBlock(nil)
		if err != nil {
			// 最后一个
			if errors.Is(err, io.EOF) {
				if flag.CanRead() {
					err := bf.f.seek(BlkFileHeadSize)
					if err != nil {
						return err
					}
				}
				return nil
			} else {
				return err
			}
		}
		bf.meta.Store(n.key, n)
	}
}

func (bf *blockV2) Close() error {
	return bf.f.Close()
}

func (bf *blockV2) WriteFile(path string) error {
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

func (bf *blockV2) ReadFile(path string) (*BlkHeader, error) {
	return bf.f.ReadFile(path)
}

func (bf *blockV2) WriteBlock(header *BlkHeader, reader io.Reader) error {
	if _, ok := bf.meta.LoadOrStore(header.Key, &blkNode{
		key:  header.Key,
		size: header.Size,
	}); ok {
		return fmt.Errorf("Block with key %s have been written. ", header.Key)
	}
	return bf.f.WriteBlock(header, reader)
}

func (bf *blockV2) NeedSize() bool {
	return false
}

func (bf *blockV2) ReadBlock(w io.Writer) (*BlkHeader, error) {
	return bf.f.ReadBlock(w)
}

func (bf *blockV2) Keys() []string {
	var ret []string
	bf.meta.Range(func(key, value interface{}) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (bf *blockV2) ReadBlockByKey(key string, w io.Writer) (int64, error) {
	if v, ok := bf.meta.Load(key); ok {
		node := v.(*blkNode)
		if node.invalid() {
			return 0, fmt.Errorf("Block with key: %s not found. ", key)
		}
		err := bf.f.seek(node.offset)
		if err != nil {
			return 0, err
		}
		var n int64
		if w != nil {
			r := io.LimitReader(bf.f.file, node.size)
			n, node.originSize, err = bf.f.compressor.Decompress(w, r)
		} else {
			n, err = bf.f.file.Seek(node.size, io.SeekCurrent)
			n = n - bf.f.cur
		}
		bf.f.cur += n
		if err != nil {
			return node.originSize, err
		}
		if n != node.size {
			return node.originSize, errors.New("Read size is not match then Header Size! ")
		}
		return node.originSize, nil
	} else {
		return 0, fmt.Errorf("Block with key: %s not found. ", key)
	}
}

func (bf *blockV2) Flush() error {
	return bf.f.Flush()
}

type blockV2Opts struct{}

var BlockV2Opts blockV2Opts

func (opts blockV2Opts) WithCompressor(compressor compressor.Compressor) BlocksV2Opt {
	return func(f *blockV2) {
		f.f.WithCompressor(compressor)
	}
}

func (opts blockV2Opts) WithGzip() BlocksV2Opt {
	return func(f *blockV2) {
		f.f.WithCompressor(compressor.NewGzipCompressor())
	}
}

func (opts blockV2Opts) WithZlib() BlocksV2Opt {
	return func(f *blockV2) {
		f.f.WithCompressor(compressor.NewZlibCompressor())
	}
}

func (opts blockV2Opts) WithBlkFile(bf *BlkFileV2) BlocksV2Opt {
	return func(f *blockV2) {
		f.f = bf
	}
}

func (opts blockV2Opts) LocalFile(path string) BlocksV2Opt {
	return func(f *blockV2) {
		f.f = NewBlkFileV2(path)
	}
}

func (opts blockV2Opts) WithOpener(openers Opener) BlocksV2Opt {
	return func(f *blockV2) {
		f.f = NewBlkFileV2WithOpener(openers)
	}
}
