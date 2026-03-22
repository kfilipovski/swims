package store

import (
	"fmt"
	"strings"

	"github.com/kfilipovski/swims/internal/dolt"
	"github.com/kfilipovski/swims/internal/model"
)

type TimeStore struct {
	DB *dolt.Dolt
}

func (s *TimeStore) UpsertTimes(swimmerID int64, times []model.Time) error {
	if len(times) == 0 {
		return nil
	}

	// Ensure all referenced events and meets exist first
	if err := s.ensureEvents(times); err != nil {
		return fmt.Errorf("upserting events: %w", err)
	}
	if err := s.ensureMeets(times); err != nil {
		return fmt.Errorf("upserting meets: %w", err)
	}

	// Batch insert times in groups of 50
	for i := 0; i < len(times); i += 50 {
		end := i + 50
		if end > len(times) {
			end = len(times)
		}
		if err := s.upsertBatch(times[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func (s *TimeStore) ensureEvents(times []model.Time) error {
	seen := map[string]bool{}
	var sb strings.Builder
	for _, t := range times {
		if seen[t.EventCode] {
			continue
		}
		seen[t.EventCode] = true
		sb.WriteString(fmt.Sprintf(
			"INSERT IGNORE INTO events (event_code, distance, stroke, course) VALUES (%s, %d, %s, %s);\n",
			sqlStr(t.EventCode), t.Distance, sqlStr(t.Stroke), sqlStr(t.Course),
		))
	}
	return s.DB.SQLBatch(sb.String())
}

func (s *TimeStore) ensureMeets(times []model.Time) error {
	seen := map[string]bool{}
	var sb strings.Builder
	for _, t := range times {
		if seen[t.MeetName] || t.MeetName == "" {
			continue
		}
		seen[t.MeetName] = true
		sb.WriteString(fmt.Sprintf(
			"INSERT IGNORE INTO meets (meet_name) VALUES (%s);\n",
			sqlStr(t.MeetName),
		))
	}
	return s.DB.SQLBatch(sb.String())
}

func (s *TimeStore) upsertBatch(times []model.Time) error {
	var sb strings.Builder
	for _, t := range times {
		dateVal := "NULL"
		if t.SwimDate != "" {
			dateVal = sqlStr(t.SwimDate)
		}
		sb.WriteString(fmt.Sprintf(
			"REPLACE INTO times (swimmer_id, event_id, meet_id, swim_time, sort_key, age_at_meet, power_points, time_standard, lsc_code, team_name, swim_date, synced_at) VALUES (%d, (SELECT id FROM events WHERE event_code = %s), (SELECT id FROM meets WHERE meet_name = %s), %s, %s, %d, %.2f, %s, %s, %s, %s, NOW());\n",
			t.SwimmerID,
			sqlStr(t.EventCode),
			sqlStr(t.MeetName),
			sqlStr(t.SwimTime),
			sqlStr(t.SortKey),
			t.AgeAtMeet,
			t.PowerPoints,
			sqlStr(t.TimeStandard),
			sqlStr(t.LscCode),
			sqlStr(t.TeamName),
			dateVal,
		))
	}
	return s.DB.SQLBatch(sb.String())
}

type TimesFilter struct {
	SwimmerIDs []int64
	Event      string // "200 BR" (distance + stroke, no course)
	Course     string
	Since      string
	Year       int // calendar year (Jan 1 - Dec 31)
	Season     int // competition year (Sep 1 prior year - Aug 31)
	Age        int
	Meet       string
	Sort       string // "date" (default, newest first) or "time" (fastest first)
	Best       bool
}

func (s *TimeStore) QueryTimes(f TimesFilter) ([]model.Time, error) {
	query := buildTimesQuery(f)
	return s.queryToModels(query)
}

func swimmerIDsWhere(alias string, ids []int64) string {
	if len(ids) == 1 {
		return fmt.Sprintf("%s.swimmer_id = %d", alias, ids[0])
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return fmt.Sprintf("%s.swimmer_id IN (%s)", alias, strings.Join(parts, ","))
}

func buildTimesQuery(f TimesFilter) string {
	cols := "e.event_code, t.swim_time, t.sort_key, t.age_at_meet, t.power_points, t.time_standard, m.meet_name, t.lsc_code, t.team_name, t.swim_date, t.swimmer_id, e.distance, e.stroke, e.course"
	from := "times t JOIN events e ON t.event_id = e.id JOIN meets m ON t.meet_id = m.id"
	swimmerWhere := swimmerIDsWhere("t", f.SwimmerIDs)

	eventWhere := buildEventWhere(f.Event, f.Course)

	if f.Best {
		bestSub := fmt.Sprintf("SELECT t0.event_id, t0.swimmer_id, MIN(t0.sort_key) AS min_sk FROM times t0 JOIN events e0 ON t0.event_id = e0.id WHERE %s AND e0.course = %s", swimmerIDsWhere("t0", f.SwimmerIDs), sqlStr(f.Course))
		if f.Event != "" {
			dist, stroke := parseEvent(f.Event)
			bestSub += fmt.Sprintf(" AND e0.distance = %d AND e0.stroke = %s", dist, sqlStr(stroke))
		}
		if f.Age > 0 {
			bestSub += fmt.Sprintf(" AND t0.age_at_meet = %d", f.Age)
		}
		if f.Meet != "" {
			bestSub += fmt.Sprintf(" AND t0.meet_id IN (SELECT id FROM meets WHERE LOWER(meet_name) LIKE LOWER(%s))", sqlStr("%"+f.Meet+"%"))
		}
		bestSub += buildDateWhere(f, "t0")
		bestSub += " GROUP BY t0.event_id, t0.swimmer_id"
		query := fmt.Sprintf("SELECT %s FROM %s JOIN (%s) best ON t.event_id = best.event_id AND t.swimmer_id = best.swimmer_id AND t.sort_key = best.min_sk WHERE %s", cols, from, bestSub, swimmerWhere)
		query += eventWhere
		if f.Age > 0 {
			query += fmt.Sprintf(" AND t.age_at_meet = %d", f.Age)
		}
		if f.Meet != "" {
			query += fmt.Sprintf(" AND LOWER(m.meet_name) LIKE LOWER(%s)", sqlStr("%"+f.Meet+"%"))
		}
		query += buildDateWhere(f, "t")
		query += " ORDER BY " + buildOrderBy(f.Sort)
		return query
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", cols, from, swimmerWhere)
	query += eventWhere
	if f.Age > 0 {
		query += fmt.Sprintf(" AND t.age_at_meet = %d", f.Age)
	}
	if f.Meet != "" {
		query += fmt.Sprintf(" AND LOWER(m.meet_name) LIKE LOWER(%s)", sqlStr("%"+f.Meet+"%"))
	}
	query += buildDateWhere(f, "t")
	query += " ORDER BY " + buildOrderBy(f.Sort)
	return query
}

func buildEventWhere(event, course string) string {
	var w string
	if event != "" {
		dist, stroke := parseEvent(event)
		w += fmt.Sprintf(" AND e.distance = %d AND e.stroke = %s", dist, sqlStr(stroke))
	}
	if course != "" {
		w += fmt.Sprintf(" AND e.course = %s", sqlStr(course))
	}
	return w
}

func parseEvent(event string) (int, string) {
	parts := strings.Fields(event)
	if len(parts) >= 2 {
		dist := 0
		fmt.Sscanf(parts[0], "%d", &dist)
		return dist, parts[1]
	}
	return 0, event
}

func buildOrderBy(sort string) string {
	switch sort {
	case "time":
		return "t.sort_key ASC"
	case "points":
		return "t.power_points DESC, e.event_code"
	default:
		return "t.swim_date DESC, e.event_code"
	}
}

func buildDateWhere(f TimesFilter, alias string) string {
	col := alias + ".swim_date"
	if f.Season > 0 {
		from := fmt.Sprintf("%d-09-01", f.Season-1)
		to := fmt.Sprintf("%d-08-31", f.Season)
		return fmt.Sprintf(" AND %s >= %s AND %s <= %s", col, sqlStr(from), col, sqlStr(to))
	}
	if f.Year > 0 {
		from := fmt.Sprintf("%d-01-01", f.Year)
		to := fmt.Sprintf("%d-12-31", f.Year)
		return fmt.Sprintf(" AND %s >= %s AND %s <= %s", col, sqlStr(from), col, sqlStr(to))
	}
	if f.Since != "" {
		return fmt.Sprintf(" AND %s >= %s", col, sqlStr(f.Since))
	}
	return ""
}

func (s *TimeStore) queryToModels(query string) ([]model.Time, error) {
	rows, err := s.DB.QueryRows(query)
	if err != nil {
		return nil, err
	}

	var times []model.Time
	for _, row := range rows {
		times = append(times, model.Time{
			SwimmerID:    int64(jsonNum(row["swimmer_id"])),
			EventCode:    jsonStr(row["event_code"]),
			Distance:     int(jsonNum(row["distance"])),
			Stroke:       jsonStr(row["stroke"]),
			Course:       jsonStr(row["course"]),
			SwimTime:     jsonStr(row["swim_time"]),
			SortKey:      jsonStr(row["sort_key"]),
			AgeAtMeet:    int(jsonNum(row["age_at_meet"])),
			PowerPoints:  jsonNum(row["power_points"]),
			TimeStandard: jsonStr(row["time_standard"]),
			MeetName:     jsonStr(row["meet_name"]),
			LscCode:      jsonStr(row["lsc_code"]),
			TeamName:     jsonStr(row["team_name"]),
			SwimDate:     jsonStr(row["swim_date"]),
		})
	}
	return times, nil
}
