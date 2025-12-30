package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tim80411/badminton-helper/internal/lineapi"
	"gopkg.in/yaml.v3"
)

// Config 應用程式配置
type Config struct {
	// LINE Bot 設定
	LineChannelAccessToken string
	LineChannelSecret      string
	LineAdminIDs           []string

	// Google Sheets 設定
	GoogleSheetsSpreadsheetID string
	GoogleCredentialsPath     string // 憑證檔案路徑（本地開發）
	GoogleCredentialsBase64   string // Base64 編碼的憑證 JSON（雲端部署）

	// 伺服器設定
	ServerPort string
	GinMode    string

	// 活動設定
	ActivitySettings lineapi.ActivitySettings
}

// Load 載入配置
func Load(envPath, settingsPath string) (*Config, error) {
	// 載入 .env 檔案
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil && os.Getenv("GIN_MODE") != "debug" {
			return nil, fmt.Errorf("載入 .env 檔案失敗: %w", err)
		}
	}

	// 讀取環境變數
	config := &Config{
		LineChannelAccessToken:    os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		LineChannelSecret:         os.Getenv("LINE_CHANNEL_SECRET"),
		GoogleSheetsSpreadsheetID: os.Getenv("GOOGLE_SHEETS_SPREADSHEET_ID"),
		GoogleCredentialsPath:     os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		GoogleCredentialsBase64:   os.Getenv("GOOGLE_CREDENTIALS_BASE64"),
		ServerPort:                getEnvOrDefault("SERVER_PORT", "8080"),
		GinMode:                   getEnvOrDefault("GIN_MODE", "debug"),
	}

	// 解析管理員 ID 列表
	adminIDsStr := os.Getenv("LINE_ADMIN_IDS")
	if adminIDsStr != "" {
		config.LineAdminIDs = strings.Split(adminIDsStr, ",")
		// 去除空白
		for i := range config.LineAdminIDs {
			config.LineAdminIDs[i] = strings.TrimSpace(config.LineAdminIDs[i])
		}
	}

	// 驗證必要欄位
	if err := config.validate(); err != nil {
		return nil, err
	}

	// 載入活動設定
	if settingsPath != "" {
		settings, err := loadActivitySettings(settingsPath)
		if err != nil {
			return nil, fmt.Errorf("載入活動設定失敗: %w", err)
		}
		config.ActivitySettings = settings
	}

	return config, nil
}

// validate 驗證配置
func (c *Config) validate() error {
	if c.LineChannelAccessToken == "" {
		return fmt.Errorf("LINE_CHANNEL_ACCESS_TOKEN 未設定")
	}
	if c.LineChannelSecret == "" {
		return fmt.Errorf("LINE_CHANNEL_SECRET 未設定")
	}
	if c.GoogleSheetsSpreadsheetID == "" {
		return fmt.Errorf("GOOGLE_SHEETS_SPREADSHEET_ID 未設定")
	}
	// 至少要有一種 Google 憑證方式
	if c.GoogleCredentialsPath == "" && c.GoogleCredentialsBase64 == "" {
		return fmt.Errorf("必須設定 GOOGLE_APPLICATION_CREDENTIALS 或 GOOGLE_CREDENTIALS_BASE64 其中之一")
	}
	return nil
}

// loadActivitySettings 載入活動設定
func loadActivitySettings(path string) (lineapi.ActivitySettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return lineapi.ActivitySettings{}, err
	}

	var config struct {
		Activity lineapi.ActivitySettings `yaml:"activity"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return lineapi.ActivitySettings{}, err
	}

	return config.Activity, nil
}

// getEnvOrDefault 取得環境變數或預設值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
