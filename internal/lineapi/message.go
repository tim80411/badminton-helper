package lineapi

import (
	"context"
	"fmt"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

// ReplyMessage 使用 reply token 回覆訊息
func (c *Client) ReplyMessage(ctx context.Context, replyToken string, messages []messaging_api.MessageInterface) error {
	req := &messaging_api.ReplyMessageRequest{
		ReplyToken:           replyToken,
		Messages:             messages,
		NotificationDisabled: ptr(true), // 不顯示通知
	}

	_, err := c.messagingAPI.ReplyMessage(req)
	if err != nil {
		return fmt.Errorf("回覆訊息失敗: %w", err)
	}

	return nil
}

// PushMessage 使用 push API 發送訊息
func (c *Client) PushMessage(ctx context.Context, targetID string, messages []messaging_api.MessageInterface) error {
	req := &messaging_api.PushMessageRequest{
		To:       targetID,
		Messages: messages,
	}

	_, err := c.messagingAPI.PushMessage(req, "")
	if err != nil {
		return fmt.Errorf("發送訊息失敗: %w", err)
	}

	return nil
}

// ptr 輔助函數：建立指標
func ptr[T any](v T) *T {
	return &v
}
