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
	"os"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get specify data(file) by key from jenga file",
	Run: func(cmd *cobra.Command, args []string) {
		jengaPath := viper.GetString(ParamShortJengaFile)
		key := viper.GetString(ParamGetKey)
		dest := viper.GetString(ParamTargetFile)
		if jengaPath == "" {
			fatal("jenga path is empty")
		}
		if key == "" {
			fatal("source is empty")
		}
		if dest == "" {
			fatal("dest is empty")
		}
		blks := jenga.NewJenga(jengaPath, jenga.V2())
		err := blks.Open(jenga.OpFlagReadOnly)
		if err != nil {
			fatal(err.Error())
		}
		defer blks.Close()
		f, err := os.OpenFile(dest, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fatal(err.Error())
		}
		defer f.Close()
		_, err = blks.Read(key, f)
		if err != nil {
			fatal(err.Error())
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	fs := getCmd.Flags()
	fs.StringP(ParamGetKey, ParamShortGetKey, "", "key of data")
	setValue(fs, ParamGetKey, ParamShortGetKey)
	fs.StringP(ParamTargetFile, ParamShortTargetFile, "", "Target path to write")
	setValue(fs, ParamTargetFile, ParamShortTargetFile)
}
