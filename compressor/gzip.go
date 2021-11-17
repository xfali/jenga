// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import (
	"compress/gzip"
	"io"
)

const (
	TypeGzip = 1

	NoCompression      GzipCompressLevel = gzip.NoCompression
	BestSpeed          GzipCompressLevel = gzip.BestSpeed
	BestCompression    GzipCompressLevel = gzip.BestCompression
	DefaultCompression GzipCompressLevel = gzip.DefaultCompression
	HuffmanOnly        GzipCompressLevel = gzip.HuffmanOnly
)

type GzipCompressLevel int

type gzipCompressor struct {
	level GzipCompressLevel
	buf   []byte
}

type GzipOpt func(c *gzipCompressor)

func NewGzipCompressor(opts ...GzipOpt) *gzipCompressor {
	ret := &gzipCompressor{
		level: DefaultCompression,
	}
	for _, opt := range opts {
		opt(ret)
	}
	if ret.buf == nil {
		ret.buf = make([]byte, DefaultBufferSize)
	}
	return ret
}

// 压缩类型
func (c *gzipCompressor) Type() Type {
	return TypeGzip
}

// 将srcReader的数据压缩至dstWriter
// 参数dstWriter：压缩数据写入的writer
// 参数srcReader：原始数据读取的reader
// 返回before：原始数据大小
// 返回after：压缩后数据大小
// 返回err：发生错误时返回，无错误返回nil
func (c *gzipCompressor) Compress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error) {
	w := NewSizeWriter(dstWriter)
	z, err := gzip.NewWriterLevel(w, int(c.level))
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		e := z.Close()
		if e != nil {
			err = e
		}
		after = w.Size()
	}()
	before, err = io.CopyBuffer(z, srcReader, c.buf)
	return
}

// 将srcReader的数据解压至dstWriter
// 参数dstWriter：解压数据写入的writer
// 参数srcReader：压缩数据读取的reader
// 返回before：压缩数据大小
// 返回after：解压后数据大小
// 返回err：发生错误时返回，无错误返回nil
func (c *gzipCompressor) Decompress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error) {
	r := NewSizeReader(srcReader)
	z, err := gzip.NewReader(r)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		e := z.Close()
		if e != nil {
			err = e
		}
		before = r.Size()
	}()
	after, err = io.CopyBuffer(dstWriter, z, c.buf)
	return
}

type gzipOpts struct{}

var GzipOpts gzipOpts

func (opt gzipOpts) WithBuffer(buf []byte) GzipOpt {
	return func(c *gzipCompressor) {
		c.buf = buf
	}
}

func (opt gzipOpts) BufferSize(size int) GzipOpt {
	return func(c *gzipCompressor) {
		c.buf = make([]byte, size)
	}
}

func (opt gzipOpts) Level(level GzipCompressLevel) GzipOpt {
	return func(c *gzipCompressor) {
		c.level = level
	}
}
