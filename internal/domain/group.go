package domain

import "time"

// Group 代表一個 LINE 群組
type Group struct {
	GroupID     string
	GroupName   string
	PictureURL  string
	JoinDate    time.Time
	MemberCount int
}
