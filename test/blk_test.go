// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/jenga"
	"github.com/xfali/jenga/blk"
	"strings"
	"testing"
)

func TestBlkFile(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		f := jengablk.NewBlkFile("./test.blk")
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		err = f.WriteFile("./test.json")
		if err != nil {
			t.Fatal(err)
		}
		err = f.WriteFile("./test2.json")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		f := jengablk.NewBlkFile("./test.blk")
		err := f.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		buf := &strings.Builder{}
		h, err := f.ReadBlock(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(h)
		t.Log(buf)

		buf.Reset()
		h, err = f.ReadBlock(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(h)
		t.Log(buf)
	})
}
