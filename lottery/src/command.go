package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/telebot.v4"
)

var markdownOption = &tb.SendOptions{
	ParseMode: "MarkdownV2",
}

var loc *time.Location

// CreateLottery 创建抽奖活动
func CreateLottery(c tb.Context) error {
	lang := c.Sender().LanguageCode

	// 仅允许管理员+私聊
	if c.Chat().Type != tb.ChatPrivate {
		return nil
	}
	if !isAdmin(c.Sender().ID) {
		c.Reply(i18nGetString("nopermission", lang))
		return nil
	}

	// 读取创建信息
	values := strings.SplitN(c.Text(), " ", 2)
	if len(values) != 2 {
		c.Reply(i18nGetString("Create_contenterror", lang), markdownOption)
		return nil
	}
	var CreateEvent CreateEventType
	err := json.Unmarshal([]byte(values[1]), &CreateEvent)
	if err != nil {
		c.Reply(i18nGetString("Create_syntaxerror", lang), markdownOption)
		return nil
	}

	if CreateEvent.DrawDeadline == 0 {
		CreateEvent.DeadlineEnabled = false
	}
	if CreateEvent.DrawMemberCount == 0 {
		CreateEvent.MemberCountEnabled = false
	}
	if len(CreateEvent.Prizes) == 0 {
		c.Reply(i18nGetString("Create_noprizeset", lang))
		return nil
	}

	EventID := UUIDShort()

	// 加入数据库
	lottery := LotteryEventType{
		EventID:            EventID,
		LotteryDescription: CreateEvent.LotteryDescription,
		Prizes:             CreateEvent.Prizes,
		InChannelRequired:  CreateEvent.InChannelRequired,
		InGroupRequired:    CreateEvent.InGroupRequired,
		Created:            time.Now().Unix(),
		DeadlineEnabled:    CreateEvent.DeadlineEnabled,
		DrawDeadline:       CreateEvent.DrawDeadline,
		MemberCountEnabled: CreateEvent.MemberCountEnabled,
		DrawMemberCount:    CreateEvent.DrawMemberCount,
		Usercount:          0,
		Completed:          0,
		Owner: UserinfoType{
			ID:          c.Sender().ID,
			Displayname: (c.Sender().FirstName + " " + c.Sender().LastName),
			Username:    c.Sender().Username,
		},
		LuckyUsers: []UserinfoType{},
	}

	if err := DB.Create(&lottery).Error; err != nil {
		log.Fatal(err)
	}
	message := fmt.Sprintf(i18nGetString("Create_success", lang), EscapeMarkdownV2(EventID))

	c.Reply(message, markdownOption)

	return nil
}

// Help 显示帮助信息
func Help(c tb.Context) error {
	lang := c.Sender().LanguageCode
	// 仅允许私聊
	if c.Chat().Type != tb.ChatPrivate {
		return nil
	}

	if isAdmin(c.Sender().ID) {
		return c.Reply(fmt.Sprintf(i18nGetString("Help", lang), i18nGetString("Admin_commands", lang)))
	} else {
		return c.Reply(fmt.Sprintf(i18nGetString("Help", lang), i18nGetString("User_commands", lang)))
	}

}

// About 显示关于信息
func About(c tb.Context) error {
	lang := c.Sender().LanguageCode
	// 仅允许私聊
	if c.Chat().Type != tb.ChatPrivate {
		return nil
	}

	return c.Reply(fmt.Sprintf(i18nGetString("About", lang), conf.Bot.OSS))

}

