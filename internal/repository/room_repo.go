package repository

import (
	"context"
	"fmt"

	"github.com/tim80411/badminton-helper/internal/domain"
)

const (
	RoomsSheet = "Rooms"
)

// RoomRepository 定義聊天室資訊的資料操作介面
type RoomRepository interface {
	Save(ctx context.Context, room *domain.Room) error
	FindByID(ctx context.Context, roomID string) (*domain.Room, error)
	GetAll(ctx context.Context) ([]*domain.Room, error)
}

type roomRepo struct {
	client *SheetsClient
}

// NewRoomRepository 建立新的聊天室 repository
func NewRoomRepository(client *SheetsClient) RoomRepository {
	return &roomRepo{
		client: client,
	}
}

// Save 儲存聊天室資訊（不重複儲存）
func (r *roomRepo) Save(ctx context.Context, room *domain.Room) error {
	// 確保 sheet 存在
	headers := []interface{}{"Room ID", "Join Date"}
	err := r.client.EnsureSheetExists(ctx, RoomsSheet, headers)
	if err != nil {
		return fmt.Errorf("確保 Rooms sheet 存在失敗: %w", err)
	}

	// 檢查是否已存在
	existing, _ := r.FindByID(ctx, room.RoomID)
	if existing != nil {
		return nil // 已存在，不重複新增
	}

	// 新增聊天室
	values := []interface{}{
		room.RoomID,
		room.JoinDate.Format("2006-01-02 15:04:05"),
	}
	err = r.client.AppendRow(ctx, RoomsSheet, values)
	if err != nil {
		return fmt.Errorf("新增聊天室資訊失敗: %w", err)
	}

	return nil
}

// FindByID 根據聊天室 ID 查詢聊天室資訊
func (r *roomRepo) FindByID(ctx context.Context, roomID string) (*domain.Room, error) {
	readRange := fmt.Sprintf("%s!A:B", RoomsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return nil, fmt.Errorf("讀取聊天室資訊失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return nil, fmt.Errorf("找不到聊天室")
	}

	for _, row := range valueRange.Values[1:] {
		if len(row) > 0 && row[0] == roomID {
			joinDate, _ := parseTimestamp(row[1])
			return &domain.Room{
				RoomID:   getStringValue(row, 0),
				JoinDate: joinDate,
			}, nil
		}
	}

	return nil, fmt.Errorf("找不到 ID 為 %s 的聊天室", roomID)
}

// GetAll 取得所有聊天室
func (r *roomRepo) GetAll(ctx context.Context) ([]*domain.Room, error) {
	readRange := fmt.Sprintf("%s!A:B", RoomsSheet)
	valueRange, err := r.client.ReadRange(ctx, readRange)
	if err != nil {
		return nil, fmt.Errorf("讀取聊天室資訊失敗: %w", err)
	}

	if len(valueRange.Values) <= 1 {
		return []*domain.Room{}, nil
	}

	var rooms []*domain.Room
	for _, row := range valueRange.Values[1:] {
		if len(row) == 0 {
			continue
		}

		joinDate, _ := parseTimestamp(row[1])
		room := &domain.Room{
			RoomID:   getStringValue(row, 0),
			JoinDate: joinDate,
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}
