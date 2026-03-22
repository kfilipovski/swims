package cmd

import (
	"fmt"
	"time"

	"github.com/kfilipovski/swims/internal/model"
	"github.com/kfilipovski/swims/internal/store"
	"github.com/kfilipovski/swims/internal/usas"
	"github.com/spf13/cobra"
)

var syncTimesCmd = &cobra.Command{
	Use:   "times [swimmer...]",
	Short: "Fetch swim times from USA Swimming and save to local DB",
	Long:  "Fetch swim times. Arguments are swimmer names (case-insensitive match on tracked swimmers). If no swimmers specified, syncs all tracked swimmers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		event, _ := cmd.Flags().GetString("event")
		course, _ := cmd.Flags().GetString("course")
		since, _ := cmd.Flags().GetString("since")
		full, _ := cmd.Flags().GetBool("full")

		if !db.IsInitialized() {
			if err := db.Init(); err != nil {
				return err
			}
			if err := db.EnsureSchema(); err != nil {
				return err
			}
		}

		var eventCode string
		if event != "" {
			eventCode = event + " " + course
		}

		client := usas.NewClient()
		ts := &store.TimeStore{DB: db}
		ss := &store.SwimmerStore{DB: db}

		swimmers, err := resolveSwimmers(ss, args)
		if err != nil {
			return err
		}
		if len(swimmers) == 0 {
			fmt.Println("No tracked swimmers. Run 'swims sync swimmer' first.")
			return nil
		}

		return syncSwimmers(client, ts, ss, swimmers, eventCode, since, full)
	},
}

func resolveSince(swimmer *model.Swimmer, sinceFlag string, full bool) string {
	if full {
		return ""
	}
	if sinceFlag != "" {
		return sinceFlag
	}
	if swimmer != nil && swimmer.TimesSyncedAt != "" {
		return swimmer.TimesSyncedAt
	}
	return ""
}

func syncSwimmers(client *usas.Client, ts *store.TimeStore, ss *store.SwimmerStore, swimmers []model.Swimmer, eventCode string, sinceFlag string, full bool) error {
	today := time.Now().Format("2006-01-02")
	total := 0

	for _, s := range swimmers {
		since := resolveSince(&s, sinceFlag, full)
		if since != "" {
			fmt.Printf("Fetching times for %s since %s...\n", s.FullName, since)
		} else {
			fmt.Printf("Fetching all times for %s...\n", s.FullName)
		}

		times, err := client.FetchTimes(s.SwimmerID, eventCode, since)
		if err != nil {
			fmt.Printf("  Error: %v (skipping)\n", err)
			continue
		}

		if len(times) == 0 {
			fmt.Println("  No new times.")
			continue
		}

		if err := ts.UpsertTimes(s.SwimmerID, times); err != nil {
			fmt.Printf("  Error saving: %v (skipping)\n", err)
			continue
		}
		if err := ss.UpdateAfterSync(s.SwimmerID, today); err != nil {
			fmt.Printf("  Error updating sync date: %v (skipping)\n", err)
			continue
		}
		total += len(times)
		fmt.Printf("  %d time(s)\n", len(times))
	}

	if err := db.Add(); err != nil {
		return err
	}
	if err := db.Commit(fmt.Sprintf("sync: times for %d swimmers", len(swimmers))); err != nil {
		return err
	}

	fmt.Printf("\nSaved %d total time(s) for %d swimmer(s).\n", total, len(swimmers))
	return nil
}

func init() {
	syncTimesCmd.Flags().String("event", "", "optional event filter, e.g. '200 BR'")
	syncTimesCmd.Flags().String("course", "SCY", "course (SCY, SCM, LCM)")
	syncTimesCmd.Flags().String("since", "", "only fetch times after this date (YYYY-MM-DD), overrides last sync date")
	syncTimesCmd.Flags().Bool("full", false, "ignore last sync date, fetch all times")
	syncCmd.AddCommand(syncTimesCmd)
}
