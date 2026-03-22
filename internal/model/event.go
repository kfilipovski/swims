package model

import (
	"fmt"
	"strconv"
	"strings"
)

type Event struct {
	ID        int64
	EventCode string
	Distance  int
	Stroke    string
	Course    string
}

// ParseEventCode splits "200 BR SCY" into (200, "BR", "SCY").
func ParseEventCode(code string) (distance int, stroke string, course string, err error) {
	parts := strings.Fields(code)
	if len(parts) != 3 {
		return 0, "", "", fmt.Errorf("invalid event code %q: expected 3 parts", code)
	}
	distance, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid distance in event code %q: %w", code, err)
	}
	return distance, parts[1], parts[2], nil
}
