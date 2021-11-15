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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xfali/jenga"
	"os"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add file to jenga",
	Run: func(cmd *cobra.Command, args []string) {
		jengaPath := viper.GetString(ParamShortJengaFile)
		source := viper.GetString(ParamSourceFile)
		if jengaPath == "" {
			fatal("jenga path is empty")
		}
		if source == "" {
			fatal("source is empty")
		}
		blks := jenga.NewJenga(jengaPath, jenga.V2Gzip())
		err := blks.Open(jenga.OpFlagCreate | jenga.OpFlagWriteOnly)
		if err != nil {
			fatal(err.Error())
		}
		defer blks.Close()
		f, err := os.Open(source)
		if err != nil {
			fatal(err.Error())
		}
		defer f.Close()
		err = blks.Write(source, -1, f)
		if err != nil {
			fatal(err.Error())
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	fs := addCmd.Flags()
	fs.StringP(ParamSourceFile, ParamShortSourceFile, "", "Source file to add")
	setValue(fs, ParamSourceFile, ParamShortSourceFile)
}
