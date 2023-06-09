/*
Copyright © 2022 SeanWang
*/
package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/version"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:     "packagist-mirror-next",
		Version: version.GetVersion(),
		Short:   "Packagist Mirror Next",
		Long:    "Packagist Mirror Next is a tool for mirroring Packagist.org repository.",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 初始化配置
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(logx.InitLogx)

	// 注册config配置
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "etc/packagist-mirror.yaml", "config filex (default is $PWD/etc/packagist-mirror.yaml)")
}

// initConfig 初始化配置
func initConfig() {
	fmt.Println("111", cfgFile)
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}
