package usas

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kfilipovski/swims/internal/model"
)

func (c *Client) SearchSwimmers(firstName, lastName string) ([]model.Swimmer, error) {
	metadata := []JaqlMetadata{
		{Jaql: JaqlField{Title: "Name", Dim: "[Persons.FullName]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "Club", Dim: "[Persons.ClubName]", Datatype: "text", Sort: "asc"}},
		{Jaql: JaqlField{Title: "LSC", Dim: "[Persons.LscCode]", Datatype: "text"}},
		{Jaql: JaqlField{Title: "Age", Dim: "[Persons.Age]", Datatype: "numeric"}},
		{Jaql: JaqlField{Title: "PersonKey", Dim: "[Persons.PersonKey]", Datatype: "numeric"}},
		{
			Jaql:  JaqlField{Title: "FirstAndPreferredName", Dim: "[Persons.FirstAndPreferredName]", Datatype: "text", Filter: map[string]interface{}{"contains": firstName}},
			Panel: "scope",
		},
		{
			Jaql:  JaqlField{Title: "LastName", Dim: "[Persons.LastName]", Datatype: "text", Filter: map[string]interface{}{"contains": lastName}},
			Panel: "scope",
		},
	}

	resp, err := c.query("Public Person Search", metadata)
	if err != nil {
		return nil, fmt.Errorf("searching swimmers: %w", err)
	}

	var swimmers []model.Swimmer
	for _, row := range resp.Values {
		if len(row) < 5 {
			continue
		}

		swimmerID, _ := parseNumericCell(row[4])
		age, _ := parseNumericCell(row[3])

		swimmers = append(swimmers, model.Swimmer{
			SwimmerID: int64(swimmerID),
			FullName:  row[0].Text,
			ClubName:  row[1].Text,
			LscCode:   row[2].Text,
			Age:       int(age),
		})
	}
	return swimmers, nil
}

func parseNumericCell(cell JaqlCell) (float64, error) {
	if cell.Data != nil {
		var f float64
		if err := json.Unmarshal(cell.Data, &f); err == nil {
			return f, nil
		}
		var s string
		if err := json.Unmarshal(cell.Data, &s); err == nil {
			return strconv.ParseFloat(s, 64)
		}
	}
	return strconv.ParseFloat(cell.Text, 64)
}
