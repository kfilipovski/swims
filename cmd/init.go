package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the dolt database and schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.Init(); err != nil {
			return fmt.Errorf("initializing dolt: %w", err)
		}
		if err := db.EnsureSchema(); err != nil {
			return fmt.Errorf("creating schema: %w", err)
		}
		if err := db.Add(); err != nil {
			return err
		}
		if err := db.Commit("init: create schema"); err != nil {
			return err
		}
		fmt.Println("Database initialized.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
