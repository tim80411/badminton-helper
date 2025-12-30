package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/tim80411/badminton-helper/internal/config"
	"github.com/tim80411/badminton-helper/internal/handler"
	"github.com/tim80411/badminton-helper/internal/lineapi"
	"github.com/tim80411/badminton-helper/internal/repository"
	"github.com/tim80411/badminton-helper/internal/service"
	"github.com/tim80411/badminton-helper/internal/util"
	"go.uber.org/zap"
)

func main() {
	// 初始化 logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("初始化 logger 失敗: %v", err)
	}
	defer logger.Sync()

	// 載入配置
	cfg, err := config.Load("configs/.env", "configs/settings.yaml")
	if err != nil {
		logger.Fatal("載入配置失敗", zap.Error(err))
	}

	// 設定 Gin 模式
	gin.SetMode(cfg.GinMode)

	// 設定時區
	if tz := os.Getenv("TZ"); tz != "" {
		os.Setenv("TZ", tz)
	}

	logger.Info("應用程式啟動中...",
		zap.String("port", cfg.ServerPort),
		zap.String("mode", cfg.GinMode),
	)

	// 初始化 Google Sheets 客戶端
	ctx := context.Background()
	credentials := repository.SheetsCredentials{
		FilePath:   cfg.GoogleCredentialsPath,
		JSONBase64: cfg.GoogleCredentialsBase64,
	}
	sheetsClient, err := repository.NewSheetsClient(ctx, credentials, cfg.GoogleSheetsSpreadsheetID)
	if err != nil {
		logger.Fatal("初始化 Google Sheets 客戶端失敗", zap.Error(err))
	}

	// 確保所有必要的 Sheet 存在並設定表頭
	requiredSheets := map[string][]interface{}{
		"Registrations": {"Timestamp", "Group ID", "ActivityId", "User ID", "User Name", "Registration Type", "Message", "Court Assignment", "isCancelled"},
		"Groups":        {"Group ID", "Group Name", "Picture URL", "Join Date", "Member Count"},
		"Rooms":         {"Room ID", "Join Date"},
		"Messages":      {"Timestamp", "Source ID", "User ID", "Message"},
	}

	for sheetName, headers := range requiredSheets {
		if err := sheetsClient.EnsureSheetExists(ctx, sheetName, headers); err != nil {
			logger.Fatal("確保 Sheet 存在失敗",
				zap.String("sheet", sheetName),
				zap.Error(err),
			)
		}
	}

	// 初始化 Repository 層
	registrationRepo := repository.NewRegistrationRepository(sheetsClient)
	groupRepo := repository.NewGroupRepository(sheetsClient)
	roomRepo := repository.NewRoomRepository(sheetsClient)
	messageRepo := repository.NewMessageRepository(sheetsClient)

	// 初始化 LINE Bot 客戶端
	lineClient, err := lineapi.NewClient(cfg.LineChannelAccessToken)
	if err != nil {
		logger.Fatal("初始化 LINE Bot 客戶端失敗", zap.Error(err))
	}

	// 初始化 Service 層
	registrationService := service.NewRegistrationService(registrationRepo, lineClient)
	cancellationService := service.NewCancellationService(registrationRepo, lineClient)
	weeklyListService := service.NewWeeklyListService(registrationRepo, lineClient)
	schedulerService := service.NewSchedulerService(groupRepo, lineClient, cfg.ActivitySettings, logger)

	// 初始化 Handler 層
	webhookHandler := handler.NewWebhookHandler(
		cfg.LineChannelSecret,
		groupRepo,
		roomRepo,
		messageRepo,
		registrationService,
		cancellationService,
		weeklyListService,
		logger,
	)
	healthHandler := handler.NewHealthHandler()

	// 設定路由
	router := gin.Default()

	// 健康檢查端點
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// LINE webhook 端點
	router.POST("/webhook", webhookHandler.Handle)

	// 設定定時任務（每週二 13:00）
	c := cron.New(cron.WithLocation(util.TaipeiTZ))
	_, err = c.AddFunc("0 13 * * 2", func() {
		logger.Info("執行定時報名通知任務")
		if err := schedulerService.SendScheduledMessage(ctx); err != nil {
			logger.Error("定時報名通知任務執行失敗", zap.Error(err))
		}
	})
	if err != nil {
		logger.Fatal("設定定時任務失敗", zap.Error(err))
	}

	c.Start()
	logger.Info("定時任務已啟動", zap.String("schedule", "每週二 13:00 (Asia/Taipei)"))

	// 建立 HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}

	// 在 goroutine 中啟動伺服器
	go func() {
		logger.Info("HTTP Server 啟動", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP Server 啟動失敗", zap.Error(err))
		}
	}()

	// 等待中斷信號以優雅關閉伺服器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在關閉伺服器...")

	// 停止定時任務
	c.Stop()

	// 優雅關閉，等待最多 5 秒
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("伺服器關閉失敗", zap.Error(err))
	}

	logger.Info("伺服器已關閉")
}
