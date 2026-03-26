package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kfilipovski/swims/internal/format"
	"github.com/kfilipovski/swims/internal/model"
	"github.com/kfilipovski/swims/internal/store"
	"github.com/kfilipovski/swims/internal/usas"
	"github.com/spf13/cobra"
)

var syncSearchCmd = &cobra.Command{
	Use:   "swimmer",
	Short: "Search USA Swimming for swimmers and save to local DB",
	RunE: func(cmd *cobra.Command, args []string) error {
		first, _ := cmd.Flags().GetString("first")
		last, _ := cmd.Flags().GetString("last")
		club, _ := cmd.Flags().GetString("club")
		saveAll, _ := cmd.Flags().GetBool("all")
		if club == "" && appConfig != nil {
			club = appConfig.DefaultClub
		}

		if first == "" && last == "" {
			return fmt.Errorf("at least one of --first or --last is required")
		}

		if !db.IsInitialized() {
			if err := db.Init(); err != nil {
				return err
			}
			if err := db.EnsureSchema(); err != nil {
				return err
			}
		}

		client := usas.NewClient(cfgManager)
		swimmers, err := client.SearchSwimmers(first, last)
		if err != nil {
			return err
		}

		if len(swimmers) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		// Client-side club filter (API doesn't support club in search)
		if club != "" {
			clubLower := strings.ToLower(club)
			var filtered []model.Swimmer
			for _, s := range swimmers {
				if strings.Contains(strings.ToLower(s.ClubName), clubLower) {
					filtered = append(filtered, s)
				}
			}
			if len(filtered) == 0 {
				fmt.Printf("No results matching club %q.\n", club)
				return nil
			}
			swimmers = filtered
		}

		tbl := format.NewTable("#", "Name", "Age", "LSC", "Club", "SwimmerID")
		for i, s := range swimmers {
			tbl.Row(fmt.Sprintf("%d", i+1), s.FullName, fmt.Sprintf("%d", s.Age), s.LscCode, s.ClubName, fmt.Sprintf("%d", s.SwimmerID))
		}
		tbl.Flush()

		var selected []model.Swimmer
		if saveAll || len(swimmers) == 1 {
			selected = swimmers
		} else {
			selected, err = promptSelection(swimmers)
			if err != nil {
				return err
			}
		}

		if len(selected) == 0 {
			fmt.Println("No swimmers selected.")
			return nil
		}

		ss := &store.SwimmerStore{DB: db}
		if err := ss.Upsert(selected); err != nil {
			return fmt.Errorf("saving swimmers: %w", err)
		}
		if err := db.Add(); err != nil {
			return err
		}
		if err := db.Commit(fmt.Sprintf("sync: search %s %s", first, last)); err != nil {
			return err
		}

		fmt.Printf("\nSaved %d swimmer(s) to database.\n", len(selected))
		return nil
	},
}

func promptSelection(swimmers []model.Swimmer) ([]model.Swimmer, error) {
	fmt.Printf("\nSelect swimmers to save (e.g. 1,3,5 or 1-3 or all): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	input = strings.TrimSpace(input)

	if input == "" {
		return nil, nil
	}
	if strings.EqualFold(input, "all") {
		return swimmers, nil
	}

	indices, err := parseSelection(input, len(swimmers))
	if err != nil {
		return nil, err
	}

	var selected []model.Swimmer
	for _, idx := range indices {
		selected = append(selected, swimmers[idx])
	}
	return selected, nil
}

func parseSelection(input string, max int) ([]int, error) {
	seen := map[int]bool{}
	var indices []int

	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid selection: %q", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid selection: %q", part)
			}
			if start < 1 || end > max || start > end {
				return nil, fmt.Errorf("range %d-%d out of bounds (1-%d)", start, end, max)
			}
			for i := start; i <= end; i++ {
				if !seen[i-1] {
					seen[i-1] = true
					indices = append(indices, i-1)
				}
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid selection: %q", part)
			}
			if n < 1 || n > max {
				return nil, fmt.Errorf("selection %d out of bounds (1-%d)", n, max)
			}
			if !seen[n-1] {
				seen[n-1] = true
				indices = append(indices, n-1)
			}
		}
	}
	return indices, nil
}

func init() {
	syncSearchCmd.Flags().String("first", "", "first/preferred name (contains match)")
	syncSearchCmd.Flags().String("last", "", "last name (contains match)")
	syncSearchCmd.Flags().String("club", "", "filter results by club name (contains match, default: config club)")
	syncSearchCmd.Flags().Bool("all", false, "save all results without prompting")
	syncCmd.AddCommand(syncSearchCmd)
}
