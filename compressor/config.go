// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import (
	"fmt"
	"sync"
)

const (
	DefaultBufferSize = 32 * 1024
)

func init() {
	nameMap.Store(uint16(TypeNone), "No Compress")
	nameMap.Store(uint16(TypeGzip), "gzip")
	nameMap.Store(uint16(TypeZlib), "zlib")
}

var nameMap = sync.Map{}

func GetName(compressType uint16) string {
	if v, ok := nameMap.Load(compressType); ok {
		return v.(string)
	} else {
		return fmt.Sprintf("Unknown Compress Type: %d", compressType)
	}
}

func RegisterType(compressType uint16, compressName string) bool {
	_, loaded := nameMap.LoadOrStore(compressType, compressName)
	return !loaded
}
