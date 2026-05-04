package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sessionNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new session",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello, superstar")
	},
}

func init() {
	sessionCmd.AddCommand(sessionNewCmd)
}
