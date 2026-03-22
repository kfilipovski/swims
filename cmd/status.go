package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !db.IsInitialized() {
			fmt.Println("Database not initialized. Run 'swims init'.")
			return nil
		}

		rows, err := db.QueryRows(`
			SELECT
				(SELECT COUNT(*) FROM swimmers) AS swimmers,
				(SELECT COUNT(*) FROM times) AS times,
				(SELECT COUNT(*) FROM events) AS events,
				(SELECT COUNT(*) FROM meets) AS meets
		`)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			return nil
		}

		r := rows[0]
		fmt.Printf("Swimmers:  %.0f\n", toFloat(r["swimmers"]))
		fmt.Printf("Times:     %.0f\n", toFloat(r["times"]))
		fmt.Printf("Events:    %.0f\n", toFloat(r["events"]))
		fmt.Printf("Meets:     %.0f\n", toFloat(r["meets"]))
		return nil
	},
}

func toFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
