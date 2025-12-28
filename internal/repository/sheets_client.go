package repository

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient 封裝 Google Sheets API 客戶端
type SheetsClient struct {
	service       *sheets.Service
	spreadsheetID string
}

// NewSheetsClient 建立新的 Sheets 客戶端
func NewSheetsClient(ctx context.Context, credentialsPath, spreadsheetID string) (*SheetsClient, error) {
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("無法建立 Sheets 服務: %w", err)
	}

	return &SheetsClient{
		service:       service,
		spreadsheetID: spreadsheetID,
	}, nil
}

// ReadRange 讀取指定範圍的資料
func (c *SheetsClient) ReadRange(ctx context.Context, readRange string) (*sheets.ValueRange, error) {
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("讀取範圍 %s 失敗: %w", readRange, err)
	}
	return resp, nil
}

// AppendRow 在指定 sheet 的最後新增一行資料
func (c *SheetsClient) AppendRow(ctx context.Context, sheetName string, values []interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := c.service.Spreadsheets.Values.Append(
		c.spreadsheetID,
		sheetName,
		valueRange,
	).ValueInputOption("RAW").Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("新增資料到 %s 失敗: %w", sheetName, err)
	}
	return nil
}

// UpdateCell 更新指定儲存格的值
func (c *SheetsClient) UpdateCell(ctx context.Context, cellRange string, value interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{{value}},
	}

	_, err := c.service.Spreadsheets.Values.Update(
		c.spreadsheetID,
		cellRange,
		valueRange,
	).ValueInputOption("RAW").Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("更新儲存格 %s 失敗: %w", cellRange, err)
	}
	return nil
}

// BatchUpdate 批次更新多個儲存格
func (c *SheetsClient) BatchUpdate(ctx context.Context, data []*sheets.ValueRange) error {
	batchUpdateRequest := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
		Data:             data,
	}

	_, err := c.service.Spreadsheets.Values.BatchUpdate(
		c.spreadsheetID,
		batchUpdateRequest,
	).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("批次更新失敗: %w", err)
	}
	return nil
}

// EnsureSheetExists 確保指定的 sheet 存在，如果不存在則建立
func (c *SheetsClient) EnsureSheetExists(ctx context.Context, sheetName string, headers []interface{}) error {
	// 檢查 sheet 是否存在
	spreadsheet, err := c.service.Spreadsheets.Get(c.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("取得 spreadsheet 資訊失敗: %w", err)
	}

	// 檢查是否已存在
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return nil // Sheet 已存在
		}
	}

	// 建立新的 sheet
	requests := []*sheets.Request{
		{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err = c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("建立 sheet %s 失敗: %w", sheetName, err)
	}

	// 新增表頭
	if len(headers) > 0 {
		err = c.AppendRow(ctx, sheetName, headers)
		if err != nil {
			return fmt.Errorf("新增表頭失敗: %w", err)
		}
	}

	return nil
}
