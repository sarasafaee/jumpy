package commands

import (
	"github.com/spf13/cobra"
	"myBlockchain/internal/app"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a DAG Node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		app.Start(listenPort, targetPeer)
	},
}
