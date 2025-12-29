# 階段 1: 建置階段
FROM golang:1.24-alpine AS builder

# 安裝必要的建置工具
RUN apk add --no-cache git ca-certificates tzdata

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製原始碼
COPY . .

# 建置應用程式
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o server ./cmd/server

# 階段 2: 執行階段
FROM alpine:latest

# 安裝 CA 證書和時區資料
RUN apk --no-cache add ca-certificates tzdata

# 建立非 root 使用者
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 設定工作目錄
WORKDIR /home/appuser

# 從建置階段複製二進位檔
COPY --from=builder --chown=appuser:appuser /app/server ./server

# 複製配置檔案
COPY --chown=appuser:appuser configs/ ./configs/

# 建立 credentials 目錄
RUN mkdir -p ./credentials && chown appuser:appuser ./credentials

# 設定環境變數
ENV TZ=Asia/Taipei
ENV GIN_MODE=release

# 切換到非 root 使用者
USER appuser

# 暴露端口
EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 執行應用程式
CMD ["./server"]
