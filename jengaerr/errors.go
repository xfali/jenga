// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jengaerr

import (
	"fmt"
)

var (
	OpenJengaError            = newError(1001, "Jenga open failed. ")
	OpenRWFlagError           = newError(1002, "%s format flag cannot contains both OpFlagReadOnly and OpFlagWriteOnly. ")
	JengaBrokenError          = newError(1003, "Jenga file format not match, maybe broken. ")
	OpenFlagError             = newError(1004, "Cannot open with flag %d. ")
	DataFormatNotSupportError = newError(1101, "Cannot support format type: %d. ")
	VersionNotSupportError    = newError(1102, "Version: %d not support, [%s] expect: %d. ")
	OpenFileError             = newError(1201, "Cannot open file %s with flag %d. ")

	WriteFlagError            = newError(2001, "Jenga write failed. Need open with OpFlagWriteOnly flag. ")
	WriteFailedError          = newError(2002, "Jenga write failed. ")
	WriteSizeNotMatchError    = newError(2003, "Write size is not match then Header Size! ")
	WriteExistKeyError        = newError(2011, "Block with key %s have been written. ")
	WriteKeyFilteredError     = newError(2012, "Key is filtered, cannot be add. ")
	WriteWithoutSizeFuncError = newError(2021, "%s need a block size map function. ")
	WriteSizeError            = newError(2022, "blkJenga param size %d is Illegal, it must be actual reader data size. ")

	ReadFlagError              = newError(3001, "Jenga read failed. Need open with OpFlagReadOnly flag. ")
	ReadFailedError            = newError(3002, "Jenga read failed. ")
	ReadBlockVarintFailedError = newError(3005, "Cannot parse varint. ")
	ReadKeySizeNotMatchError   = newError(3011, "Read key length is not match record size! ")
	ReadNodeSizeNotMatchError  = newError(3012, "Read size is not match the Node Size! ")
	ReadKeyNotFoundError       = newError(3021, "Block with key: %s not found. ")

	TarNotExistsError        = newError(13001, "Tar file %s not exists. ")
	TarReadFileNotFoundError = newError(13002, "Read from tar failed, Cannot found file: %s. ")
)

type ErrCode struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func newError(code int, message string) *ErrCode {
	return &ErrCode{
		Code:    code,
		Message: message,
	}
}

func (e *ErrCode) Error() string {
	return fmt.Sprintf(`{"module":"jenga", "code":%d, "msg":"%s"}`, e.Code, e.Message)
}

func (e *ErrCode) Equal(err error) bool {
	if o, ok := err.(*ErrCode); ok {
		return e.Code == o.Code
	}
	return false
}

func (e *ErrCode) Format(args ...interface{}) *ErrCode {
	return &ErrCode{
		Code:    e.Code,
		Message: fmt.Sprintf(e.Message, args...),
	}
}
