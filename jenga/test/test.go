// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package main

import (
	"github.com/spf13/cobra"
	"github.com/xfali/jenga/jenga/cmd"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "zz",
		Short: "A compress tools",
	}
	cmd.AddTo(rootCmd)
	cobra.CheckErr(rootCmd.Execute())
}
