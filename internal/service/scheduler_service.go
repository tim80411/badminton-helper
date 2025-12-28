package service

import (
	"context"
	"fmt"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/tim80411/badminton-helper/internal/lineapi"
	"github.com/tim80411/badminton-helper/internal/repository"
	"github.com/tim80411/badminton-helper/internal/util"
	"go.uber.org/zap"
)

// SchedulerService 定時任務服務
type SchedulerService struct {
	groupRepo  repository.GroupRepository
	lineClient *lineapi.Client
	settings   lineapi.ActivitySettings
	logger     *zap.Logger
}

// NewSchedulerService 建立新的定時任務服務
func NewSchedulerService(
	groupRepo repository.GroupRepository,
	lineClient *lineapi.Client,
	settings lineapi.ActivitySettings,
	logger *zap.Logger,
) *SchedulerService {
	return &SchedulerService{
		groupRepo:  groupRepo,
		lineClient: lineClient,
		settings:   settings,
		logger:     logger,
	}
}

// SendScheduledMessage 發送定時報名通知
func (s *SchedulerService) SendScheduledMessage(ctx context.Context) error {
	s.logger.Info("開始發送定時報名通知")

	// 取得所有群組
	groups, err := s.groupRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("取得群組列表失敗", zap.Error(err))
		return fmt.Errorf("取得群組列表失敗: %w", err)
	}

	// 生成活動 ID
	activityID := util.GenerateActivityID()

	// 建立活動報名 Flex Message
	flexMsg := lineapi.BuildActivityFlexMessage(activityID, s.settings)

	// 發送到每個群組
	successCount := 0
	failCount := 0

	for _, group := range groups {
		err := s.lineClient.PushMessage(ctx, group.GroupID, []messaging_api.MessageInterface{flexMsg})
		if err != nil {
			s.logger.Error("發送訊息到群組失敗",
				zap.String("groupID", group.GroupID),
				zap.String("groupName", group.GroupName),
				zap.Error(err),
			)
			failCount++
		} else {
			s.logger.Info("成功發送訊息到群組",
				zap.String("groupID", group.GroupID),
				zap.String("groupName", group.GroupName),
			)
			successCount++
		}
	}

	s.logger.Info("定時報名通知發送完成",
		zap.Int("成功", successCount),
		zap.Int("失敗", failCount),
		zap.Int("總計", len(groups)),
	)

	return nil
}
