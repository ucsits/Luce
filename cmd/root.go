package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ucsits/Luce/fsmgr"
	"github.com/ucsits/Luce/rpc"
)

var (
	port    string
	dataDir string
)

var rootCmd = &cobra.Command{
	Use:   "luce",
	Short: "luce is a blockchain organizational transparency application",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		chain := fsmgr.LoadOrCreate(dataDir)

		cfg := rpc.DefaultConfig()
		cfg.Port = port
		cfg.DataDir = dataDir

		server := rpc.NewServer(cfg, chain)
		if err := server.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "8080", "RPC server port")
	rootCmd.Flags().StringVar(&dataDir, "data-dir", ".", "blockchain data directory")
}
