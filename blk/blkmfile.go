// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengablk

import (
	"errors"
	"github.com/xfali/jenga/flags"
	"github.com/xfali/jenga/jengaerr"
	"io"
	"os"
	"sync"
)

type GetSizeFunc func(string) int64

type blockV1 struct {
	f        *BlkFile
	sizeFunc GetSizeFunc
	meta     sync.Map
}

type BlocksV1Opt func(f *blockV1)

func NewV1BlockFile(path string, opts ...BlocksV1Opt) *blockV1 {
	newOpt := []BlocksV1Opt{BlockV1Opts.LocalFile(path)}
	return NewV1Blocks(append(newOpt, opts...)...)
}

func NewV1Blocks(opts ...BlocksV1Opt) *blockV1 {
	ret := &blockV1{
		sizeFunc: GetFileSize,
	}
	for _, opt := range opts {
		opt(ret)
	}
	if ret.f == nil {
		panic("Blocks cannot open!")
	}
	return ret
}

func (bf *blockV1) Open(flag flags.OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return jengaerr.OpenRWFlagError.Format("BlockV1")
	}
	err := bf.f.Open(flag)
	if err != nil {
		return err
	}
	return bf.loadMeta(flag)
}

func (bf *blockV1) loadMeta(flag flags.OpenFlag) error {
	if bf.f.cur != BlkFileHeadSize {
		err := bf.f.seek(BlkFileHeadSize)
		if err != nil {
			return err
		}
	}
	for {
		h, err := bf.f.readBlock(nil)
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
		bf.meta.Store(h.key, h)
	}
}

func (bf *blockV1) Close() error {
	return bf.f.Close()
}

func (bf *blockV1) WriteFile(path string) error {
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

func (bf *blockV1) ReadFile(path string) (*BlkHeader, error) {
	return bf.f.ReadFile(path)
}

func (bf *blockV1) WriteBlock(header *BlkHeader, reader io.Reader) error {
	if _, ok := bf.meta.LoadOrStore(header.Key, &blkNode{
		key:  header.Key,
		size: header.Size,
	}); ok {
		return jengaerr.WriteExistKeyError.Format(header.Key)
	}
	if header.Size <= 0 && reader != nil {
		if bf.sizeFunc != nil {
			size := bf.sizeFunc(header.Key)
			if size > 0 {
				header.Size = size
				return bf.f.WriteBlock(header, reader)
			}
		}
		return jengaerr.WriteWithoutSizeFuncError.Format("BlockV1")
	}
	return bf.f.WriteBlock(header, reader)
}

func (bf *blockV1) NeedSize() bool {
	return bf.sizeFunc == nil
}

func (bf *blockV1) ReadBlock(w io.Writer) (*BlkHeader, error) {
	return bf.f.ReadBlock(w)
}

func (bf *blockV1) Keys() []string {
	var ret []string
	bf.meta.Range(func(key, value interface{}) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (bf *blockV1) ReadBlockByKey(key string, w io.Writer) (int64, error) {
	if v, ok := bf.meta.Load(key); ok {
		node := v.(*blkNode)
		if node.invalid() {
			return 0, jengaerr.ReadKeyNotFoundError.Format(key)
		}
		err := bf.f.seek(node.offset)
		if err != nil {
			return 0, err
		}
		var n int64
		if w != nil {
			r := io.LimitReader(bf.f.file, node.size)
			n, err = io.CopyBuffer(w, r, bf.f.buf)
		} else {
			n, err = bf.f.file.Seek(node.size, io.SeekCurrent)
			n = n - bf.f.cur
		}
		bf.f.cur += n
		if err != nil {
			return n, err
		}
		if n != node.size {
			return n, jengaerr.ReadNodeSizeNotMatchError
		}
		return n, nil
	} else {
		return 0, jengaerr.ReadKeyNotFoundError.Format(key)
	}
}

func (bf *blockV1) Flush() error {
	return bf.f.Flush()
}

type blockV1Opts struct{}

var BlockV1Opts blockV1Opts

type KeySize struct {
	Key  string
	Size int64
}

func (opts blockV1Opts) WithKeySizes(ks ...KeySize) BlocksV1Opt {
	m := map[string]int64{}
	for _, v := range ks {
		m[v.Key] = v.Size
	}
	return func(f *blockV1) {
		f.sizeFunc = func(s string) int64 {
			return m[s]
		}
	}
}

func (opts blockV1Opts) WithSizeFunc(sizeFunc GetSizeFunc) BlocksV1Opt {
	return func(f *blockV1) {
		f.sizeFunc = sizeFunc
	}
}

func (opts blockV1Opts) FileKey() BlocksV1Opt {
	return func(f *blockV1) {
		f.sizeFunc = GetFileSize
	}
}

func GetFileSize(s string) int64 {
	info, err := os.Stat(s)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (opts blockV1Opts) WithBlkFile(bf *BlkFile) BlocksV1Opt {
	return func(f *blockV1) {
		f.f = bf
	}
}

func (opts blockV1Opts) LocalFile(path string) BlocksV1Opt {
	return func(f *blockV1) {
		f.f = NewBlkFile(path)
	}
}

func (opts blockV1Opts) WithOpener(openers Opener) BlocksV1Opt {
	return func(f *blockV1) {
		f.f = NewBlkFileWithOpener(openers)
	}
}
