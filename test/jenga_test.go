// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/jenga"
	"github.com/xfali/jenga/blk"
	"os"
	"strings"
	"testing"
)

func TestJengaV1(t *testing.T) {
	blks := jenga.NewJenga("./test.db")
	t.Run("write", func(t *testing.T) {
		err := blks.Open(jenga.OpFlagCreate | jenga.OpFlagWriteOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer blks.Close()
		f, err := os.Open(testFile)
		if err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(testFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		err = blks.Write(testFile, info.Size(), f)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		err := blks.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer blks.Close()
		l := blks.KeyList()
		t.Log("keys:")
		for _, v := range l {
			t.Log(v)
		}
		b := &strings.Builder{}
		_, err = blks.Read(testFile, b)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b.String())
	})
}

func TestJengaV2(t *testing.T) {
	blks := jenga.NewJenga("./test.db", jenga.V2(jengablk.MFileV2Opts.WithGzip()))
	t.Run("write", func(t *testing.T) {
		err := blks.Open(jenga.OpFlagCreate | jenga.OpFlagWriteOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer blks.Close()
		f, err := os.Open(testFile)
		if err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(testFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		err = blks.Write(testFile, info.Size(), f)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		err := blks.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer blks.Close()
		l := blks.KeyList()
		t.Log("keys:")
		for _, v := range l {
			t.Log(v)
		}
		b := &strings.Builder{}
		_, err = blks.Read(testFile, b)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b.String())
	})
}
