package repository

import (
	"context"
	"fmt"

	"github.com/tim80411/badminton-helper/internal/domain"
)

const (
	MessagesSheet = "Messages"
)

// MessageRepository 定義訊息記錄的資料操作介面
type MessageRepository interface {
	Save(ctx context.Context, message *domain.Message) error
}

type messageRepo struct {
	client *SheetsClient
}

// NewMessageRepository 建立新的訊息 repository
func NewMessageRepository(client *SheetsClient) MessageRepository {
	return &messageRepo{
		client: client,
	}
}

// Save 儲存訊息記錄
func (r *messageRepo) Save(ctx context.Context, message *domain.Message) error {
	// 確保 sheet 存在
	headers := []interface{}{"Timestamp", "Source ID", "User ID", "Message"}
	err := r.client.EnsureSheetExists(ctx, MessagesSheet, headers)
	if err != nil {
		return fmt.Errorf("確保 Messages sheet 存在失敗: %w", err)
	}

	// 新增訊息
	values := []interface{}{
		message.Timestamp.Format("2006-01-02 15:04:05"),
		message.SourceID,
		message.UserID,
		message.Content,
	}
	err = r.client.AppendRow(ctx, MessagesSheet, values)
	if err != nil {
		return fmt.Errorf("新增訊息記錄失敗: %w", err)
	}

	return nil
}
