package model

type Swimmer struct {
	SwimmerID      int64
	FullName       string
	ClubName       string
	LscCode        string
	Age            int
	TimesSyncedAt  string // YYYY-MM-DD or empty
}
