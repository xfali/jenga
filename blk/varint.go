// Copyright (C) 2019-2021, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package jengablk

import "io"

const (
    MaxVarUintBufSize = 10
)

type VarInt struct {
    data [MaxVarUintBufSize]byte
    cur  int
}

//读取可变整数，完成返回true,未完成还需继续读取返回false
func (v *VarInt) Load(d []byte) bool {
    for n := 0; n < len(d); n++ {
        if v.LoadByte(d[n]) {
            return true
        }
    }
    return false
}

//读取可变整数，完成返回true,未完成还需继续读取返回false
func (v *VarInt) LoadByte(d byte) bool {
    v.data[v.cur] = d
    v.cur++
    if d>>7 == 0 {
        return true
    }
    return false
}

func NewFromReader(r io.Reader) *VarInt {
    v := &VarInt{}
    for {
        b, _, err := v.LoadFromReader(r)
        if err != nil {
            return nil
        }
        if b {
            return v
        }
    }
}

func (v *VarInt) InitFromUInt64(x uint64) {
    n := EncodeVaruint(v.data[:], x)
    v.cur = n
}

func (v *VarInt) LoadFromReader(r io.Reader) (bool, int, error) {
    size := 0
    for {
        n, err := r.Read(v.data[v.cur : v.cur+1])
        if err != nil {
            return false, n, err
        }
        size += n
        if v.data[v.cur]>>7 == 0 {
            v.cur++
            return true, size, nil
        }
        v.cur++
    }
    return false, size, nil
}

//长度
func (v *VarInt) Length() int {
    return v.cur
}

//读取可变整数，完成返回true,未完成还需继续读取返回false
func (v *VarInt) Bytes() []byte {
    return v.data[:v.cur]
}

func (v *VarInt) ToInt() int64 {
    return DecodeVarint(v.Bytes())
}

func (v *VarInt) ToUint() uint64 {
    return DecodeVaruint(v.Bytes())
}

//Base 128 Varint的介绍：https://developers.google.com/protocol-buffers/docs/encoding
//Base 128 Varint，为什么叫128？其实，就是因为只采用7bit的空间来存储有效数据，7bit当然最大只能存储128了。
//常规的一个byte是8个bit位，但在Base 128 Varint编码中，将最高的第8位用来作为一个标志位，
//如果这一位是1，就表示这个字节后面还有其他字节，如果这个位是0的话，就表示这是最后一个字节了，
//这样一来，就可以准确的知道一个整数的结束位置了。
func EncodeVaruint(buf []byte, x uint64) int {
    n := 0
    for x > 127 {
        buf[n] = byte(0x80 | (x & 0x7F))
        n++
        x >>= 7
    }
    buf[n] = byte(x)
    n++
    return n
}

func CalcVaruintLen(x uint64) int {
    n := 0
    for x > 127 {
        n++
        x >>= 7
    }
    n++
    return n
}

func DecodeVaruint(buf []byte) uint64 {
    var n, shift uint = 0, 0
    var x, c uint64 = 0, 0
    for ; shift < 64; shift += 7 {
        c = uint64(buf[n])
        n++
        x |= uint64((c & 0x7F) << shift)
        if (c & 0x80) == 0 {
            break
        }
    }

    return x
}

func DecodeVarint(buf []byte) int64 {
    //ux := DecodeVaruint(buf) // ok to continue in presence of error
    //x := int64(ux >> 1)
    //if ux&1 != 0 {
    //    x = ^x
    //}
    //return x
    return int64(DecodeVaruint(buf))
}
