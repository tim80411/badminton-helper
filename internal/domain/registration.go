package domain

import "time"

// Registration 代表一筆報名記錄
type Registration struct {
	Timestamp        time.Time
	GroupID          string
	ActivityID       string
	UserID           string
	UserName         string
	RegistrationType string // "季繳" or "臨打"
	Message          string
	CourtAssignment  string
	IsCancelled      bool
}

// RegistrationType 常數
const (
	RegistrationTypeSeason = "季繳"
	RegistrationTypeCasual = "臨打"
)
