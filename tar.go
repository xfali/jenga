// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package jenga

import (
	"archive/tar"
	"errors"
	"github.com/xfali/jenga/jengaerr"
	"io"
	"os"
	"path/filepath"
)

type tarJenga struct {
	path string
	file *os.File
	flag OpenFlag
	w    *tar.Writer
}

func NewTar(tarPath string) *tarJenga {
	return &tarJenga{
		path: tarPath,
		file: nil,
	}
}

func (jenga *tarJenga) Open(flag OpenFlag) error {
	if flag.CanWrite() && flag.CanRead() {
		return jengaerr.OpenRWFlagError.Format("Tar")
	}
	jenga.flag = flag

	_, err := os.Stat(jenga.path)
	if err == nil {
		if jenga.flag.CanRead() {
			f, err := os.Open(jenga.path)
			if err != nil {
				return err
			}
			jenga.file = f
			return nil
		} else if jenga.flag.CanWrite() {
			f, err := os.OpenFile(jenga.path, os.O_WRONLY, 0666)
			if err != nil {
				return err
			}
			if _, err = f.Seek(-1024, io.SeekEnd); err != nil {
				_ = f.Close()
				return jengaerr.OpenJengaError
			}
			jenga.file = f
			jenga.w = tar.NewWriter(jenga.file)
			return nil
		}
	} else {
		if jenga.flag.NeedCreate() {
			f, err := os.OpenFile(jenga.path, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			jenga.file = f
			jenga.w = tar.NewWriter(jenga.file)
			return nil
		}
	}
	return jengaerr.OpenJengaError
}

func (jenga *tarJenga) KeyList() []string {
	if jenga.file != nil {
		if f, err := os.Open(jenga.path); err == nil {
			defer f.Close()
			var ret []string
			r := tar.NewReader(f)
			for {
				h, err := r.Next()
				if err != nil {
					return ret
				}
				ret = append(ret, h.Name)
			}
		}
	}
	return nil
}

// 强制同步数据
func (jenga *tarJenga) Sync() error {
	return jenga.file.Sync()
}

func (jenga *tarJenga) Write(path string, r io.Reader) (int64, error) {
	if !jenga.flag.CanWrite() {
		return 0, jengaerr.WriteFlagError
	}
	if info, err := os.Stat(path); err == nil {
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return 0, jengaerr.WriteFailedError
		}
		err = jenga.w.WriteHeader(hdr)
		if err != nil {
			return 0, jengaerr.WriteFailedError
		}
		file, err := os.Open(path)
		if err != nil {
			return 0, jengaerr.WriteFailedError
		}
		defer file.Close()

		n, err := io.Copy(jenga.w, file)
		if err != nil {
			return n, jengaerr.WriteFailedError
		}
		return n, nil
	} else {
		return 0, jengaerr.TarNotExistsError.Format(path)
	}
}

func (jenga *tarJenga) Read(path string, w io.Writer) (int64, error) {
	if !jenga.flag.CanRead() {
		return 0, jengaerr.ReadFlagError
	}
	r := tar.NewReader(jenga.file)
	path = filepath.Base(path)
	for {
		h, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, jengaerr.TarReadFileNotFoundError.Format(path)
			} else {
				return 0, err
			}
		}
		if h.Name == path {
			n, err := io.Copy(w, r)
			if err != nil {
				return n, jengaerr.ReadFailedError
			}
			return n, nil
		}
	}
}

func (jenga *tarJenga) Close() (err error) {
	if jenga.w != nil {
		e := jenga.w.Close()
		if e != nil {
			err = e
		}
	}
	if jenga.file != nil {
		e := jenga.file.Close()
		if e != nil {
			err = e
		}
	}
	return nil
}
