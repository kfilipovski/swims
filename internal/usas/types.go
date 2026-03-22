package usas

import (
	"encoding/json"
	"fmt"
)

type JaqlResponse struct {
	Values [][]JaqlCell `json:"values"`
}

type JaqlCell struct {
	Text string
	Data json.RawMessage
}

func (c *JaqlCell) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if t, ok := raw["text"]; ok {
		var s string
		if err := json.Unmarshal(t, &s); err == nil {
			c.Text = s
		}
	}
	if d, ok := raw["data"]; ok {
		c.Data = d
		// If text is empty, try to use data as text
		if c.Text == "" {
			var s string
			if err := json.Unmarshal(d, &s); err == nil {
				c.Text = s
			} else {
				var f float64
				if err := json.Unmarshal(d, &f); err == nil {
					c.Text = fmt.Sprintf("%v", f)
				}
			}
		}
	}
	return nil
}

type JaqlField struct {
	Title    string            `json:"title"`
	Dim      string            `json:"dim"`
	Datatype string            `json:"datatype"`
	Sort     string            `json:"sort,omitempty"`
	Level    string            `json:"level,omitempty"`
	Filter   map[string]interface{} `json:"filter,omitempty"`
}

type JaqlMetadata struct {
	Jaql   JaqlField          `json:"jaql"`
	Panel  string             `json:"panel,omitempty"`
	Format map[string]interface{} `json:"format,omitempty"`
}

type JaqlRequest struct {
	Metadata   []JaqlMetadata `json:"metadata"`
	Datasource string         `json:"datasource"`
	By         string         `json:"by"`
	Count      int            `json:"count"`
}
