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
	FlagTargetFile       = "flag.target.path"
	ParamTargetFile      = "target-file"
	FlagShortTargetFile  = "flag.short.target.path"
	ParamShortTargetFile = "f"
	FlagJengaFile        = "flag.jenga.path"
	ParamShortJengaFile  = "j"
	ParamGetKey          = "key"
	ParamShortGetKey     = "k"
)

func setValue(fs *pflag.FlagSet, names ...string) bool {
	for _, s := range names {
		f := fs.Lookup(s)
		if f != nil {
			_ = viper.BindPFlag(names[0], f)
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
