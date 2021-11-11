// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import (
	"bytes"
	"strings"
	"testing"
)

func TestGzip(t *testing.T) {
	b := bytes.NewBuffer(nil)
	z := NewGzipCompressor()
	n1, n2, err := z.Compress(b, strings.NewReader("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(n1, " ", n2)

	t.Log(b.String())

	s := &strings.Builder{}
	n1, n2, err = z.Decompress(s, b)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(n1, " ", n2)
	t.Log(s.String())
}

func TestZlib(t *testing.T) {
	b := bytes.NewBuffer(nil)
	z := NewZlibCompressor()
	n1, n2, err := z.Compress(b, strings.NewReader("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(n1, " ", n2)

	t.Log(b.String())

	s := &strings.Builder{}
	n1, n2, err = z.Decompress(s, b)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(n1, " ", n2)
	t.Log(s.String())
}

