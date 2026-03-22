package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kfilipovski/swims/internal/dolt"
	"github.com/spf13/cobra"
)

var (
	dataDir string
	db      *dolt.Dolt
)

var rootCmd = &cobra.Command{
	Use:   "swims",
	Short: "USA Swimming Data Hub CLI with Dolt persistence",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dataDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			dataDir = cwd
		}

		dataDir, _ = filepath.Abs(dataDir)
		db = dolt.New(dataDir)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "path to dolt database directory (default: current directory)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
