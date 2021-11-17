// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import (
	"compress/zlib"
	"io"
)

const (
	TypeZlib = 2
)

type zlibCompressor struct {
	buf []byte
}

type ZlibOpt func(c *zlibCompressor)

func NewZlibCompressor(opts ...ZlibOpt) *zlibCompressor {
	ret := &zlibCompressor{}

	for _, opt := range opts {
		opt(ret)
	}

	if ret.buf == nil {
		ret.buf = make([]byte, DefaultBufferSize)
	}

	return ret
}

// 压缩类型
func (c *zlibCompressor) Type() Type {
	return TypeZlib
}

// 将srcReader的数据压缩至dstWriter
// 参数dstWriter：压缩数据写入的writer
// 参数srcReader：原始数据读取的reader
// 返回before：原始数据大小
// 返回after：压缩后数据大小
// 返回err：发生错误时返回，无错误返回nil
func (c *zlibCompressor) Compress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error) {
	w := NewSizeWriter(dstWriter)
	z := zlib.NewWriter(w)
	r := NewSizeReader(srcReader)
	defer func() {
		e := z.Close()
		if e != nil {
			err = e
		}
		after = w.Size()
	}()
	before, err = io.CopyBuffer(z, r, c.buf)
	return
}

// 将srcReader的数据解压至dstWriter
// 参数dstWriter：解压数据写入的writer
// 参数srcReader：压缩数据读取的reader
// 返回before：压缩数据大小
// 返回after：解压后数据大小
// 返回err：发生错误时返回，无错误返回nil
func (c *zlibCompressor) Decompress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error) {
	r := NewSizeReader(srcReader)
	z, err := zlib.NewReader(r)
	if err != nil {
		return 0, 0, err
	}
	w := NewSizeWriter(dstWriter)
	defer func() {
		e := z.Close()
		if e != nil {
			err = e
		}
		before = r.Size()
	}()
	after, err = io.CopyBuffer(w, z, c.buf)
	return
}

type zlibOpts struct{}

var ZlibOpts zlibOpts

func (opt zlibOpts) WithBuffer(buf []byte) ZlibOpt {
	return func(c *zlibCompressor) {
		c.buf = buf
	}
}

func (opt zlibOpts) BufferSize(size int) ZlibOpt {
	return func(c *zlibCompressor) {
		c.buf = make([]byte, size)
	}
}
