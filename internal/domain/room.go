package domain

import "time"

// Room 代表一個 LINE 聊天室
type Room struct {
	RoomID   string
	JoinDate time.Time
}
