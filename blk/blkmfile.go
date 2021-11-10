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

type BlkMFile struct {
	f    *BlkFile
	meta sync.Map
}

func NewBlkMFile(path string) *BlkMFile {
	return &BlkMFile{
		f: NewBlkFile(path),
	}
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
	return bf.f.WriteBlock(header, reader)
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

func (bf *BlkMFile) ReadBlockByKey(key string, w io.Writer) error {
	if v, ok := bf.meta.Load(key); ok {
		header := v.(*BlkHeader)
		if header.Invalid() {
			return fmt.Errorf("Block with key: %s not found. ", key)
		}
		err := bf.f.seek(header.offset)
		if err != nil {
			return err
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
			return err
		}
		if n != header.Size {
			return errors.New("Read size is not match then Header Size! ")
		}
		return nil
	} else {
		return fmt.Errorf("Block with key: %s not found. ", key)
	}
}

func (bf *BlkMFile) Flush() error {
	return bf.f.Flush()
}
