// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xfali/jenga/blk"
	"github.com/xfali/jenga/compressor"
	"os"
)

// listCmd represents the list command
var info = &cobra.Command{
	Use:   "info",
	Short: "show jenga file info",
	Run: func(cmd *cobra.Command, args []string) {
		jengaPath := rootViper.GetString(ParamJengaFile)
		if jengaPath == "" {
			fatal("Jenga path is empty, add jenga with flags: -j or --jenga-file")
		}
		debug("Jenga file: %s\n", jengaPath)
		f, err := os.Open(jengaPath)
		if err != nil {
			fatal("jenga file %s open failed: %v. ", jengaPath, err)
		}
		defer f.Close()

		h, err := jengablk.ReadFileHeader(f)
		if err != nil {
			fatal("Read jenga file %s failed: %v. ", jengaPath, err)
		}
		output("Version:\t%d\n", h.Version)
		output("Data format:\t%d (%s)\n", h.DataFormat, compressor.GetName(h.DataFormat))
		output("Reserve:\t%d\n", h.Reserve)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(info)
}
