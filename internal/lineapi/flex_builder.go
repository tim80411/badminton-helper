package lineapi

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/tim80411/badminton-helper/internal/domain"
)

// ActivitySettings 活動設定
type ActivitySettings struct {
	Place string
	Time  string
	Price string
	Quota string
}

// BuildActivityFlexMessage 建立活動報名 Flex Message
func BuildActivityFlexMessage(activityID string, settings ActivitySettings) messaging_api.MessageInterface {
	if activityID == "" {
		// 生成新的活動 ID
		now := time.Now()
		activityID = fmt.Sprintf("%s-%s", now.Format("20060102"), uuid.New().String())
	}

	bubble := &messaging_api.FlexBubble{
		Hero: &messaging_api.FlexImage{
			Url:         "https://i.imgur.com/sUrFITq.png",
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  messaging_api.FlexImageASPECT_MODE_COVER,
			Action: &messaging_api.UriAction{
				Uri: "https://line.me/",
			},
		},
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   "phah kiû lah 🏸",
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Size:   "xl",
				},
				&messaging_api.FlexBox{
					Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
					Margin:  "lg",
					Spacing: "sm",
					Contents: []messaging_api.FlexComponentInterface{
						buildInfoRow("Place", settings.Place),
						buildInfoRow("Time", settings.Time),
						buildInfoRow("Price", settings.Price),
						buildInfoRow("Quota", settings.Quota),
						&messaging_api.FlexText{
							Text:   "＊想報名幾位就按幾次按鈕，例如：報名臨打 2 位，就按兩次「臨打」",
							Color:  "#666666",
							Size:   "sm",
							Wrap:   true,
							Margin: "md",
						},
						&messaging_api.FlexText{
							Text:  "⚠️ 臨時有事需取消請提前告知",
							Color: "#EA0000",
							Size:  "sm",
						},
					},
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing: "sm",
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_PRIMARY,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.PostbackAction{
						Label:       "半年繳 / 季繳",
						Data:        fmt.Sprintf("activityId=%s,message=季繳+1", activityID),
						DisplayText: "季繳 +1",
					},
				},
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_SECONDARY,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.PostbackAction{
						Label:       "臨打",
						Data:        fmt.Sprintf("activityId=%s,message=臨打+1", activityID),
						DisplayText: "臨打 +1",
					},
				},
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_LINK,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.PostbackAction{
						Label:       "本週名單",
						Data:        fmt.Sprintf("action=weekly_list,activityId=%s", activityID),
						DisplayText: "查看本週名單",
					},
					Color: "#17c950",
				},
				&messaging_api.FlexBox{
					Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
					Contents: []messaging_api.FlexComponentInterface{},
					Margin:   "sm",
				},
			},
			Flex: 0,
		},
	}

	return &messaging_api.FlexMessage{
		AltText:  "羽球活動報名",
		Contents: bubble,
	}
}

