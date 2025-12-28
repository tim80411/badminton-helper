package util

import (
	"time"
)

var (
	// TaipeiTZ 台北時區
	TaipeiTZ = time.FixedZone("Asia/Taipei", 8*60*60)
)

// GetCurrentWeekRange 計算本週範圍（週一 00:00 ~ 週日 23:59）
func GetCurrentWeekRange(tz *time.Location) (weekStart, weekEnd time.Time) {
	if tz == nil {
		tz = TaipeiTZ
	}

	now := time.Now().In(tz)
	currentDay := now.Weekday()

	// 計算本週一的偏移量
	var daysToMonday int
	if currentDay == time.Sunday {
		daysToMonday = -6 // 週日的話，本週一是6天前
	} else {
		daysToMonday = -int(currentDay) + 1 // 週一是1，週二是2...週六是6
	}

	// 設定本週一（週的開始）00:00:00
	weekStart = time.Date(
		now.Year(),
		now.Month(),
		now.Day()+daysToMonday,
		0, 0, 0, 0,
		tz,
	)

	// 設定本週日（週的結束）23:59:59
	weekEnd = time.Date(
		weekStart.Year(),
		weekStart.Month(),
		weekStart.Day()+6,
		23, 59, 59, 999999999,
		tz,
	)

	return weekStart, weekEnd
}

// IsSaturday 判斷是否為週六
func IsSaturday(t time.Time, tz *time.Location) bool {
	if tz == nil {
		tz = TaipeiTZ
	}
	return t.In(tz).Weekday() == time.Saturday
}

// NowInTaipei 取得台北時區的當前時間
func NowInTaipei() time.Time {
	return time.Now().In(TaipeiTZ)
}
