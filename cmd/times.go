package cmd

import (
	"fmt"
	"strings"

	"github.com/kfilipovski/swims/internal/format"
	"github.com/kfilipovski/swims/internal/model"
	"github.com/kfilipovski/swims/internal/store"
	"github.com/spf13/cobra"
)

var timesCmd = &cobra.Command{
	Use:   "times [swimmer...]",
	Short: "Query swim times from local database",
	Long:  "Query swim times. Arguments are swimmer names (case-insensitive match on tracked swimmers). If no swimmers specified, shows times for all tracked swimmers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		event, _ := cmd.Flags().GetString("event")
		course, _ := cmd.Flags().GetString("course")
		since, _ := cmd.Flags().GetString("since")
		year, _ := cmd.Flags().GetInt("year")
		season, _ := cmd.Flags().GetInt("season")
		age, _ := cmd.Flags().GetInt("age")
		meet, _ := cmd.Flags().GetString("meet")
		sortFlag, _ := cmd.Flags().GetString("sort")
		best, _ := cmd.Flags().GetBool("best")
		graph, _ := cmd.Flags().GetString("graph")

		ss := &store.SwimmerStore{DB: db}
		swimmers, err := resolveSwimmers(ss, args)
		if err != nil {
			return err
		}
		if len(swimmers) == 0 {
			fmt.Println("No matching swimmers found.")
			return nil
		}

		// Build swimmer ID list and name lookup
		ids := make([]int64, len(swimmers))
		nameMap := make(map[int64]string)
		for i, s := range swimmers {
			ids[i] = s.SwimmerID
			nameMap[s.SwimmerID] = s.FullName
		}

		ts := &store.TimeStore{DB: db}
		times, err := ts.QueryTimes(store.TimesFilter{
			SwimmerIDs: ids,
			Event:      event,
			Course:     course,
			Since:      since,
			Year:       year,
			Season:     season,
			Age:        age,
			Meet:       meet,
			Sort:       sortFlag,
			Best:       best,
		})
		if err != nil {
			return err
		}

		if len(times) == 0 {
			fmt.Println("No times found.")
			return nil
		}

		if graph != "" {
			printGraph(times, nameMap, graph)
		} else {
			printTimes(times, nameMap)
		}
		return nil
	},
}

func resolveSwimmers(ss *store.SwimmerStore, args []string) ([]model.Swimmer, error) {
	if len(args) == 0 {
		return ss.List("")
	}

	seen := map[int64]bool{}
	var result []model.Swimmer
	for _, name := range args {
		matches, err := ss.SearchByName(name)
		if err != nil {
			return nil, fmt.Errorf("searching for %q: %w", name, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no tracked swimmer matching %q", name)
		}
		for _, s := range matches {
			if !seen[s.SwimmerID] {
				seen[s.SwimmerID] = true
				result = append(result, s)
			}
		}
	}
	return result, nil
}

func printTimes(times []model.Time, nameMap map[int64]string) {
	multiSwimmer := len(nameMap) > 1

	var headers []string
	if multiSwimmer {
		headers = []string{"Swimmer", "Event", "Course", "Time", "Age", "Points", "Standard", "Date", "Team", "Meet"}
	} else {
		headers = []string{"Event", "Course", "Time", "Age", "Points", "Standard", "Date", "Team", "Meet"}
	}

	tbl := format.NewTable(headers...)
	for _, t := range times {
		event := fmt.Sprintf("%d %s", t.Distance, t.Stroke)
		row := []string{
			event,
			t.Course,
			t.SwimTime,
			fmt.Sprintf("%d", t.AgeAtMeet),
			fmt.Sprintf("%.0f", t.PowerPoints),
			t.TimeStandard,
			t.SwimDate,
			t.TeamName,
			t.MeetName,
		}
		if multiSwimmer {
			name := nameMap[t.SwimmerID]
			// Shorten to first + last initial for compact display
			row = append([]string{shortenName(name)}, row...)
		}
		tbl.Row(row...)
	}
	tbl.Flush()
}

func printGraph(times []model.Time, nameMap map[int64]string, mode string) {
	type key struct {
		swimmerID int64
		event     string
	}
	groups := make(map[key][]model.Time)
	var order []key
	for _, t := range times {
		k := key{t.SwimmerID, fmt.Sprintf("%d %s", t.Distance, t.Stroke)}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], t)
	}

	invertY := mode == "time" // lower time = better = bottom of chart

	for _, k := range order {
		g := groups[k]
		// Build points in chronological order
		var points []format.Point
		for i := len(g) - 1; i >= 0; i-- {
			t := g[i]
			var val float64
			if mode == "points" {
				val = t.PowerPoints
			} else {
				val = format.TimeToSeconds(t.SwimTime)
			}
			if val > 0 {
				points = append(points, format.Point{
					Label: t.SwimDate,
					Value: val,
				})
			}
		}
		if len(points) == 0 {
			continue
		}

		label := k.event + " " + g[0].Course
		if len(nameMap) > 1 {
			label = shortenName(nameMap[k.swimmerID]) + " - " + label
		}
		fmt.Printf("\n%s (%d swims)\n", label, len(points))
		fmt.Print(format.Graph(points, 80, 15, invertY))
	}
}

func shortenName(full string) string {
	parts := strings.Fields(full)
	if len(parts) <= 2 {
		return full
	}
	// First name + last name (drop middle)
	return parts[0] + " " + parts[len(parts)-1]
}

func init() {
	timesCmd.Flags().String("event", "", "filter by event, e.g. '200 BR'")
	timesCmd.Flags().String("course", "SCY", "course (SCY, SCM, LCM)")
	timesCmd.Flags().String("since", "", "filter by date (YYYY-MM-DD)")
	timesCmd.Flags().Int("year", 0, "filter by calendar year (Jan-Dec)")
	timesCmd.Flags().Int("season", 0, "filter by competition year (Sep 1 prior year - Aug 31)")
	timesCmd.Flags().Int("age", 0, "filter by age at meet")
	timesCmd.Flags().String("meet", "", "filter by meet name (contains match)")
	timesCmd.Flags().String("sort", "date", "sort order: date (newest first), time (fastest first), or points (highest first)")
	timesCmd.Flags().Bool("best", false, "show only best time per event")
	timesCmd.Flags().String("graph", "", "show ASCII progression graph: time or points")
	rootCmd.AddCommand(timesCmd)
}
