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
	"path/filepath"

	"github.com/spf13/cobra"
)

var getViper = viper.New()
// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get specify data(file) by key from jenga file",
	Run: func(cmd *cobra.Command, args []string) {
		jengaPath := rootViper.GetString(ParamJengaFile)
		key := getViper.GetString(ParamGetKey)
		dest := getViper.GetString(ParamTargetFile)
		if jengaPath == "" {
			fatal("Jenga path is empty, add jenga with flags: -j or --jenga-file")
		}
		debug("Get data from jenga file: %s\n", jengaPath)
		if dest == "" {
			fatal("Target is empty, add target with flags: -f or --target-file")
		}
		info, err := os.Stat(dest)
		isDir := false
		if err != nil {
			if key == "" {
				err = os.MkdirAll(dest, 0666)
				isDir = true
				if err != nil {
					fatal(err.Error())
				}
			}
		} else {
			isDir = info.IsDir()
		}

		blks := jenga.NewJenga(jengaPath, jenga.V2())
		err = blks.Open(jenga.OpFlagReadOnly)
		if err != nil {
			fatal(err.Error())
		}
		defer blks.Close()
		if isDir {
			if key != "" {
				getFile(blks, key, filepath.Join(dest, key))
			}
			getDir(blks, dest)
		} else {
			getFile(blks, key, dest)
		}
		os.Exit(0)
	},
}

func getDir(j jenga.Jenga, target string) {
	debug("Get file to dir %s\n", target)
	keys := j.KeyList()
	for _, v := range keys {
		getFile(j, v, filepath.Join(target, v))
	}
	os.Exit(0)
}

func getFile(j jenga.Jenga, key string, target string) {
	debug("Get file: key %s file to %s\n", key, target)
	_, err := os.Stat(target)
	if err == nil {
		fatal("Get file failed, file %s is exists", target)
	}
	f, err := os.OpenFile(target, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fatal(err.Error())
	}
	defer f.Close()
	n, err := j.Read(key, f)
	if err != nil {
		fatal(err.Error())
	} else {
		debug("Get file: %s success, size: %d\n", target, n)
	}
}

func init() {
	rootCmd.AddCommand(getCmd)

	fs := getCmd.Flags()
	fs.StringP(ParamGetKey, ParamShortGetKey, "", "key of data")
	setValue(getViper, fs, ParamGetKey, ParamShortGetKey)
	fs.StringP(ParamTargetFile, ParamShortTargetFile, "", "Target path to write")
	setValue(getViper, fs, ParamTargetFile, ParamShortTargetFile)
}
