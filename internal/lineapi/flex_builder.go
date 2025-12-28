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
			Size:        ptr(messaging_api.FlexImageSize("full")),
			AspectRatio: ptr(messaging_api.FlexImageAspectRatio("20:13")),
			AspectMode:  ptr(messaging_api.FlexImageAspectMode("cover")),
			Action: &messaging_api.UriAction{
				Uri: "https://line.me/",
			},
		},
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLayout("vertical"),
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   "phah kiû lah 🏸",
					Weight: ptr(messaging_api.FlexTextWeight("bold")),
					Size:   ptr(messaging_api.FlexTextSize("xl")),
				},
				&messaging_api.FlexBox{
					Layout:  messaging_api.FlexBoxLayout("vertical"),
					Margin:  ptr(messaging_api.FlexComponentMargin("lg")),
					Spacing: ptr(messaging_api.FlexComponentSpacing("sm")),
					Contents: []messaging_api.FlexComponentInterface{
						buildInfoRow("Place", settings.Place),
						buildInfoRow("Time", settings.Time),
						buildInfoRow("Price", settings.Price),
						buildInfoRow("Quota", settings.Quota),
						&messaging_api.FlexText{
							Text:   "＊想報名幾位就按幾次按鈕，例如：報名臨打 2 位，就按兩次「臨打」",
							Color:  ptr("#666666"),
							Size:   ptr(messaging_api.FlexTextSize("sm")),
							Wrap:   ptr(true),
							Margin: ptr(messaging_api.FlexComponentMargin("md")),
						},
						&messaging_api.FlexText{
							Text:  "⚠️ 臨時有事需取消請提前告知",
							Color: ptr("#EA0000"),
			Size:   ptr(messaging_api.FlexTextSize("sm")),
						},
					},
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLayout("vertical"),
			Spacing: ptr(messaging_api.FlexComponentSpacing("sm")),
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexButton{
					Style:  ptr(messaging_api.FlexButtonStyle("primary")),
					Height: ptr(messaging_api.FlexButtonHeight("sm")),
					Action: &messaging_api.PostbackAction{
						Label:       "半年繳 / 季繳",
						Data:        fmt.Sprintf("activityId=%s,message=季繳+1", activityID),
						DisplayText: "季繳 +1",
					},
				},
				&messaging_api.FlexButton{
					Style:  ptr(messaging_api.FlexButtonStyle("secondary")),
					Height: ptr(messaging_api.FlexButtonHeight("sm")),
					Action: &messaging_api.PostbackAction{
						Label:       "臨打",
						Data:        fmt.Sprintf("activityId=%s,message=臨打+1", activityID),
						DisplayText: "臨打 +1",
					},
				},
				&messaging_api.FlexButton{
					Style:  ptr(messaging_api.FlexButtonStyle("link")),
					Height: ptr(messaging_api.FlexButtonHeight("sm")),
					Action: &messaging_api.PostbackAction{
						Label:       "本週名單",
						Data:        fmt.Sprintf("action=weekly_list,activityId=%s", activityID),
						DisplayText: "查看本週名單",
					},
					Color: ptr("#17c950"),
				},
				&messaging_api.FlexBox{
					Layout:   messaging_api.FlexBoxLayout("vertical"),
					Contents: []messaging_api.FlexComponentInterface{},
					Margin:   ptr(messaging_api.FlexComponentMargin("sm")),
				},
			},
			Flex: ptr(0),
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
			Weight: ptr(messaging_api.FlexTextWeight("bold")),
			Size:   ptr(messaging_api.FlexTextSize("xl")),
			Margin: ptr(messaging_api.FlexComponentMargin("md")),
		},
	}

	// 季繳名單
	if len(seasonTicket) > 0 {
		contents = append(contents, &messaging_api.FlexText{
			Text:   "季繳：",
			Weight: ptr(messaging_api.FlexTextWeight("bold")),
			Size:   ptr(messaging_api.FlexTextSize("md")),
			Color:  ptr("#333333"),
			Margin: ptr(messaging_api.FlexComponentMargin("lg")),
		})

		for i, reg := range seasonTicket {
			displayText := fmt.Sprintf("%d. %s", i+1, reg.UserName)
			if isSaturday && reg.CourtAssignment != "" {
				displayText += fmt.Sprintf("：%s", reg.CourtAssignment)
			}

			contents = append(contents, &messaging_api.FlexText{
				Text:   displayText,
				Size:   ptr(messaging_api.FlexTextSize("sm")),
				Color:  ptr("#666666"),
				Margin: ptr(messaging_api.FlexComponentMargin("xs")),
			})
		}
	}

	// 臨打名單
	if len(casualPlay) > 0 {
		startIndex := len(seasonTicket) + 1

		contents = append(contents, &messaging_api.FlexText{
			Text:   "臨打：",
			Weight: ptr(messaging_api.FlexTextWeight("bold")),
			Size:   ptr(messaging_api.FlexTextSize("md")),
			Color:  ptr("#333333"),
			Margin: ptr(messaging_api.FlexComponentMargin("lg")),
		})

		for i, reg := range casualPlay {
			displayText := fmt.Sprintf("%d. %s", startIndex+i, reg.UserName)
			if isSaturday && reg.CourtAssignment != "" {
				displayText += fmt.Sprintf("：%s", reg.CourtAssignment)
			}

			contents = append(contents, &messaging_api.FlexText{
				Text:   displayText,
				Size:   ptr(messaging_api.FlexTextSize("sm")),
				Color:  ptr("#666666"),
				Margin: ptr(messaging_api.FlexComponentMargin("xs")),
			})
		}
	}

	// 如果沒有任何報名
	if len(seasonTicket) == 0 && len(casualPlay) == 0 {
		contents = append(contents, &messaging_api.FlexText{
			Text:   "本週還沒有人報名唷！快來報名吧 🏸",
			Size:   ptr(messaging_api.FlexTextSize("md")),
			Color:  ptr("#999999"),
			Margin: ptr(messaging_api.FlexComponentMargin("lg")),
			Align:  ptr(messaging_api.FlexTextAlign("center")),
		})
	}

	// 添加提示文字
	if !isSaturday {
		contents = append(contents, &messaging_api.FlexText{
			Text:   "💡 場地分配將在週六公布",
			Size:   ptr(messaging_api.FlexTextSize("xs")),
			Color:  ptr("#999999"),
			Margin: ptr(messaging_api.FlexComponentMargin("lg")),
			Align:  ptr(messaging_api.FlexTextAlign("center")),
		})
	}

	bubble := &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout:   messaging_api.FlexBoxLayout("vertical"),
			Contents: contents,
			Spacing:  ptr(messaging_api.FlexComponentSpacing("sm")),
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLayout("vertical"),
			Spacing: ptr(messaging_api.FlexComponentSpacing("sm")),
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:  "想報名的話請點選上面的按鈕唷！",
					Size:  ptr(messaging_api.FlexTextSize("xs")),
					Color: ptr("#aaaaaa"),
					Align: ptr(messaging_api.FlexTextAlign("center")),
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
			Layout: messaging_api.FlexBoxLayout("vertical"),
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexBox{
					Layout:  messaging_api.FlexBoxLayout("vertical"),
					Margin:  ptr(messaging_api.FlexComponentMargin("lg")),
					Spacing: ptr(messaging_api.FlexComponentSpacing("sm")),
					Contents: []messaging_api.FlexComponentInterface{
						&messaging_api.FlexText{
							Text:   fmt.Sprintf("已記錄 %s 的 %s 報名！想變卦請點按鈕退場 👇", userName, registrationType),
							Color:  ptr("#272727"),
							Size:   ptr(messaging_api.FlexTextSize("md")),
							Wrap:   ptr(true),
							Margin: ptr(messaging_api.FlexComponentMargin("xs")),
						},
					},
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLayout("vertical"),
			Spacing: ptr(messaging_api.FlexComponentSpacing("sm")),
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexButton{
					Style:  ptr(messaging_api.FlexButtonStyle("primary")),
					Height: ptr(messaging_api.FlexButtonHeight("sm")),
					Action: &messaging_api.PostbackAction{
						Label:       "取消報名",
						Data:        fmt.Sprintf("action=cancel,registrationId=%s,userName=%s", registrationID, userName),
						DisplayText: fmt.Sprintf("%s 申請取消報名！", userName),
					},
					Color: ptr("#CE0000"),
				},
				&messaging_api.FlexBox{
					Layout:   messaging_api.FlexBoxLayout("vertical"),
					Contents: []messaging_api.FlexComponentInterface{},
					Margin:   ptr(messaging_api.FlexComponentMargin("sm")),
				},
			},
			Flex: ptr(0),
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
		Layout:  messaging_api.FlexBoxLayout("baseline"),
		Spacing: ptr(messaging_api.FlexComponentSpacing("sm")),
		Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{
				Text:  label,
				Color: ptr("#aaaaaa"),
				Size:  ptr(messaging_api.FlexTextSize("sm")),
				Flex:  ptr(1),
			},
			&messaging_api.FlexText{
				Text:  value,
				Wrap:  ptr(true),
				Color: ptr("#666666"),
				Size:  ptr(messaging_api.FlexTextSize("sm")),
				Flex:  ptr(5),
			},
		},
	}
}
