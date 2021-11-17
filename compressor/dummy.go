// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import (
	"io"
)

const(
	TypeNone = 0
)

type bufferCompressor struct {
	buf []byte
}

func NewBufferCompressor(size int) *bufferCompressor {
	if size <= 0 {
		size = DefaultBufferSize
	}
	return &bufferCompressor{
		buf: make([]byte, size),
	}
}

// 压缩类型
func (c *bufferCompressor) Type() Type {
	return TypeNone
}

// 将srcReader的数据压缩至dstWriter
// 参数dstWriter：压缩数据写入的writer
// 参数srcReader：原始数据读取的reader
// 返回before：原始数据大小
// 返回after：压缩后数据大小
// 返回err：发生错误时返回，无错误返回nil
func (c *bufferCompressor) Compress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error) {
	n, err := io.CopyBuffer(dstWriter, srcReader, c.buf)
	return n, n, err
}

// 将srcReader的数据解压至dstWriter
// 参数dstWriter：解压数据写入的writer
// 参数srcReader：压缩数据读取的reader
// 返回before：压缩数据大小
// 返回after：解压后数据大小
// 返回err：发生错误时返回，无错误返回nil
func (c *bufferCompressor) Decompress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error) {
	n, err := io.CopyBuffer(dstWriter, srcReader, c.buf)
	return n, n, err
}
