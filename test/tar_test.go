// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/jenga"
	"os"
	"strings"
	"testing"
)

var testFile = "./test2.json"

func TestTar(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		tar := jenga.NewTar("./test.tar")
		err := tar.Open(jenga.OpFlagCreate | jenga.OpFlagWriteOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer tar.Close()
		f, err := os.Open(testFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		err = tar.Write(testFile, 0, f)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		tar := jenga.NewTar("./test.tar")
		err := tar.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer tar.Close()
		l := tar.KeyList()
		t.Log("keys:")
		for _, v := range l {
			t.Log(v)
		}
		b := &strings.Builder{}
		_, err = tar.Read(testFile, b)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b.String())
	})
}
