// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jenga

import (
	"github.com/xfali/jenga/flags"
	"io"
)

const (
	// 只读
	OpFlagReadOnly OpenFlag = 1
	// 只写
	OpFlagWriteOnly OpenFlag = 1 << 1
	// 如不存在则创建
	OpFlagCreate OpenFlag = 1 << 2
)

type OpenFlag = flags.OpenFlag

type Jenga interface {
	// 打开
	// flag：打开标志，一般而言不能同时包含OpFlagReadOnly和OpFlagWriteOnly
	Open(flag OpenFlag) error

	// 关闭
	Close() error

	// 获得Key列表
	KeyList() []string

	Writer
	Reader
}

type Writer interface {
	// 使用key保存数据
	// key: 数据关联的key
	// size: 数据长度，选填0或者实际的长度（根据具体的实现）
	// r: 写入数据的reader
	// err: 当出错时返回
	Write(key string, size int64, r io.Reader) (err error)
}

type Reader interface {
	// 使用key获取数据
	// key: 数据关联的key
	// w: 接收数据的writer
	// size: 读取数据的长度
	// err: 当出错时返回
	Read(key string, w io.Writer) (size int64, err error)
}
