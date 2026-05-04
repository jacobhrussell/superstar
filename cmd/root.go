package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "superstar",
	Short: "Superstar CLI",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	path, err := configPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not resolve home directory:", err)
		return
	}
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "warning: failed to read config:", err)
	}
}
