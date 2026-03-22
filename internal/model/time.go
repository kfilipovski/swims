package model

// Time represents a swim time as returned from queries (joined with event/meet).
type Time struct {
	SwimmerID    int64
	EventCode    string
	Distance     int
	Stroke       string
	Course       string
	SwimTime     string
	SortKey      string
	AgeAtMeet    int
	PowerPoints  float64
	TimeStandard string
	MeetName     string
	LscCode      string
	TeamName     string
	SwimDate     string // YYYY-MM-DD
}
