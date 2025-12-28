package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/tim80411/badminton-helper/internal/domain"
)

const (
	RegistrationsSheet = "Registrations"
)

// RegistrationRepository 定義報名記錄的資料操作介面
type RegistrationRepository interface {
	Create(ctx context.Context, reg *domain.Registration) error
	GetWeeklyRegistrations(ctx context.Context, sourceID, activityID string, weekStart, weekEnd time.Time) ([]*domain.Registration, error)
	FindByID(ctx context.Context, registrationID string) (*domain.Registration, error)
	MarkAsCancelled(ctx context.Context, registrationID string) error
}

type registrationRepo struct {
	client *SheetsClient
}

// NewRegistrationRepository 建立新的報名記錄 repository
func NewRegistrationRepository(client *SheetsClient) RegistrationRepository {
	return &registrationRepo{
		client: client,
	}
}

// Create 新增報名記錄
func (r *registrationRepo) Create(ctx context.Context, reg *domain.Registration) error {
	// 確保 sheet 存在
	headers := []interface{}{
		"Timestamp", "Group ID", "ActivityId", "User ID", "User Name",
		"Registration Type", "Message", "Court Assignment", "isCancelled",
	}
	err := r.client.EnsureSheetExists(ctx, RegistrationsSheet, headers)
	if err != nil {
		return fmt.Errorf("確保 Registrations sheet 存在失敗: %w", err)
	}

	// 新增資料
	values := []interface{}{
		reg.Timestamp,
		reg.GroupID,
		reg.ActivityID,
		reg.UserID,
		reg.UserName,
		reg.RegistrationType,
		reg.Message,
		reg.CourtAssignment,
		reg.IsCancelled,
	}

	err = r.client.AppendRow(ctx, RegistrationsSheet, values)
	if err != nil {
		return fmt.Errorf("新增報名記錄失敗: %w", err)
	}

	return nil
}

// GetWeeklyRegistrations 取得本週的報名記錄
func (r *registrationRepo) GetWeeklyRegistrations(ctx context.Context, sourceID, activityID string, weekStart, weekEnd time.Time) ([]*domain.Registration, error) {
	// 讀取所有資料
	readRange := fmt.Sprintf("%s!A:I", RegistrationsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return nil, fmt.Errorf("讀取報名記錄失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return []*domain.Registration{}, nil // 只有表頭或沒有資料
	}

	var registrations []*domain.Registration

	// 跳過表頭，從第二行開始處理
	for _, row := range valueRange.Values[1:] {
		if len(row) < 6 {
			continue // 資料不完整，跳過
		}

		// 解析時間
		timestamp, err := parseTimestamp(row[0])
		if err != nil {
			continue // 時間格式錯誤，跳過
		}

		// 檢查是否在本週範圍內
		if timestamp.Before(weekStart) || timestamp.After(weekEnd) {
			continue
		}

		// 檢查 Group ID 和 Activity ID 是否匹配
		groupID := getStringValue(row, 1)
		actID := getStringValue(row, 2)
		if groupID != sourceID || actID != activityID {
			continue
		}

		// 檢查是否已取消
		isCancelled := getBoolValue(row, 8)
		if isCancelled {
			continue // 已取消，跳過
		}

		registration := &domain.Registration{
			Timestamp:        timestamp,
			GroupID:          groupID,
			ActivityID:       actID,
			UserID:           getStringValue(row, 3),
			UserName:         getStringValue(row, 4),
			RegistrationType: getStringValue(row, 5),
			Message:          getStringValue(row, 6),
			CourtAssignment:  getStringValue(row, 7),
			IsCancelled:      isCancelled,
		}

		registrations = append(registrations, registration)
	}

	return registrations, nil
}

// FindByID 根據 registration ID 查詢報名記錄
// registration ID 格式: {timestamp}-{userId}
func (r *registrationRepo) FindByID(ctx context.Context, registrationID string) (*domain.Registration, error) {
	// 讀取所有資料
	readRange := fmt.Sprintf("%s!A:I", RegistrationsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return nil, fmt.Errorf("讀取報名記錄失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return nil, fmt.Errorf("找不到報名記錄")
	}

	// 跳過表頭，從第二行開始處理
	for _, row := range valueRange.Values[1:] {
		if len(row) < 4 {
			continue
		}

		timestamp, err := parseTimestamp(row[0])
		if err != nil {
			continue
		}

		userID := getStringValue(row, 3)
		currentID := generateRegistrationID(timestamp, userID)

		if currentID == registrationID {
			return &domain.Registration{
				Timestamp:        timestamp,
				GroupID:          getStringValue(row, 1),
				ActivityID:       getStringValue(row, 2),
				UserID:           userID,
				UserName:         getStringValue(row, 4),
				RegistrationType: getStringValue(row, 5),
				Message:          getStringValue(row, 6),
				CourtAssignment:  getStringValue(row, 7),
				IsCancelled:      getBoolValue(row, 8),
			}, nil
		}
	}

	return nil, fmt.Errorf("找不到 ID 為 %s 的報名記錄", registrationID)
}

// MarkAsCancelled 標記報名記錄為已取消
func (r *registrationRepo) MarkAsCancelled(ctx context.Context, registrationID string) error {
	// 讀取所有資料
	readRange := fmt.Sprintf("%s!A:I", RegistrationsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return fmt.Errorf("讀取報名記錄失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return fmt.Errorf("找不到報名記錄")
	}

	// 找到對應的行並更新
	for i, row := range valueRange.Values[1:] {
		if len(row) < 4 {
			continue
		}

		timestamp, err := parseTimestamp(row[0])
		if err != nil {
			continue
		}

		userID := getStringValue(row, 3)
		currentID := generateRegistrationID(timestamp, userID)

		if currentID == registrationID {
			// 找到了，更新第 9 欄 (isCancelled)
			// i+2 因為：i 是從 values[1:] 開始計數，+1 跳過表頭，+1 因為 Sheets 從 1 開始
			rowNum := i + 2
			cellRange := fmt.Sprintf("%s!I%d", RegistrationsSheet, rowNum)
			err = r.client.UpdateCell(ctx, cellRange, true)
			if err != nil {
				return fmt.Errorf("更新取消狀態失敗: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("找不到 ID 為 %s 的報名記錄", registrationID)
}

// 輔助函數：解析時間戳
func parseTimestamp(value interface{}) (time.Time, error) {
	str, ok := value.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("時間戳格式錯誤")
	}

	// 嘗試多種時間格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, str)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("無法解析時間: %s", str)
}

// 輔助函數：從 row 中取得字串值
func getStringValue(row []interface{}, index int) string {
	if index >= len(row) {
		return ""
	}
	if row[index] == nil {
		return ""
	}
	str, ok := row[index].(string)
	if !ok {
		return fmt.Sprintf("%v", row[index])
	}
	return str
}

// 輔助函數：從 row 中取得布林值
func getBoolValue(row []interface{}, index int) bool {
	if index >= len(row) {
		return false
	}
	if row[index] == nil {
		return false
	}

	// 嘗試直接轉換為 bool
	if b, ok := row[index].(bool); ok {
		return b
	}

	// 嘗試從字串轉換
	str := getStringValue(row, index)
	b, err := strconv.ParseBool(str)
	if err != nil {
		// "TRUE" 字串
		return str == "TRUE" || str == "true" || str == "True"
	}
	return b
}

// 輔助函數：生成 registration ID
func generateRegistrationID(timestamp time.Time, userID string) string {
	return fmt.Sprintf("%d-%s", timestamp.UnixMilli(), userID)
}
