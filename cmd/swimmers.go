package cmd

import (
	"fmt"

	"github.com/kfilipovski/swims/internal/format"
	"github.com/kfilipovski/swims/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked swimmers",
	RunE: func(cmd *cobra.Command, args []string) error {
		lsc, _ := cmd.Flags().GetString("lsc")

		ss := &store.SwimmerStore{DB: db}
		swimmers, err := ss.List(lsc)
		if err != nil {
			return err
		}

		if len(swimmers) == 0 {
			fmt.Println("No swimmers tracked. Run 'swims add' first.")
			return nil
		}

		tbl := format.NewTable("SwimmerID", "Name", "Age", "LSC", "Club", "Last Sync")
		for _, s := range swimmers {
			sync := s.TimesSyncedAt
			if sync == "" {
				sync = "never"
			}
			tbl.Row(fmt.Sprintf("%d", s.SwimmerID), s.FullName, fmt.Sprintf("%d", s.Age), s.LscCode, s.ClubName, sync)
		}
		tbl.Flush()
		return nil
	},
}

func init() {
	listCmd.Flags().String("lsc", "", "filter by LSC code")
	rootCmd.AddCommand(listCmd)
}
