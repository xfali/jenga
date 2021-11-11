// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import "io"

type Type uint16

type Compressor interface {
	// 压缩类型
	Type() Type

	// 将srcReader的数据压缩至dstWriter
	// 参数dstWriter：压缩数据写入的writer
	// 参数srcReader：原始数据读取的reader
	// 返回before：原始数据大小
	// 返回after：压缩后数据大小
	// 返回err：发生错误时返回，无错误返回nil
	Compress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error)

	// 将srcReader的数据解压至dstWriter
	// 参数dstWriter：解压数据写入的writer
	// 参数srcReader：压缩数据读取的reader
	// 返回before：压缩数据大小
	// 返回after：解压后数据大小
	// 返回err：发生错误时返回，无错误返回nil
	Decompress(dstWriter io.Writer, srcReader io.Reader) (before int64, after int64, err error)
}

func (t Type) Value() uint16 {
	return uint16(t)
}

func ToType(t uint16) Type {
	return Type(t)
}
