package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"github.com/tim80411/badminton-helper/internal/domain"
	"github.com/tim80411/badminton-helper/internal/repository"
	"github.com/tim80411/badminton-helper/internal/service"
	"go.uber.org/zap"
)

// WebhookHandler 處理 LINE webhook 請求
type WebhookHandler struct {
	channelSecret       string
	groupRepo           repository.GroupRepository
	roomRepo            repository.RoomRepository
	messageRepo         repository.MessageRepository
	registrationService *service.RegistrationService
	cancellationService *service.CancellationService
	weeklyListService   *service.WeeklyListService
	logger              *zap.Logger
}

// NewWebhookHandler 建立新的 webhook handler
func NewWebhookHandler(
	channelSecret string,
	groupRepo repository.GroupRepository,
	roomRepo repository.RoomRepository,
	messageRepo repository.MessageRepository,
	registrationService *service.RegistrationService,
	cancellationService *service.CancellationService,
	weeklyListService *service.WeeklyListService,
	logger *zap.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		channelSecret:       channelSecret,
		groupRepo:           groupRepo,
		roomRepo:            roomRepo,
		messageRepo:         messageRepo,
		registrationService: registrationService,
		cancellationService: cancellationService,
		weeklyListService:   weeklyListService,
		logger:              logger,
	}
}

// Handle 處理 LINE webhook 請求
func (h *WebhookHandler) Handle(c *gin.Context) {
	// 解析事件（SDK 內部會自動驗證簽名）
	cb, err := webhook.ParseRequest(h.channelSecret, c.Request)
	if err != nil {
		if err == webhook.ErrInvalidSignature {
			h.logger.Warn("簽名驗證失敗")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}
		h.logger.Error("解析 webhook 請求失敗", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook request"})
		return
	}

	// 處理每個事件
	ctx := c.Request.Context()
	for _, event := range cb.Events {
		if err := h.handleEvent(ctx, event); err != nil {
			h.logger.Error("處理事件失敗",
				zap.String("type", string(event.GetType())),
				zap.Error(err),
			)
			// 繼續處理其他事件，不中斷
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleEvent 處理單個事件
func (h *WebhookHandler) handleEvent(ctx context.Context, event webhook.EventInterface) error {
	switch e := event.(type) {
	case webhook.JoinEvent:
		return h.handleJoinEvent(ctx, e)
	case webhook.PostbackEvent:
		return h.handlePostbackEvent(ctx, e)
	case webhook.MessageEvent:
		return h.handleMessageEvent(ctx, e)
	default:
		h.logger.Info("未處理的事件類型", zap.String("type", string(event.GetType())))
		return nil
	}
}

// handleJoinEvent 處理加入群組/聊天室事件
func (h *WebhookHandler) handleJoinEvent(ctx context.Context, event webhook.JoinEvent) error {
	source := event.Source
	timestamp := time.Unix(0, event.Timestamp*int64(time.Millisecond))

	switch s := source.(type) {
	case webhook.GroupSource:
		// 保存群組資訊
		group := &domain.Group{
			GroupID:     s.GroupId,
			GroupName:   "", // 需要額外 API 呼叫獲取
			PictureURL:  "", // 需要額外 API 呼叫獲取
			JoinDate:    timestamp,
			MemberCount: 0, // 需要額外 API 呼叫獲取
		}

		if err := h.groupRepo.SaveOrUpdate(ctx, group); err != nil {
			return fmt.Errorf("保存群組資訊失敗: %w", err)
		}

		h.logger.Info("成功加入群組", zap.String("groupID", s.GroupId))

	case webhook.RoomSource:
		// 保存聊天室資訊
		room := &domain.Room{
			RoomID:   s.RoomId,
			JoinDate: timestamp,
		}

		if err := h.roomRepo.Save(ctx, room); err != nil {
			return fmt.Errorf("保存聊天室資訊失敗: %w", err)
		}

		h.logger.Info("成功加入聊天室", zap.String("roomID", s.RoomId))
	}

	return nil
}

// handlePostbackEvent 處理 postback 事件
func (h *WebhookHandler) handlePostbackEvent(ctx context.Context, event webhook.PostbackEvent) error {
	source := event.Source
	postbackData := event.Postback.Data
	replyToken := event.ReplyToken
	timestamp := time.Unix(0, event.Timestamp*int64(time.Millisecond))

	// 解析 postback data
	params := parsePostbackData(postbackData)
	action := params["action"]
	if strings.Contains(params["message"], "季繳") {
		action = "register_season"
	} else if strings.Contains(params["message"], "臨打") {
		action = "register_casual"
	}

	// 決定 source ID 和 user ID
	var sourceID, userID string
	switch s := source.(type) {
	case webhook.GroupSource:
		sourceID = s.GroupId
		userID = s.UserId
	case webhook.RoomSource:
		sourceID = s.RoomId
		userID = s.UserId
	case webhook.UserSource:
		sourceID = s.UserId
		userID = s.UserId
	default:
		return fmt.Errorf("未知的 source type")
	}

	// 根據 action 處理
	switch action {
	case "register_season", "register_casual":
		// 報名處理
		activityID := params["activityId"]

		var message string
		if action == "register_season" {
			message = "季繳"
		} else {
			message = "臨打"
		}

		return h.registrationService.HandleRegistration(
			ctx,
			sourceID,
			activityID,
			message,
			userID,
			replyToken,
			timestamp,
		)

	case "cancel":
		// 取消報名
		registrationID := params["registrationId"]

		return h.cancellationService.HandleCancellation(
			ctx,
			sourceID,
			registrationID,
			userID,
			replyToken,
		)

	case "weekly_list":
		// 查詢本週名單
		activityID := params["activityId"]

		return h.weeklyListService.HandleWeeklyListRequest(
			ctx,
			sourceID,
			activityID,
			replyToken,
		)

	default:
		h.logger.Warn("未知的 postback action", zap.String("action", action))
		return nil
	}
}

// handleMessageEvent 處理訊息事件
func (h *WebhookHandler) handleMessageEvent(ctx context.Context, event webhook.MessageEvent) error {
	source := event.Source
	timestamp := time.Unix(0, event.Timestamp*int64(time.Millisecond))

	// 只處理文字訊息
	if event.Message.GetType() != "text" {
		return nil
	}

	textMsg, ok := event.Message.(webhook.TextMessageContent)
	if !ok {
		return nil
	}

	// 決定 source ID 和 user ID
	var sourceID, userID string
	switch s := source.(type) {
	case webhook.GroupSource:
		sourceID = s.GroupId
		userID = s.UserId
	case webhook.RoomSource:
		sourceID = s.RoomId
		userID = s.UserId
	case webhook.UserSource:
		sourceID = s.UserId
		userID = s.UserId
	default:
		return fmt.Errorf("未知的 source type")
	}

	// 保存訊息記錄
	message := &domain.Message{
		Timestamp: timestamp,
		SourceID:  sourceID,
		UserID:    userID,
		Content:   textMsg.Text,
	}

	if err := h.messageRepo.Save(ctx, message); err != nil {
		h.logger.Error("保存訊息記錄失敗", zap.Error(err))
		// 不返回錯誤，避免影響主流程
	}

	return nil
}

// parsePostbackData 解析 postback data
// 格式: "action=cancel,registrationId=xxx,userName=yyy"
func parsePostbackData(data string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(data, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}
