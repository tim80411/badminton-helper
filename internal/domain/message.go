package domain

import "time"

// Message 代表一則訊息記錄
type Message struct {
	Timestamp time.Time
	SourceID  string
	UserID    string
	Content   string
}