// BuildWeeklyListFlexMessage 建立週名單 Flex Message
func BuildWeeklyListFlexMessage(seasonTicket, casualPlay []*domain.Registration, isSaturday bool) messaging_api.MessageInterface {
	contents := []messaging_api.FlexComponentInterface{
		&messaging_api.FlexText{
			Text:   "本週報名名單 📋",
			Weight: messaging_api.FlexTextWEIGHT_BOLD,
			Size:   "xl",
			Margin: "md",
		},
	}

	// 季繳名單
	if len(seasonTicket) > 0 {
		contents = append(contents, &messaging_api.FlexText{
			Text:   "季繳：",
			Weight: messaging_api.FlexTextWEIGHT_BOLD,
			Size:   "md",
			Color:  "#333333",
			Margin: "lg",
		})

		for i, reg := range seasonTicket {
			displayText := fmt.Sprintf("%d. %s", i+1, reg.UserName)
			if isSaturday && reg.CourtAssignment != "" {
				displayText += fmt.Sprintf("：%s", reg.CourtAssignment)
			}

			contents = append(contents, &messaging_api.FlexText{
				Text:   displayText,
				Size:   "sm",
				Color:  "#666666",
				Margin: "xs",
			})
		}
	}

	// 臨打名單
	if len(casualPlay) > 0 {
		startIndex := len(seasonTicket) + 1

		contents = append(contents, &messaging_api.FlexText{
			Text:   "臨打：",
			Weight: messaging_api.FlexTextWEIGHT_BOLD,
			Size:   "md",
			Color:  "#333333",
			Margin: "lg",
		})

		for i, reg := range casualPlay {
			displayText := fmt.Sprintf("%d. %s", startIndex+i, reg.UserName)
			if isSaturday && reg.CourtAssignment != "" {
				displayText += fmt.Sprintf("：%s", reg.CourtAssignment)
			}

			contents = append(contents, &messaging_api.FlexText{
				Text:   displayText,
				Size:   "sm",
				Color:  "#666666",
				Margin: "xs",
			})
		}
	}

	// 如果沒有任何報名
	if len(seasonTicket) == 0 && len(casualPlay) == 0 {
		contents = append(contents, &messaging_api.FlexText{
			Text:   "本週還沒有人報名唷！快來報名吧 🏸",
			Size:   "md",
			Color:  "#999999",
			Margin: "lg",
			Align:  messaging_api.FlexTextALIGN_CENTER,
		})
	}

	// 添加提示文字
	if !isSaturday {
		contents = append(contents, &messaging_api.FlexText{
			Text:   "💡 場地分配將在週六公布",
			Size:   "xs",
			Color:  "#999999",
			Margin: "lg",
			Align:  messaging_api.FlexTextALIGN_CENTER,
		})
	}

	bubble := &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: contents,
			Spacing:  "sm",
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing: "sm",
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:  "想報名的話請點選上面的按鈕唷！",
					Size:  "xs",
					Color: "#aaaaaa",
					Align: messaging_api.FlexTextALIGN_CENTER,
				},
			},
		},
	}

	return &messaging_api.FlexMessage{
		AltText:  "本週羽球報名名單",
		Contents: bubble,
	}
}

// BuildRegistrationConfirmMessage 建立報名確認 Flex Message
func BuildRegistrationConfirmMessage(userName, registrationType, registrationID string) messaging_api.MessageInterface {
	bubble := &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexBox{
					Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
					Margin:  "lg",
					Spacing: "sm",
					Contents: []messaging_api.FlexComponentInterface{
						&messaging_api.FlexText{
							Text:   fmt.Sprintf("已記錄 %s 的 %s 報名！想變卦請點按鈕退場 👇", userName, registrationType),
							Color:  "#272727",
							Size:   "md",
							Wrap:   true,
							Margin: "xs",
						},
					},
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing: "sm",
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_PRIMARY,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.PostbackAction{
						Label:       "取消報名",
						Data:        fmt.Sprintf("action=cancel,registrationId=%s,userName=%s", registrationID, userName),
						DisplayText: fmt.Sprintf("%s 申請取消報名！", userName),
					},
					Color: "#CE0000",
				},
				&messaging_api.FlexBox{
					Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
					Contents: []messaging_api.FlexComponentInterface{},
					Margin:   "sm",
				},
			},
			Flex: 0,
		},
	}

	return &messaging_api.FlexMessage{
		AltText:  fmt.Sprintf("%s 的 %s 報名確認", userName, registrationType),
		Contents: bubble,
	}
}

// buildInfoRow 建立資訊列（Place/Time/Price/Quota）
func buildInfoRow(label, value string) messaging_api.FlexComponentInterface {
	return &messaging_api.FlexBox{
		Layout:  messaging_api.FlexBoxLAYOUT_BASELINE,
		Spacing: "sm",
		Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{
				Text:  label,
				Color: "#aaaaaa",
				Size:  "sm",
				Flex:  1,
			},
			&messaging_api.FlexText{
				Text:  value,
				Wrap:  true,
				Color: "#666666",
				Size:  "sm",
				Flex:  5,
			},
		},
	}
}
