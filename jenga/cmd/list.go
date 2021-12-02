/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/viper"
	"github.com/xfali/jenga"
	"github.com/xfali/jenga/blk"
	"os"

	"github.com/spf13/cobra"
)

var listViper = viper.New()

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list added data keys in jenga file",
	Run: func(cmd *cobra.Command, args []string) {
		jengaPath := rootViper.GetString(ParamJengaFile)
		regexp := listViper.GetString(ParamKeyFilter)
		if jengaPath == "" {
			fatal("Jenga path is empty, add jenga with flags: -j or --jenga-file")
		}
		debug("Jenga file: %s\n", jengaPath)
		var blks jenga.Jenga
		if regexp != "" {
			blks = jenga.NewJenga(jengaPath, jenga.V2(jengablk.BlockV2Opts.WithKeyMatch(regexp)))
		} else {
			blks = jenga.NewJenga(jengaPath, jenga.V2())
		}

		err := blks.Open(jenga.OpFlagReadOnly)
		if err != nil {
			fatal(err.Error())
		}
		defer blks.Close()
		keys := blks.KeyList()
		for _, v := range keys {
			output("%s\n", v)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	fs := listCmd.Flags()
	fs.StringP(ParamKeyFilter, ParamShortKeyFilter, "", "key filter")
	setValue(listViper, fs, ParamKeyFilter, ParamShortKeyFilter)
}
