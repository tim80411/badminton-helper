package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tim80411/badminton-helper/internal/domain"
)

const (
	GroupsSheet = "Groups"
)

// GroupRepository 定義群組資訊的資料操作介面
type GroupRepository interface {
	SaveOrUpdate(ctx context.Context, group *domain.Group) error
	FindByID(ctx context.Context, groupID string) (*domain.Group, error)
	GetAll(ctx context.Context) ([]*domain.Group, error)
}

type groupRepo struct {
	client *SheetsClient
}

// NewGroupRepository 建立新的群組 repository
func NewGroupRepository(client *SheetsClient) GroupRepository {
	return &groupRepo{
		client: client,
	}
}

// SaveOrUpdate 儲存或更新群組資訊
func (r *groupRepo) SaveOrUpdate(ctx context.Context, group *domain.Group) error {
	// 確保 sheet 存在
	headers := []interface{}{
		"Group ID", "Group Name", "Picture URL", "Join Date", "Member Count",
	}
	err := r.client.EnsureSheetExists(ctx, GroupsSheet, headers)
	if err != nil {
		return fmt.Errorf("確保 Groups sheet 存在失敗: %w", err)
	}

	// 檢查群組是否已存在
	readRange := fmt.Sprintf("%s!A:E", GroupsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return fmt.Errorf("讀取群組資訊失敗: %w", err)
	}

	// 找到群組所在的行
	rowNum := -1
	if len(valueRange.Values) > 1 {
		for i, row := range valueRange.Values[1:] {
			if len(row) > 0 && row[0] == group.GroupID {
				rowNum = i + 2 // i+2: i 從 values[1:] 開始，+1 跳過表頭，+1 因為 Sheets 從 1 開始
				break
			}
		}
	}

	if rowNum > 0 {
		// 更新現有群組
		updates := []*sheets.ValueRange{
			{
				Range:  fmt.Sprintf("%s!B%d", GroupsSheet, rowNum),
				Values: [][]interface{}{{group.GroupName}},
			},
			{
				Range:  fmt.Sprintf("%s!C%d", GroupsSheet, rowNum),
				Values: [][]interface{}{{group.PictureURL}},
			},
			{
				Range:  fmt.Sprintf("%s!E%d", GroupsSheet, rowNum),
				Values: [][]interface{}{{group.MemberCount}},
			},
		}
		err = r.client.BatchUpdate(ctx, updates)
		if err != nil {
			return fmt.Errorf("更新群組資訊失敗: %w", err)
		}
	} else {
		// 新增群組
		values := []interface{}{
			group.GroupID,
			group.GroupName,
			group.PictureURL,
			group.JoinDate.Format("2006-01-02 15:04:05"),
			group.MemberCount,
		}
		err = r.client.AppendRow(ctx, GroupsSheet, values)
		if err != nil {
			return fmt.Errorf("新增群組資訊失敗: %w", err)
		}
	}

	return nil
}

// FindByID 根據群組 ID 查詢群組資訊
func (r *groupRepo) FindByID(ctx context.Context, groupID string) (*domain.Group, error) {
	readRange := fmt.Sprintf("%s!A:E", GroupsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return nil, fmt.Errorf("讀取群組資訊失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return nil, fmt.Errorf("找不到群組")
	}

	for _, row := range valueRange.Values[1:] {
		if len(row) > 0 && row[0] == groupID {
			joinDate, _ := parseTimestamp(row[3])
			memberCount, _ := strconv.Atoi(getStringValue(row, 4))

			return &domain.Group{
				GroupID:     getStringValue(row, 0),
				GroupName:   getStringValue(row, 1),
				PictureURL:  getStringValue(row, 2),
				JoinDate:    joinDate,
				MemberCount: memberCount,
			}, nil
		}
	}

	return nil, fmt.Errorf("找不到 ID 為 %s 的群組", groupID)
}

// GetAll 取得所有群組
func (r *groupRepo) GetAll(ctx context.Context) ([]*domain.Group, error) {
	readRange := fmt.Sprintf("%s!A:E", GroupsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return nil, fmt.Errorf("讀取群組資訊失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return []*domain.Group{}, nil
	}

	var groups []*domain.Group
	for _, row := range valueRange.Values[1:] {
		if len(row) == 0 {
			continue
		}

		joinDate, _ := parseTimestamp(row[3])
		memberCount, _ := strconv.Atoi(getStringValue(row, 4))

		group := &domain.Group{
			GroupID:     getStringValue(row, 0),
			GroupName:   getStringValue(row, 1),
			PictureURL:  getStringValue(row, 2),
			JoinDate:    joinDate,
			MemberCount: memberCount,
		}
		groups = append(groups, group)
	}

	return groups, nil
}
