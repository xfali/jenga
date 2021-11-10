// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package flags

type OpenFlag int
const (
	// 只读
	OpFlagReadOnly  OpenFlag = 1
	// 只写
	OpFlagWriteOnly OpenFlag = 1 << 1
	// 如不存在则创建
	OpFlagCreate    OpenFlag = 1 << 2
)

func (f OpenFlag) CanRead() bool {
	return f&OpFlagReadOnly != 0
}

func (f OpenFlag) CanWrite() bool {
	return f&OpFlagWriteOnly != 0
}

func (f OpenFlag) NeedCreate() bool {
	return f&OpFlagCreate != 0
}
