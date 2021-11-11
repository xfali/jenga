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
	"sync"
)

type GetSizeFunc func(string) int64

type BlkMFile struct {
	f        *BlkFile
	sizeFunc GetSizeFunc
	meta     sync.Map
}

type MFileOpt func(f *BlkMFile)

func NewBlkMFile(path string, opts ...MFileOpt) *BlkMFile {
	ret := &BlkMFile{
		f: NewBlkFile(path),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (bf *BlkMFile) Open(flag flags.OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return errors.New("Tar format flag cannot contains both OpFlagReadOnly and OpFlagWriteOnly. ")
	}
	err := bf.f.Open(flag)
	if err != nil {
		return err
	}
	return bf.loadMeta(flag)
}

func (bf *BlkMFile) loadMeta(flag flags.OpenFlag) error {
	if bf.f.cur != BlkFileHeadSize {
		err := bf.f.seek(BlkFileHeadSize)
		if err != nil {
			return err
		}
	}
	for {
		h, err := bf.f.ReadBlock(nil)
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
		bf.meta.Store(h.Key, h)
	}
}

func (bf *BlkMFile) Close() error {
	return bf.f.Close()
}

func (bf *BlkMFile) WriteFile(path string) error {
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

func (bf *BlkMFile) ReadFile(path string) (*BlkHeader, error) {
	return bf.f.ReadFile(path)
}

func (bf *BlkMFile) WriteBlock(header *BlkHeader, reader io.Reader) error {
	if _, ok := bf.meta.LoadOrStore(header.Key, header); ok {
		return fmt.Errorf("Block with key %s have been written. ", header.Key)
	}
	if header.Size == 0 && reader != nil {
		if bf.sizeFunc != nil {
			size := bf.sizeFunc(header.Key)
			if size > 0 {
				header.Size = size
				return bf.f.WriteBlock(header, reader)
			}
		}
		return fmt.Errorf("Block with key %s size is 0. ", header.Key)
	}
	return bf.f.WriteBlock(header, reader)
}

func (bf *BlkMFile) NeedSize() bool {
	return bf.sizeFunc == nil
}

func (bf *BlkMFile) ReadBlock(w io.Writer) (*BlkHeader, error) {
	return bf.f.ReadBlock(w)
}

func (bf *BlkMFile) Keys() []string {
	var ret []string
	bf.meta.Range(func(key, value interface{}) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (bf *BlkMFile) ReadBlockByKey(key string, w io.Writer) (int64, error) {
	if v, ok := bf.meta.Load(key); ok {
		header := v.(*BlkHeader)
		if header.Invalid() {
			return 0, fmt.Errorf("Block with key: %s not found. ", key)
		}
		err := bf.f.seek(header.offset)
		if err != nil {
			return 0, err
		}
		var n int64
		if w != nil {
			r := io.LimitReader(bf.f.file, header.Size)
			n, err = io.CopyBuffer(w, r, bf.f.buf)
		} else {
			n, err = bf.f.file.Seek(header.Size, io.SeekCurrent)
			n = n - bf.f.cur
		}
		bf.f.cur += n
		if err != nil {
			return n, err
		}
		if n != header.Size {
			return n, errors.New("Read size is not match then Header Size! ")
		}
		return n, nil
	} else {
		return 0, fmt.Errorf("Block with key: %s not found. ", key)
	}
}

func (bf *BlkMFile) Flush() error {
	return bf.f.Flush()
}

type mfileOpts struct{}

var MFileOpts mfileOpts

func (opts mfileOpts) WithSizeFun(sizeFunc GetSizeFunc) MFileOpt {
	return func(f *BlkMFile) {
		f.sizeFunc = sizeFunc
	}
}

func (opts mfileOpts) FileKey() MFileOpt {
	return func(f *BlkMFile) {
		f.sizeFunc = func(s string) int64 {
			info, err := os.Stat(s)
			if err != nil {
				return 0
			}
			return info.Size()
		}
	}
}
