package lineapi

import (
	"context"
	"fmt"
)

// UserProfile 用戶個人資料
type UserProfile struct {
	UserID      string
	DisplayName string
	PictureURL  string
	Language    string
}

// GroupSummary 群組摘要資訊
type GroupSummary struct {
	GroupID     string
	GroupName   string
	PictureURL  string
	MemberCount int
}

// GetUserProfile 取得用戶個人資料
func (c *Client) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	resp, _, err := c.messagingAPI.GetProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("取得用戶資料失敗: %w", err)
	}

	return &UserProfile{
		UserID:      resp.UserId,
		DisplayName: resp.DisplayName,
		PictureURL:  resp.PictureUrl,
		Language:    resp.Language,
	}, nil
}

// GetGroupMemberProfile 取得群組成員個人資料
func (c *Client) GetGroupMemberProfile(ctx context.Context, groupID, userID string) (*UserProfile, error) {
	resp, _, err := c.messagingAPI.GetGroupMemberProfile(groupID, userID)
	if err != nil {
		return nil, fmt.Errorf("取得群組成員資料失敗: %w", err)
	}

	return &UserProfile{
		UserID:      resp.UserId,
		DisplayName: resp.DisplayName,
		PictureURL:  resp.PictureUrl,
		Language:    resp.Language,
	}, nil
}

// GetGroupSummary 取得群組摘要資訊
func (c *Client) GetGroupSummary(ctx context.Context, groupID string) (*GroupSummary, error) {
	resp, _, err := c.messagingAPI.GetGroupSummary(groupID)
	if err != nil {
		return nil, fmt.Errorf("取得群組摘要失敗: %w", err)
	}

	return &GroupSummary{
		GroupID:     resp.GroupId,
		GroupName:   resp.GroupName,
		PictureURL:  resp.PictureUrl,
		MemberCount: int(resp.Count),
	}, nil
}
