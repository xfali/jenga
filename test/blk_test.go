// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"errors"
	"github.com/xfali/jenga"
	"github.com/xfali/jenga/blk"
	"github.com/xfali/jenga/compressor"
	"io"
	"os"
	"strings"
	"testing"
)

func TestBlkFileV1(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		f := jengablk.NewBlkFile("./test.blk")
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = f.WriteFile("./test.json")
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.WriteFile("./test2.json")
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
		for {
			buf.Reset()
			h, err := f.ReadBlock(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.Log("finished")
					return
				}
				t.Fatal(err)
			}
			t.Log(h)
			t.Log(buf)
		}
	})
}

func TestBlkFileV2(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		f := jengablk.NewBlkFileV2("./test.blk")
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = f.WriteFile("./test.json")
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.WriteFile("./test2.json")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		f := jengablk.NewBlkFileV2("./test.blk")
		err := f.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		buf := &strings.Builder{}
		for {
			buf.Reset()
			h, err := f.ReadBlock(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.Log("finished")
					return
				}
				t.Fatal(err)
			}
			t.Log(h)
			t.Log(buf)
		}
	})
}

func TestV1BlockFile(t *testing.T) {
	t.Run("write1", func(t *testing.T) {
		f := jengablk.NewV1BlockFile("./test.blk")
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = f.WriteFile("./test.json")
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.WriteFile("./test.json")
		if err == nil {
			t.Fatal("cannot write same file")
		}
	})

	t.Run("write2", func(t *testing.T) {
		f := jengablk.NewV1BlockFile("./test.blk")
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = f.WriteFile("./test2.json")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		f := jengablk.NewV1BlockFile("./test.blk")
		err := f.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		buf := &strings.Builder{}
		for {
			buf.Reset()
			h, err := f.ReadBlock(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.Log("finished")
					break
				}
				t.Fatal(err)
			}
			t.Log(h)
			t.Log(buf)
		}
		keys := f.Keys()
		for _, v := range keys {
			buf.Reset()
			_, err := f.ReadBlockByKey(v, buf)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("=====key==== ", v)
			t.Log(buf)
		}
	})
}

func TestV2BlockFile(t *testing.T) {
	_ = compressor.NewGzipCompressor()
	f := jengablk.NewV2BlockFile("./test.blk", jengablk.BlockV2Opts.WithZlib())
	t.Run("write1", func(t *testing.T) {
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = f.WriteFile("./test.json")
		if err != nil {
			t.Fatal(err)
		}
		info, _ := os.Stat("./test.json")
		t.Log("size:", info.Size())
		_, err = f.WriteFile("./test.json")
		if err == nil {
			t.Fatal("cannot write same file")
		}
	})

	t.Run("write2", func(t *testing.T) {
		err := f.Open(jenga.OpFlagWriteOnly | jenga.OpFlagCreate)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		info, _ := os.Stat("./test2.json")
		t.Log("size:", info.Size())
		_, err = f.WriteFile("./test2.json")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		err := f.Open(jenga.OpFlagReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		buf := &strings.Builder{}
		for {
			buf.Reset()
			h, err := f.ReadBlock(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.Log("finished")
					break
				}
				t.Fatal(err)
			}
			t.Log(h)
			t.Log(buf)
		}
		keys := f.Keys()
		for _, v := range keys {
			buf.Reset()
			_, err := f.ReadBlockByKey(v, buf)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("=====key==== ", v)
			t.Log(buf)
		}
	})
}
