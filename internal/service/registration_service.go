package service

import (
	"context"
	"fmt"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/tim80411/badminton-helper/internal/domain"
	"github.com/tim80411/badminton-helper/internal/lineapi"
	"github.com/tim80411/badminton-helper/internal/repository"
	"github.com/tim80411/badminton-helper/internal/util"
)

// RegistrationService 報名服務
type RegistrationService struct {
	registrationRepo repository.RegistrationRepository
	lineClient       *lineapi.Client
}

// NewRegistrationService 建立新的報名服務
func NewRegistrationService(
	registrationRepo repository.RegistrationRepository,
	lineClient *lineapi.Client,
) *RegistrationService {
	return &RegistrationService{
		registrationRepo: registrationRepo,
		lineClient:       lineClient,
	}
}

// HandleRegistration 處理報名請求
func (s *RegistrationService) HandleRegistration(
	ctx context.Context,
	groupID, activityID, message, userID, replyToken string,
	eventTimestamp time.Time,
) error {
	// 取得用戶資料
	profile, err := s.lineClient.GetGroupMemberProfile(ctx, groupID, userID)
	if err != nil {
		return fmt.Errorf("取得用戶資料失敗: %w", err)
	}

	// 判斷報名類型
	var registrationType string
	if contains(message, "臨打") {
		registrationType = domain.RegistrationTypeCasual
	} else if contains(message, "季繳") {
		registrationType = domain.RegistrationTypeSeason
	} else {
		registrationType = "其他"
	}

	// 建立報名記錄
	registration := &domain.Registration{
		Timestamp:        eventTimestamp,
		GroupID:          groupID,
		ActivityID:       activityID,
		UserID:           userID,
		UserName:         profile.DisplayName,
		RegistrationType: registrationType,
		Message:          message,
		CourtAssignment:  "",
		IsCancelled:      false,
	}

	// 儲存到 Google Sheets
	err = s.registrationRepo.Create(ctx, registration)
	if err != nil {
		return fmt.Errorf("儲存報名記錄失敗: %w", err)
	}

	// 生成 registration ID
	registrationID := util.GenerateRegistrationID(eventTimestamp, userID)

	// 建立確認訊息
	confirmMsg := lineapi.BuildRegistrationConfirmMessage(
		profile.DisplayName,
		registrationType,
		registrationID,
	)

	// 回覆訊息
	err = s.lineClient.ReplyMessage(ctx, replyToken, []messaging_api.MessageInterface{confirmMsg})
	if err != nil {
		return fmt.Errorf("回覆確認訊息失敗: %w", err)
	}

	return nil
}

// contains 檢查字串是否包含子字串
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
