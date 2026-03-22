package store

import (
	"fmt"
	"strings"

	"github.com/kfilipovski/swims/internal/dolt"
	"github.com/kfilipovski/swims/internal/model"
)

type SwimmerStore struct {
	DB *dolt.Dolt
}

func (s *SwimmerStore) Upsert(swimmers []model.Swimmer) error {
	if len(swimmers) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, p := range swimmers {
		sb.WriteString(fmt.Sprintf(
			"REPLACE INTO swimmers (swimmer_id, full_name, club_name, lsc_code, age, synced_at) VALUES (%d, %s, %s, %s, %d, NOW());\n",
			p.SwimmerID,
			sqlStr(p.FullName),
			sqlStr(p.ClubName),
			sqlStr(p.LscCode),
			p.Age,
		))
	}
	return s.DB.SQLBatch(sb.String())
}

func (s *SwimmerStore) List(lscFilter string) ([]model.Swimmer, error) {
	query := "SELECT swimmer_id, full_name, club_name, lsc_code, age, times_synced_at FROM swimmers"
	if lscFilter != "" {
		query += fmt.Sprintf(" WHERE lsc_code = %s", sqlStr(lscFilter))
	}
	query += " ORDER BY full_name"

	rows, err := s.DB.QueryRows(query)
	if err != nil {
		return nil, err
	}

	var swimmers []model.Swimmer
	for _, row := range rows {
		swimmers = append(swimmers, model.Swimmer{
			SwimmerID:     int64(jsonNum(row["swimmer_id"])),
			FullName:      jsonStr(row["full_name"]),
			ClubName:      jsonStr(row["club_name"]),
			LscCode:       jsonStr(row["lsc_code"]),
			Age:           int(jsonNum(row["age"])),
			TimesSyncedAt: jsonStr(row["times_synced_at"]),
		})
	}
	return swimmers, nil
}

func (s *SwimmerStore) Get(swimmerID int64) (*model.Swimmer, error) {
	rows, err := s.DB.QueryRows(fmt.Sprintf(
		"SELECT swimmer_id, full_name, club_name, lsc_code, age, times_synced_at FROM swimmers WHERE swimmer_id = %d", swimmerID))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	return &model.Swimmer{
		SwimmerID:     int64(jsonNum(row["swimmer_id"])),
		FullName:      jsonStr(row["full_name"]),
		ClubName:      jsonStr(row["club_name"]),
		LscCode:       jsonStr(row["lsc_code"]),
		Age:           int(jsonNum(row["age"])),
		TimesSyncedAt: jsonStr(row["times_synced_at"]),
	}, nil
}

func (s *SwimmerStore) SearchByName(name string) ([]model.Swimmer, error) {
	query := fmt.Sprintf(
		"SELECT swimmer_id, full_name, club_name, lsc_code, age, times_synced_at FROM swimmers WHERE LOWER(full_name) LIKE LOWER(%s) ORDER BY full_name",
		sqlStr("%"+name+"%"))

	rows, err := s.DB.QueryRows(query)
	if err != nil {
		return nil, err
	}

	var swimmers []model.Swimmer
	for _, row := range rows {
		swimmers = append(swimmers, model.Swimmer{
			SwimmerID:     int64(jsonNum(row["swimmer_id"])),
			FullName:      jsonStr(row["full_name"]),
			ClubName:      jsonStr(row["club_name"]),
			LscCode:       jsonStr(row["lsc_code"]),
			Age:           int(jsonNum(row["age"])),
			TimesSyncedAt: jsonStr(row["times_synced_at"]),
		})
	}
	return swimmers, nil
}

func (s *SwimmerStore) UpdateAfterSync(swimmerID int64, date string) error {
	return s.DB.SQLExec(fmt.Sprintf(
		"UPDATE swimmers SET times_synced_at = %s, age = COALESCE((SELECT MAX(age_at_meet) FROM times WHERE swimmer_id = %d), age) WHERE swimmer_id = %d",
		sqlStr(date), swimmerID, swimmerID))
}

func sqlStr(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	return "'" + s + "'"
}

func jsonStr(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func jsonNum(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}
