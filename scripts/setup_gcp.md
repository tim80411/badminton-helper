# Google Cloud Platform 設定指南

本指南將協助您設定 Google Cloud Platform，以便使用 Google Sheets API 作為應用程式的資料儲存。

## 前置要求

- Google 帳號
- 已建立 Google Spreadsheet 作為資料儲存

## 步驟 1: 建立 GCP 專案

1. 前往 [Google Cloud Console](https://console.cloud.google.com/)
2. 點擊頂部導覽列的專案選擇器
3. 點擊「新增專案」
4. 輸入專案名稱：`badminton-helper`
5. 點擊「建立」

## 步驟 2: 啟用 Google Sheets API

1. 在 GCP Console 中，確保您已選擇 `badminton-helper` 專案
2. 前往「API 和服務」→「資料庫」
3. 點擊「+ 啟用 API 和服務」
4. 搜尋「Google Sheets API」
5. 點擊「Google Sheets API」
6. 點擊「啟用」

## 步驟 3: 建立服務帳號

1. 前往「API 和服務」→「憑證」
2. 點擊「+ 建立憑證」→「服務帳號」
3. 填寫服務帳號詳細資料：
   - 服務帳號名稱：`badminton-helper-sa`
   - 服務帳號 ID：`badminton-helper-sa`（自動生成）
   - 服務帳號說明：`Service account for badminton helper bot`
4. 點擊「建立並繼續」
5. **跳過**「將此服務帳號存取權授予專案」步驟（不需要專案層級角色）
6. 點擊「繼續」
7. **跳過**「將存取權授予使用者」步驟
8. 點擊「完成」

## 步驟 4: 下載服務帳號金鑰

1. 在「憑證」頁面中，找到剛才建立的服務帳號
2. 點擊服務帳號名稱（或右側的編輯圖示）
3. 切換到「金鑰」分頁
4. 點擊「新增金鑰」→「建立新金鑰」
5. 選擇「JSON」格式
6. 點擊「建立」
7. JSON 金鑰檔案會自動下載到您的電腦

### 儲存金鑰檔案

```bash
# 將下載的 JSON 檔案重新命名並移動到專案的 credentials 目錄
mv ~/Downloads/badminton-helper-xxxxxx.json /Users/tim80411/self/badminton-helper/credentials/service-account.json

# 設定適當的權限（確保只有擁有者可以讀取）
chmod 600 /Users/tim80411/self/badminton-helper/credentials/service-account.json
```

**⚠️ 安全提醒**：
- **絕對不要**將此檔案提交到版本控制系統
- `credentials/` 目錄已在 `.gitignore` 中排除
- 妥善保管此檔案，它包含存取您的 Google Cloud 資源的憑證

## 步驟 5: 設定 Google Spreadsheet 權限

1. 開啟下載的 JSON 金鑰檔案
2. 找到 `client_email` 欄位，複製 email 地址（格式：`badminton-helper-sa@badminton-helper.iam.gserviceaccount.com`）
3. 開啟您的 Google Spreadsheet
4. 點擊右上角的「共用」按鈕
5. 貼上服務帳號的 email 地址
6. 設定權限為「編輯者」
7. 取消勾選「通知使用者」
8. 點擊「共用」

## 步驟 6: 取得 Spreadsheet ID

1. 開啟您的 Google Spreadsheet
2. 從 URL 中複製 Spreadsheet ID：
   ```
   https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit
   ```
3. 將 Spreadsheet ID 填入 `configs/.env` 檔案的 `GOOGLE_SHEETS_SPREADSHEET_ID` 欄位

## 步驟 7: 準備 Spreadsheet 結構

確保您的 Spreadsheet 包含以下工作表（Sheets）：

### 1. Registrations（報名記錄）

| 欄位 | 說明 |
|------|------|
| Timestamp | 報名時間 |
| Group ID | 群組 ID |
| ActivityId | 活動 ID |
| User ID | 使用者 ID |
| User Name | 使用者名稱 |
| Registration Type | 報名類型（季繳/臨打） |
| Message | 訊息內容 |
| Court Assignment | 場地分配 |
| isCancelled | 是否已取消 |

### 2. Groups（群組資訊）

| 欄位 | 說明 |
|------|------|
| Group ID | 群組 ID |
| Group Name | 群組名稱 |
| Picture URL | 群組圖片 URL |
| Join Date | 加入日期 |
| Member Count | 成員數量 |

### 3. Rooms（聊天室資訊）

| 欄位 | 說明 |
|------|------|
| Room ID | 聊天室 ID |
| Join Date | 加入日期 |

### 4. Messages（訊息記錄）

| 欄位 | 說明 |
|------|------|
| Timestamp | 時間戳記 |
| Source ID | 來源 ID |
| User ID | 使用者 ID |
| Message | 訊息內容 |

**注意**：應用程式會自動建立缺少的工作表，但您也可以手動建立並設定標題列。

## 步驟 8: 設定環境變數

編輯 `configs/.env` 檔案：

```env
LINE_CHANNEL_ACCESS_TOKEN=your_channel_access_token
LINE_CHANNEL_SECRET=your_channel_secret
LINE_ADMIN_IDS=U8db100745249a1cb2fc27205d6a19a87,U5b731eb8c71dff9f22cffda138413731,U1f16d9d8f822e867d6931cb1b2d89382
GOOGLE_SHEETS_SPREADSHEET_ID=your_spreadsheet_id_from_step_6
GOOGLE_APPLICATION_CREDENTIALS=./credentials/service-account.json
SERVER_PORT=8080
GIN_MODE=debug
TZ=Asia/Taipei
```

## 驗證設定

執行以下命令測試設定是否正確：

```bash
# 執行應用程式
go run cmd/server/main.go
```

如果看到以下日誌，表示設定成功：

```
{"level":"info","msg":"應用程式啟動中...","port":"8080","mode":"debug"}
{"level":"info","msg":"定時任務已啟動","schedule":"每週二 13:00 (Asia/Taipei)"}
{"level":"info","msg":"HTTP Server 啟動","addr":":8080"}
```

## 常見問題

### Q: 出現「permission denied」錯誤

**A**: 確認：
1. 服務帳號的 email 已加入 Spreadsheet 的共用設定
2. 權限設定為「編輯者」
3. 金鑰檔案路徑正確

### Q: 出現「API has not been used in project」錯誤

**A**: 確認：
1. Google Sheets API 已在專案中啟用
2. 等待幾分鐘讓 API 啟用生效

### Q: 金鑰檔案遺失或洩漏怎麼辦？

**A**:
1. 前往 GCP Console → 服務帳號
2. 刪除舊的金鑰
3. 建立新的金鑰
4. 更新應用程式中的金鑰檔案

## 成本說明

Google Sheets API 的使用在一般情況下是免費的：

- 免費配額：每天 500 次讀取請求、100 次寫入請求（每個使用者、每個專案）
- 本應用程式的正常使用量遠低於此配額
- 如需提高配額，可參考 [Google Sheets API 定價](https://developers.google.com/sheets/api/limits)

## 相關資源

- [Google Sheets API 文件](https://developers.google.com/sheets/api)
- [服務帳號文件](https://cloud.google.com/iam/docs/service-accounts)
- [Google Cloud Console](https://console.cloud.google.com/)
