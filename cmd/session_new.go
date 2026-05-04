package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sessionNewDir string

var sessionNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new session",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello, superstar")
		fmt.Println("dir:", sessionNewDir)
	},
}

func init() {
	sessionNewCmd.Flags().StringVar(&sessionNewDir, "dir", "", "directory to open in the session")
	sessionCmd.AddCommand(sessionNewCmd)
}
