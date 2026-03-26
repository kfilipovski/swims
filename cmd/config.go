package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage swims configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cfgManager.Load()
		if err != nil {
			return err
		}
		fmt.Printf("Config dir:   %s\n", cfgManager.RootDir())
		fmt.Printf("Data dir:     %s\n", cfgManager.DataDir())
		fmt.Printf("Config file:  %s\n", cfgManager.ConfigPath())
		fmt.Printf("Session file: %s\n", cfgManager.SessionPath())
		if cfg.DefaultClub == "" {
			fmt.Println("Default club: (not set)")
		} else {
			fmt.Printf("Default club: %s\n", cfg.DefaultClub)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cfgManager.Load()
		if err != nil {
			return err
		}

		switch strings.ToLower(args[0]) {
		case "club", "default-club":
			cfg.DefaultClub = strings.TrimSpace(args[1])
		default:
			return fmt.Errorf("unknown config key %q", args[0])
		}

		if err := cfgManager.Save(cfg); err != nil {
			return err
		}
		appConfig = cfg
		fmt.Printf("Set %s to %q in %s\n", args[0], args[1], cfgManager.ConfigPath())
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Unset a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cfgManager.Load()
		if err != nil {
			return err
		}

		switch strings.ToLower(args[0]) {
		case "club", "default-club":
			cfg.DefaultClub = ""
		default:
			return fmt.Errorf("unknown config key %q", args[0])
		}

		if err := cfgManager.Save(cfg); err != nil {
			return err
		}
		appConfig = cfg
		fmt.Printf("Unset %s in %s\n", args[0], cfgManager.ConfigPath())
		return nil
	},
}

var configClearSessionCmd = &cobra.Command{
	Use:   "clear-session",
	Short: "Clear the cached USA Swimming session token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgManager.ClearSession(); err != nil {
			return err
		}
		fmt.Printf("Cleared cached session in %s\n", cfgManager.SessionPath())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configClearSessionCmd)
	rootCmd.AddCommand(configCmd)
}
