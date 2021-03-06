package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jenga",
	Short: "A compress tools",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	addFlags(rootCmd)
	cobra.CheckErr(rootCmd.Execute())
}

var rootViper = viper.New()

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.main.yaml)")
}

func addFlags(root *cobra.Command) {
	fs := root.PersistentFlags()
	fs.StringP(ParamJengaFile, ParamShortJengaFile, "", "Path of jenga")
	setValue(rootViper, fs, ParamJengaFile, ParamShortJengaFile)

	fs.BoolP(ParamLogVerbose, ParamShortLogVerbose, false, "output detail")
	setValue(rootViper, fs, ParamShortLogVerbose, ParamLogVerbose)
}

func AddTo(root *cobra.Command) {
	addFlags(root)
	root.AddCommand(rootCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".main" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".main")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
