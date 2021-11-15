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
	"io/fs"
	"os"
	"path/filepath"
)

var addViper = viper.New()
// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add file to jenga",
	Run: func(cmd *cobra.Command, args []string) {
		jengaPath := viper.GetString(ParamShortJengaFile)
		key := addViper.GetString(ParamGetKey)
		source := addViper.GetString(ParamSourceFile)
		if jengaPath == "" {
			fatal("Jenga path is empty, add jenga with flags: -j or --jenga-file")
		}
		debug("Jenga file: %s\n", jengaPath)
		if source == "" {
			fatal("Source is empty, add source path with flags: -s or --source-file")
		}
		blks := jenga.NewJenga(jengaPath, jenga.V2Gzip())
		err := blks.Open(jenga.OpFlagCreate | jenga.OpFlagWriteOnly)
		if err != nil {
			fatal(err.Error())
		}
		defer blks.Close()
		info, err := os.Stat(source)
		if err != nil {
			fatal("Source %s not exists", source)
		}
		if info.IsDir() {
			addDir(blks, key, source)
		} else {
			if key == "" {
				key = filepath.Base(source)
			}
			addFile(blks, key, source)
		}
		os.Exit(0)
	},
}

func addDir(j jenga.Jenga, key, source string) {
	source = filepath.Clean(source)
	debug("Add dir: key %s dir: %s\n", key, source)
	err := filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		debug("Visit dir... found file: %s\n", path)
		var fileKey string
		if path == source {
			return nil
		}
		fileKey, _ = filepath.Rel(source, path)
		debug("File rel path: %s\n", fileKey)
		if key != "" {
			fileKey = filepath.Join(key, fileKey)
		}
		return addFile(j, fileKey, path)
	})
	if err != nil {
		fatal(err.Error())
	}
	os.Exit(0)
}

func addFile(j jenga.Jenga, key string, source string) error {
	debug("Add file: key %s file path: %s \n", key, source)
	info, err := os.Stat(source)
	if err == nil {
		if info.IsDir() {
			fatal("File %s is a directory.", source)
		}
	}
	f, err := os.Open(source)
	if err != nil {
		fatal(err.Error())
	}
	defer f.Close()
	err = j.Write(key, -1, f)
	if err != nil {
		fatal(err.Error())
	}
	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)

	fs := addCmd.Flags()
	fs.StringP(ParamSourceFile, ParamShortSourceFile, "", "Source file to add")
	setValue(addViper, fs, ParamSourceFile, ParamShortSourceFile)

	fs.StringP(ParamGetKey, ParamShortGetKey, "", "key of data")
	setValue(addViper, fs, ParamGetKey, ParamShortGetKey)
}
