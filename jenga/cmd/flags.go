// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package cmd

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

const (
	FlagSourceFile       = "flag.source.path"
	ParamSourceFile      = "source-file"
	FlagShortSourceFile  = "flag.short.source.path"
	ParamShortSourceFile = "s"
	ParamJengaGzip       = "compress-gzip"
	ParamShortJengaGzip  = "g"
	ParamJengaZlib       = "compress-zlib"
	ParamShortJengaZlib  = "z"
	FlagTargetFile       = "flag.target.path"
	ParamTargetFile      = "target-file"
	FlagShortTargetFile  = "flag.short.target.path"
	ParamShortTargetFile = "f"
	FlagJengaFile        = "flag.jenga.path"
	ParamJengaFile       = "jenga-file"
	ParamShortJengaFile  = "j"
	ParamGetKey          = "key"
	ParamShortGetKey     = "k"
	ParamKeyFilter       = "key-regexp"
	ParamShortKeyFilter  = "x"
	ParamLogVerbose      = "verbose"
	ParamShortLogVerbose = "v"
)

func setValue(v *viper.Viper, fs *pflag.FlagSet, names ...string) bool {
	for _, s := range names {
		f := fs.Lookup(s)
		if f != nil {
			if v == nil {
				_ = viper.BindPFlag(names[0], f)
			} else {
				_ = v.BindPFlag(names[0], f)
			}

			return true
		}
	}
	return false
}

func fatal(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(-1)
}

func output(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

func debug(format string, args ...interface{}) {
	v := rootViper.GetBool(ParamShortLogVerbose)
	if v {
		_, _ = fmt.Fprintf(os.Stdout, format, args...)
	}
}
