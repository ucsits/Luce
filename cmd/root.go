package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ucsits/Luce/fsmgr"
)

var rootCmd = &cobra.Command{
	Use:   "luce",
	Short: "luce is a blockchain organizational transparency application",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fsmgr.Bootstrap()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