// QueryLottery 查询抽奖信息
func QueryLottery(c tb.Context) error {
	lang := c.Sender().LanguageCode
	// 仅允许管理员私聊
	if c.Chat().Type != tb.ChatPrivate {
		return nil
	}
	if !isAdmin(c.Sender().ID) {
		c.Reply(i18nGetString("nopermission", lang))
		return nil
	}

	// 读取抽奖信息
	values := strings.Split(c.Text(), " ")
	if len(values) != 2 {
		c.Reply(i18nGetString("Query_contenterror", lang), markdownOption)
		return nil
	}

	event, err := getLotteryInfo(values[1])
	if err != nil {
		c.Reply(i18nGetString("Query_searcherror", lang))
	} else {
		Rewards := ""
		var prizeList []PrizeType
		if len(event.Prizes) > 0 {
			err := json.Unmarshal(event.Prizes, &prizeList)
			if err != nil {
				fmt.Println(i18nGetString("Query_getprizeerror", lang), err)
				return nil
			} else {
				for _, p := range prizeList {
					HiddenRewards := "\n"
					if p.IsNameHidden {
						HiddenRewards = fmt.Sprintf(i18nGetString("Query_HiddenRewards", lang), EscapeMarkdownV2(p.HiddenCap))
					}
					Rewards += fmt.Sprintf(i18nGetString("Query_DisplayRewards", lang), EscapeMarkdownV2(strconv.FormatInt(int64(p.Amount), 10)), EscapeMarkdownV2(p.Name), HiddenRewards)
				}
			}
		}
		DrawMethod := ""
		if event.MemberCountEnabled {
			DrawMethod += fmt.Sprintf(i18nGetString("Query_DrawWhenMemberCount", lang), EscapeMarkdownV2(strconv.FormatInt(event.DrawMemberCount, 10)))
		}
		if event.DeadlineEnabled {
			DrawMethod += fmt.Sprintf(i18nGetString("Query_DrawWhenDeadline", lang), EscapeMarkdownV2(time.Unix(event.DrawDeadline, 0).In(loc).Format("2006-01-02 15:04")))
		}
		if DrawMethod == "" {
			DrawMethod = i18nGetString("Query_DrawManually", lang)
		}

		message := fmt.Sprintf(i18nGetString("Query_Eventinfo", lang),
			EscapeMarkdownV2(event.EventID),
			EscapeMarkdownV2(event.LotteryDescription),
			EscapeMarkdownV2(event.Owner.Displayname),
			EscapeMarkdownV2(strconv.FormatInt(event.Owner.ID, 10)),
			EscapeMarkdownV2(event.Owner.Username),
			Rewards,
			EscapeMarkdownV2(strconv.FormatInt(event.Usercount, 10)),
			DrawMethod,
			map[bool]string{true: i18nGetString("True", lang), false: i18nGetString("False", lang)}[event.isCompleted])
		c.Reply(message, markdownOption)
	}
	return nil
}

// ListLottery 列出现有抽奖信息
func ListLottery(c tb.Context) error {
	lang := c.Sender().LanguageCode
	// 仅允许管理员+私聊
	if c.Chat().Type != tb.ChatPrivate {
		return nil
	}
	if !isAdmin(c.Sender().ID) {
		c.Reply(i18nGetString("nopermission", lang))
		return nil
	}

	events, err := listLottery()
	if err != nil {
		c.Reply(i18nGetString("List_Failed", lang) + err.Error())
	} else {
		message := i18nGetString("List_MessageHead", lang)
		if len(events) > 0 {
			for _, e := range events {
				if len(message) < 3000 {
					message += fmt.Sprintf(i18nGetString("List_MessageBody", lang),
						EscapeMarkdownV2(e.EventID), EscapeMarkdownV2(e.LotteryDescription), EscapeMarkdownV2(e.Owner.Displayname), EscapeMarkdownV2(strconv.FormatInt(e.Owner.ID, 10)),
						EscapeMarkdownV2(e.Owner.Username), EscapeMarkdownV2(strconv.FormatInt(e.Usercount, 10)), map[bool]string{true: i18nGetString("True", lang), false: i18nGetString("False", lang)}[e.isCompleted])
				} else {
					c.Reply(message, markdownOption)
					message = fmt.Sprintf(i18nGetString("List_MessageBodyWithHead", lang),
						EscapeMarkdownV2(e.EventID), EscapeMarkdownV2(e.LotteryDescription), EscapeMarkdownV2(e.Owner.Displayname), EscapeMarkdownV2(strconv.FormatInt(e.Owner.ID, 10)),
						EscapeMarkdownV2(e.Owner.Username), EscapeMarkdownV2(strconv.FormatInt(e.Usercount, 10)), map[bool]string{true: i18nGetString("True", lang), false: i18nGetString("False", lang)}[e.isCompleted])
				}
			}
			c.Reply(message, markdownOption)
		} else {
			c.Reply(i18nGetString("List_Nodata", lang))
		}
	}
	return nil
}

// DeleteLottery 删除抽奖
func DeleteLottery(c tb.Context) error {
	lang := c.Sender().LanguageCode
	// 仅允许管理员+私聊
	if c.Chat().Type != tb.ChatPrivate {
		return nil
	}
	if !isAdmin(c.Sender().ID) {
		return nil
	}

	values := strings.SplitN(c.Text(), " ", 2)
	if len(values) != 2 {
		c.Reply(i18nGetString("Query_contenterror", lang), markdownOption)
		return nil
	}

	event, err := getLotteryInfo(values[1])
	if err != nil {
		c.Reply(i18nGetString("Query_searcherror", lang))
		return nil
	} else if c.Sender().ID != event.Owner.ID {
		if !isSuperAdmin(c.Sender().ID) {
			return nil
		}
	}
	err = deleteEvent(event.ID)
	if err != nil {
		return c.Reply(fmt.Sprintf(i18nGetString("Delete_error", lang), values[1]))
	}
	return c.Reply(fmt.Sprintf(i18nGetString("Delete_Success", lang), values[1]))
}

func GetUserID(c tb.Context) error {
	lang := c.Sender().LanguageCode
	return c.Reply(fmt.Sprintf(i18nGetString("GetUserID", lang), c.Sender().ID, lang))
}
