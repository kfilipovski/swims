package usas

import (
	"fmt"
	"strings"
	"time"

	"github.com/kfilipovski/swims/internal/model"
)

func (c *Client) FetchTimes(swimmerID int64, eventCode string, since string) ([]model.Time, error) {
	metadata := []JaqlMetadata{
		{Jaql: JaqlField{Title: "Event", Dim: "[SwimEvent.EventCode]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "SwimTime", Dim: "[UsasSwimTime.SwimTimeFormatted]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "Age", Dim: "[UsasSwimTime.AgeAtMeetKey]", Datatype: "numeric"}},
		{Jaql: JaqlField{Title: "Points", Dim: "[UsasSwimTime.PowerPoints]", Datatype: "numeric"}},
		{Jaql: JaqlField{Title: "TimeStandard", Dim: "[TimeStandard.TimeStandardName]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "Meet", Dim: "[Meet.MeetName]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "LSC", Dim: "[OrgUnit.Level3Code]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "Team", Dim: "[OrgUnit.Level4Name]", Datatype: "text"}},
		{
			Jaql: JaqlField{Title: "SwimDate", Dim: "[SeasonCalendar.CalendarDate (Calendar)]", Datatype: "datetime", Level: "days"},
			Format: map[string]interface{}{
				"mask": map[string]interface{}{
					"days": "MM/dd/yyyy",
				},
			},
		},
		{Jaql: JaqlField{Title: "SortKey", Dim: "[UsasSwimTime.SortKey]", Datatype: "text", Sort: "asc"}},
		{
			Jaql:  JaqlField{Title: "PersonKey", Dim: "[UsasSwimTime.PersonKey]", Datatype: "numeric", Filter: map[string]interface{}{"equals": swimmerID}},
			Panel: "scope",
		},
	}

	if eventCode != "" {
		metadata = append(metadata, JaqlMetadata{
			Jaql:  JaqlField{Title: "EventFilter", Dim: "[SwimEvent.EventCode]", Datatype: "text", Filter: map[string]interface{}{"members": []string{eventCode}}},
			Panel: "scope",
		})
	}

	if since != "" {
		metadata = append(metadata, JaqlMetadata{
			Jaql: JaqlField{
				Title:    "DateFilter",
				Dim:      "[SeasonCalendar.CalendarDate (Calendar)]",
				Datatype: "datetime",
				Level:    "days",
				Filter: map[string]interface{}{
					"from": since,
				},
			},
			Panel: "scope",
		})
	}

	resp, err := c.query("USA Swimming Times Elasticube", metadata)
	if err != nil {
		return nil, fmt.Errorf("fetching times: %w", err)
	}

	var times []model.Time
	for _, row := range resp.Values {
		if len(row) < 10 {
			continue
		}

		ec := row[0].Text
		distance, stroke, course, err := model.ParseEventCode(ec)
		if err != nil {
			continue
		}

		ageAtMeet, _ := parseNumericCell(row[2])
		points, _ := parseNumericCell(row[3])

		swimDate := parseDate(row[8].Text)

		times = append(times, model.Time{
			SwimmerID:    swimmerID,
			EventCode:    ec,
			Distance:     distance,
			Stroke:       stroke,
			Course:       course,
			SwimTime:     row[1].Text,
			SortKey:      row[9].Text,
			AgeAtMeet:    int(ageAtMeet),
			PowerPoints:  points,
			TimeStandard: cleanTimeStandard(row[4].Text),
			MeetName:     row[5].Text,
			LscCode:      row[6].Text,
			TeamName:     row[7].Text,
			SwimDate:     swimDate,
		})
	}
	return times, nil
}

func cleanTimeStandard(s string) string {
	// Strip year-range prefix like "2020-2024 " from "2020-2024 AA"
	if len(s) > 10 && s[4] == '-' && s[9] == ' ' {
		return s[10:]
	}
	return s
}

func parseDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		// Try ISO format
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return s
		}
	}
	return t.Format("2006-01-02")
}
