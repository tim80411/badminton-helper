package service

import (
	"context"
	"fmt"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/tim80411/badminton-helper/internal/lineapi"
	"github.com/tim80411/badminton-helper/internal/repository"
)

// CancellationService 取消報名服務
type CancellationService struct {
	registrationRepo repository.RegistrationRepository
	lineClient       *lineapi.Client
}

// NewCancellationService 建立新的取消報名服務
func NewCancellationService(
	registrationRepo repository.RegistrationRepository,
	lineClient *lineapi.Client,
) *CancellationService {
	return &CancellationService{
		registrationRepo: registrationRepo,
		lineClient:       lineClient,
	}
}

// HandleCancellation 處理取消報名請求
func (s *CancellationService) HandleCancellation(
	ctx context.Context,
	groupID, registrationID, userID, replyToken string,
) error {
	// 驗證權限
	valid, needUserName, errorMsg := s.validateCancellationPermission(ctx, registrationID, userID)

	if !valid {
		if needUserName {
			// 需要取得用戶名稱來顯示錯誤訊息
			profile, err := s.lineClient.GetGroupMemberProfile(ctx, groupID, userID)
			if err != nil {
				errorMsg = "你壞壞，不要取消別人的預約👀"
			} else {
				errorMsg = fmt.Sprintf("%s 壞壞，不要取消別人的預約👀", profile.DisplayName)
			}
		}

		// 回覆錯誤訊息
		textMsg := &messaging_api.TextMessage{
			Text: errorMsg,
		}
		return s.lineClient.ReplyMessage(ctx, replyToken, []messaging_api.MessageInterface{textMsg})
	}

	// 執行取消操作
	err := s.registrationRepo.MarkAsCancelled(ctx, registrationID)
	if err != nil {
		textMsg := &messaging_api.TextMessage{
			Text: "取消報名時發生錯誤，請稍後再試。",
		}
		return s.lineClient.ReplyMessage(ctx, replyToken, []messaging_api.MessageInterface{textMsg})
	}

	// 取得報名記錄以顯示用戶名稱
	registration, _ := s.registrationRepo.FindByID(ctx, registrationID)
	userName := "用戶"
	registrationType := "報名"
	if registration != nil {
		userName = registration.UserName
		registrationType = registration.RegistrationType
	}

	// 回覆成功訊息
	successMsg := &messaging_api.TextMessage{
		Text: fmt.Sprintf("✅ %s 的 %s 報名已成功取消！", userName, registrationType),
	}
	return s.lineClient.ReplyMessage(ctx, replyToken, []messaging_api.MessageInterface{successMsg})
}

// validateCancellationPermission 驗證取消報名權限
// 返回值: (是否有效, 是否需要用戶名稱, 錯誤訊息)
func (s *CancellationService) validateCancellationPermission(
	ctx context.Context,
	registrationID, userID string,
) (bool, bool, string) {
	// 查詢報名記錄
	registration, err := s.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		return false, false, "找不到對應的報名記錄。是不是管理員在壞壞！"
	}

	// 檢查是否為同一用戶
	if registration.UserID != userID {
		return false, true, "" // 需要取得用戶名稱
	}

	// 檢查是否已經取消過
	if registration.IsCancelled {
		return false, false, "這個報名已經被取消過囉，請大人放過我的伺服器"
	}

	return true, false, ""
}
