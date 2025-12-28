package lineapi

import (
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

// Client 封裝 LINE Messaging API 客戶端
type Client struct {
	messagingAPI *messaging_api.MessagingApiAPI
	channelToken string
}

// NewClient 建立新的 LINE API 客戶端
func NewClient(channelToken, channelSecret string) (*Client, error) {
	api, err := messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		return nil, err
	}

	return &Client{
		messagingAPI: api,
		channelToken: channelToken,
	}, nil
}

// GetMessagingAPI 取得 Messaging API 實例
func (c *Client) GetMessagingAPI() *messaging_api.MessagingApiAPI {
	return c.messagingAPI
}
