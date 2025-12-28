package util

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateRegistrationID 生成報名記錄唯一 ID
// 格式: {timestamp_millis}-{userId}
func GenerateRegistrationID(timestamp time.Time, userID string) string {
	return fmt.Sprintf("%d-%s", timestamp.UnixMilli(), userID)
}

// GenerateActivityID 生成活動 ID
// 格式: YYYYMMDD-{uuid}
func GenerateActivityID() string {
	now := NowInTaipei()
	return fmt.Sprintf("%s-%s", now.Format("20060102"), uuid.New().String())
}
