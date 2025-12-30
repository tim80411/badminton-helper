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
- `GOOGLE_APPLICATION_CREDENTIALS`：GCP 服務帳號 JSON 金鑰路徑（本地開發）
- `GOOGLE_CREDENTIALS_BASE64`：Base64 編碼的 GCP 服務帳號 JSON（雲端部署，與上者二擇一）

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

## Zeabur 部署

### 1. 建立專案

在 [Zeabur](https://zeabur.com) 建立新專案，並連接你的 GitHub 倉庫。

### 2. 設定環境變數

在 Zeabur 的專案設定中，新增以下環境變數：

| 變數名稱 | 說明 |
|---------|------|
| `LINE_CHANNEL_ACCESS_TOKEN` | LINE Channel Access Token |
| `LINE_CHANNEL_SECRET` | LINE Channel Secret |
| `GOOGLE_SHEETS_SPREADSHEET_ID` | Google Spreadsheet ID |
| `GOOGLE_CREDENTIALS_BASE64` | Base64 編碼的服務帳號 JSON（見下方說明） |
| `GIN_MODE` | 設為 `release` |

### 3. 設定 Google 憑證

將 `service-account.json` 進行 **Base64 編碼**後，貼到 `GOOGLE_CREDENTIALS_BASE64` 環境變數中：

```bash
# macOS / Linux - 產生 Base64 編碼字串
base64 -i credentials/service-account.json | tr -d '\n'

# 或者使用這個命令（某些系統）
cat credentials/service-account.json | base64 | tr -d '\n'
```

將產生的 Base64 字串複製貼到 Zeabur 的環境變數中。

### 4. 部署

Zeabur 會自動偵測 Dockerfile 並進行部署。部署完成後，設定 LINE Webhook URL 為：
```
https://your-app.zeabur.app/webhook
```

---

## Docker 部署

### 使用 Docker Compose

```bash
# 進入 deployments 目錄
cd deployments

# 建置並啟動容器
docker-compose up -d

# 查看日誌
docker-compose logs -f

# 停止容器
docker-compose down
```

### 使用 Docker 指令

```bash
# 建置映像檔
docker build -t badminton-helper -f deployments/Dockerfile .

# 執行容器
docker run -d \
  --name badminton-helper \
  -p 8080:8080 \
  -v $(pwd)/credentials/service-account.json:/home/appuser/credentials/service-account.json:ro \
  -v $(pwd)/configs/settings.yaml:/home/appuser/configs/settings.yaml:ro \
  --env-file configs/.env \
  badminton-helper

# 查看日誌
docker logs -f badminton-helper
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

## License

MIT
