package commands

import (
	"github.com/spf13/cobra"
	"os"
)

var listenPort int
var targetPeer string

var rootCmd = &cobra.Command{
	Use:   "ChainOfBots",
	Short: "",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&listenPort, "listenPort", "p", 1000, "listen Port number")
	rootCmd.PersistentFlags().StringVarP(&targetPeer, "targetPeer", "t", "", "connect to one peer")
}
