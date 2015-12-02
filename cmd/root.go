// Copyright © 2015 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os/user"
	"path"
)

var probesFlag int
var concurrentFlag bool

var cfgFile string
var authAddr string
var dataAddr string
var metaAddr string
var log *logrus.Logger

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "clawiobench",
	Short: "ClawIO Benchmarking Tool",
	Long:  `This tools benchmarks different actions that can be performed against ClawIO.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.clawiobench.yaml)")
	RootCmd.PersistentFlags().IntVar(&probesFlag, "probes", 1, "The number of tests to perform")
	RootCmd.PersistentFlags().BoolVar(&concurrentFlag, "concurrent", false, "If set operations are run concurrently")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	initLogger()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".clawiobench.config") // name of config file (without extension)

	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	authAddr = viper.GetString("CLAWIO_BENCH_AUTH_ADDR")
	if authAddr == "" {
		fmt.Println("You have to specify the auth unit address")
		os.Exit(1)
	}
	dataAddr = viper.GetString("CLAWIO_BENCH_DATA_ADDR")
	if dataAddr == "" {
		fmt.Println("You have to specify the data unit address")
		os.Exit(1)
	}
	metaAddr = viper.GetString("CLAWIO_BENCH_META_ADDR")
	if metaAddr == "" {
		fmt.Println("You have to specify the meta unit address")
		os.Exit(1)
	}
}

// initLogger instantiate a logger instance that writes to $HOME/.clawiobench.log
func initLogger() {
	u, _ := user.Current()
	home := u.HomeDir
	fd, err := os.OpenFile(path.Join(home, ".clawiobench.log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log = logrus.New()
	log.Out = fd
}

func getToken() (string, error) {

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	token, err := ioutil.ReadFile(path.Join(u.HomeDir, ".clawiobench", "credentials"))
	if err != nil {
		return "", err
	}

	return string(token), nil
}
