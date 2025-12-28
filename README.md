# Badminton Helper - LINE Bot 羽球報名系統

這是一個使用 Golang 開發的 LINE Bot，用於管理羽球活動報名。

## 功能特色

- 🏸 **報名管理**：支援季繳和臨打兩種報名類型
- ✅ **取消報名**：用戶可以取消自己的報名
- 📋 **週名單查詢**：查看本週報名名單（管理員限定）
- ⏰ **定時通知**：每週二 13:00 自動發送報名通知
- 📊 **Google Sheets**：使用 Google Sheets 作為資料存儲，方便查看和編輯

## 技術棧

- **語言**：Go 1.24
- **Web 框架**：Gin
- **LINE Bot SDK**：line-bot-sdk-go v8
- **資料存儲**：Google Sheets API
- **定時任務**：robfig/cron v3

## 活動資訊

- **地點**：成功高中運動中心 8 樓
- **時間**：週六 20:00-22:00
- **費用**：單次臨打 200，含場地&球錢
- **名額**：限額 15 位

## 快速開始

### 1. 環境準備

```bash
# 確保已安裝 Go 1.24+
go version

# Clone 專案
git clone <your-repo>
cd badminton-helper
```

### 2. 設定環境變數

```bash
# 複製環境變數範例檔案
cp configs/.env.example configs/.env

# 編輯 .env 填入實際值
vim configs/.env
```

需要設定的環境變數：
- `LINE_CHANNEL_ACCESS_TOKEN`：LINE Channel Access Token
- `LINE_CHANNEL_SECRET`：LINE Channel Secret
- `GOOGLE_SHEETS_SPREADSHEET_ID`：Google Spreadsheet ID
- `GOOGLE_APPLICATION_CREDENTIALS`：GCP 服務帳號 JSON 金鑰路徑

### 3. 設定 Google Cloud

詳細步驟請參考 `scripts/setup_gcp.md`

### 4. 執行應用程式

```bash
# 安裝依賴
go mod download

# 執行
go run cmd/server/main.go
```

### 5. 設定 LINE Webhook

使用 ngrok 建立本地開發用的 webhook：

```bash
# 啟動 ngrok
ngrok http 8080

# 在 LINE Developers Console 設定 Webhook URL
# https://xxxx.ngrok.io/webhook
```

## 專案結構

```
badminton-helper/
├── cmd/server/           # 主程式進入點
├── internal/
│   ├── config/          # 配置管理
│   ├── domain/          # 資料模型
│   ├── handler/         # HTTP handlers
│   ├── service/         # 業務邏輯層
│   ├── repository/      # 資料存取層
│   ├── lineapi/         # LINE API 封裝
│   └── util/            # 工具函數
├── configs/             # 配置檔案
├── credentials/         # GCP 服務帳號金鑰（不納入版控）
└── deployments/         # Docker 相關檔案
```

## 開發狀態

### 已完成
- [x] Phase 1: 基礎設施搭建

### 進行中
- [ ] Phase 2: Google Sheets 整合
- [ ] Phase 3: LINE API 整合
- [ ] Phase 4: 業務邏輯層
- [ ] Phase 5: HTTP Handler 層
- [ ] Phase 6: 主程式與配置
- [ ] Phase 7: 工具函數
- [ ] Phase 8: Docker 化與部署

## License

MIT
