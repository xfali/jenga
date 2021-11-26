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
	OpFlagReadOnly = flags.OpFlagReadOnly
	// 只写
	OpFlagWriteOnly = flags.OpFlagWriteOnly
	// 如不存在则创建
	OpFlagCreate = flags.OpFlagCreate
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
	// param key: 数据关联的key
	// param r: 写入数据的reader
	// return size: 写入数据的长度
	// return err: 当出错时返回
	Write(key string, r io.Reader) (size int64, err error)
}

type Reader interface {
	// 使用key获取数据
	// param key: 数据关联的key
	// param w: 接收数据的writer
	// return size: 读取数据的长度
	// return err: 当出错时返回
	Read(key string, w io.Writer) (size int64, err error)
}
