// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jenga

import "errors"

var (
	OpenError        = errors.New("Jenga open failed. ")
	WriteFlagError   = errors.New("Jenga write failed. Need open with OpFlagWriteOnly flag. ")
	ReadFlagError    = errors.New("Jenga read failed. Need open with OpFlagReadOnly flag. ")
	WriteFailedError = errors.New("Jenga write failed. ")
	ReadFailedError  = errors.New("Jenga read failed. ")
)
