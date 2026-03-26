package cmd

import (
	"os"
	"path/filepath"

	"github.com/kfilipovski/swims/internal/config"
	"github.com/kfilipovski/swims/internal/dolt"
	"github.com/spf13/cobra"
)

var (
	dataDir    string
	configDir  string
	db         *dolt.Dolt
	cfgManager *config.Manager
	appConfig  *config.AppConfig
)

var rootCmd = &cobra.Command{
	Use:   "swims",
	Short: "USA Swimming Data Hub CLI with Dolt persistence",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfgManager, err = config.NewManager(configDir)
		if err != nil {
			return err
		}
		if err := cfgManager.EnsureDir(); err != nil {
			return err
		}
		appConfig, err = cfgManager.Load()
		if err != nil {
			return err
		}

		if dataDir == "" {
			dataDir = cfgManager.DataDir()
		}

		dataDir, err = filepath.Abs(dataDir)
		if err != nil {
			return err
		}
		db = dolt.New(dataDir)
		return nil
	},
}

func init() {
	home, _ := os.UserHomeDir()
	defaultConfigDir := filepath.Join(home, ".swims")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", defaultConfigDir, "path to swims config directory")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "path to dolt database directory (default: ~/.swims)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
