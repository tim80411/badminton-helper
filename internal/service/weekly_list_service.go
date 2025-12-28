package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/tim80411/badminton-helper/internal/domain"
	"github.com/tim80411/badminton-helper/internal/lineapi"
	"github.com/tim80411/badminton-helper/internal/repository"
	"github.com/tim80411/badminton-helper/internal/util"
)

// WeeklyListService 週名單服務
type WeeklyListService struct {
	registrationRepo repository.RegistrationRepository
	lineClient       *lineapi.Client
}

// NewWeeklyListService 建立新的週名單服務
func NewWeeklyListService(
	registrationRepo repository.RegistrationRepository,
	lineClient *lineapi.Client,
) *WeeklyListService {
	return &WeeklyListService{
		registrationRepo: registrationRepo,
		lineClient:       lineClient,
	}
}

// HandleWeeklyListRequest 處理週名單請求
func (s *WeeklyListService) HandleWeeklyListRequest(
	ctx context.Context,
	sourceID, activityID, replyToken string,
) error {
	// 取得本週報名資料
	weekStart, weekEnd := util.GetCurrentWeekRange(util.TaipeiTZ)
	registrations, err := s.registrationRepo.GetWeeklyRegistrations(ctx, sourceID, activityID, weekStart, weekEnd)
	if err != nil {
		errorMsg := &messaging_api.TextMessage{
			Text: "取得本週名單時發生錯誤，請稍後再試。",
		}
		return s.lineClient.ReplyMessage(ctx, replyToken, []messaging_api.MessageInterface{errorMsg})
	}

	// 分類季繳和臨打，並按時間排序
	seasonTicket, casualPlay := s.categorizeRegistrations(registrations)

	// 判斷是否為週六
	isSaturday := util.IsSaturday(util.NowInTaipei(), util.TaipeiTZ)

	// 建立 Flex Message
	weeklyListMsg := lineapi.BuildWeeklyListFlexMessage(seasonTicket, casualPlay, isSaturday)

	// 回覆訊息
	err = s.lineClient.ReplyMessage(ctx, replyToken, []messaging_api.MessageInterface{weeklyListMsg})
	if err != nil {
		return fmt.Errorf("回覆週名單訊息失敗: %w", err)
	}

	return nil
}

// categorizeRegistrations 分類並排序報名記錄
func (s *WeeklyListService) categorizeRegistrations(registrations []*domain.Registration) (
	seasonTicket, casualPlay []*domain.Registration,
) {
	for _, reg := range registrations {
		if reg.RegistrationType == domain.RegistrationTypeSeason {
			seasonTicket = append(seasonTicket, reg)
		} else if reg.RegistrationType == domain.RegistrationTypeCasual {
			casualPlay = append(casualPlay, reg)
		}
	}

	// 按報名時間排序（先報名的在前面）
	sort.Slice(seasonTicket, func(i, j int) bool {
		return seasonTicket[i].Timestamp.Before(seasonTicket[j].Timestamp)
	})

	sort.Slice(casualPlay, func(i, j int) bool {
		return casualPlay[i].Timestamp.Before(casualPlay[j].Timestamp)
	})

	return seasonTicket, casualPlay
}
