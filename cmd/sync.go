package cmd

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync data from USA Swimming Data Hub",
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
