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

type BlkMFileV2 struct {
	f    *BlkFileV2
	meta sync.Map
}

type MFileV2Opt func(f *BlkMFileV2)

func NewBlkMFileV2(path string, opts ...MFileV2Opt) *BlkMFileV2 {
	ret := &BlkMFileV2{
		f: NewBlkFileV2(path),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (bf *BlkMFileV2) Open(flag flags.OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return errors.New("Tar format flag cannot contains both OpFlagReadOnly and OpFlagWriteOnly. ")
	}
	err := bf.f.Open(flag)
	if err != nil {
		return err
	}
	return bf.loadMeta(flag)
}

func (bf *BlkMFileV2) loadMeta(flag flags.OpenFlag) error {
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

func (bf *BlkMFileV2) Close() error {
	return bf.f.Close()
}

func (bf *BlkMFileV2) WriteFile(path string) error {
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

func (bf *BlkMFileV2) ReadFile(path string) (*BlkHeader, error) {
	return bf.f.ReadFile(path)
}

func (bf *BlkMFileV2) WriteBlock(header *BlkHeader, reader io.Reader) error {
	if _, ok := bf.meta.LoadOrStore(header.Key, header); ok {
		return fmt.Errorf("Block with key %s have been written. ", header.Key)
	}
	return bf.f.WriteBlock(header, reader)
}

func (bf *BlkMFileV2) NeedSize() bool {
	return false
}

func (bf *BlkMFileV2) ReadBlock(w io.Writer) (*BlkHeader, error) {
	return bf.f.ReadBlock(w)
}

func (bf *BlkMFileV2) Keys() []string {
	var ret []string
	bf.meta.Range(func(key, value interface{}) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (bf *BlkMFileV2) ReadBlockByKey(key string, w io.Writer) (int64, error) {
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
			n, _, err = bf.f.compressor.Decompress(w, r)
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

func (bf *BlkMFileV2) Flush() error {
	return bf.f.Flush()
}

type mfileV2Opts struct{}

var MFileV2Opts mfileV2Opts

func (opts mfileV2Opts) WithCompressor(compressor compressor.Compressor) MFileV2Opt {
	return func(f *BlkMFileV2) {
		f.f.WithCompressor(compressor)
	}
}

func (opts mfileV2Opts) WithGzip() MFileV2Opt {
	return func(f *BlkMFileV2) {
		f.f.WithCompressor(compressor.NewGzipCompressor())
	}
}

func (opts mfileV2Opts) WithZlib() MFileV2Opt {
	return func(f *BlkMFileV2) {
		f.f.WithCompressor(compressor.NewZlibCompressor())
	}
}
