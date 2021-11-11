// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package compressor

import "io"

type SizeReader struct {
	r    io.Reader
	size int64
}

func NewSizeReader(r io.Reader) *SizeReader {
	return &SizeReader{
		r: r,
	}
}

func (r *SizeReader) Read(d []byte) (int, error) {
	n, err := r.r.Read(d)
	r.size += int64(n)
	return n, err
}

func (r *SizeReader) Close() error {
	return nil
}

func (r *SizeReader) Size() int64 {
	return r.size
}

type SizeWriter struct {
	w    io.Writer
	size int64
}

func NewSizeWriter(w io.Writer) *SizeWriter {
	return &SizeWriter{
		w: w,
	}
}

func (w *SizeWriter) Write(d []byte) (int, error) {
	n, err := w.w.Write(d)
	w.size += int64(n)
	return n, err
}

func (w *SizeWriter) Close() error {
	return nil
}

func (w *SizeWriter) Size() int64 {
	return w.size
}
