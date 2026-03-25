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
	} else {
		var prizes []PrizeType
		err := json.Unmarshal([]byte(CreateEvent.Prizes), &prizes)
		if err != nil {
			return c.Reply(i18nGetString("Create_noprizeset", lang))
		}
		for _, prize := range prizes {
			method := prize.ClaimRewardMethod
			switch {
			case method == "group":
			case isValidUsername(method):
			case method == "direct":
				if len(prize.Directclaim) < prize.Amount {
					return c.Reply(i18nGetString("Create_NotEnoughPrize", lang))
				}

			default:
				return c.Reply(i18nGetString("Create_NoClaimMethodSet", lang))
			}
		}
	}

	EventID := UUIDShort()

	// 加入数据库
	lottery := LotteryEventType{
		EventID:               EventID,
		LotteryDescription:    CreateEvent.LotteryDescription,
		Prizes:                CreateEvent.Prizes,
		InChannelRequired:     CreateEvent.InChannelRequired,
		InChannelRequiredLink: CreateEvent.InChannelRequiredLink,
		InGroupRequired:       CreateEvent.InGroupRequired,
		InGroupRequiredLink:   CreateEvent.InGroupRequiredLink,
		Created:               time.Now().Unix(),
		DeadlineEnabled:       CreateEvent.DeadlineEnabled,
		DrawDeadline:          CreateEvent.DrawDeadline,
		MemberCountEnabled:    CreateEvent.MemberCountEnabled,
		DrawMemberCount:       CreateEvent.DrawMemberCount,
		Usercount:             0,
		Completed:             0,
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

	btnGenerateText := tb.InlineButton{
		Unique: "genText",
		Data:   EventID,
		Text:   i18nGetString("Btn_GenInfo", lang),
	}
	markup := &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{btnGenerateText},
		},
	}

	c.Reply(message, markdownOption, markup)

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
		// 奖品
		Rewards := ""
		var prizeList []PrizeType
		if len(event.Prizes) > 0 {
			err := json.Unmarshal(event.Prizes, &prizeList)
			if err != nil {
				return c.Reply(i18nGetString("Query_getprizeerror", lang))
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

		//开奖方式
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

// GetUserID 获取自己的信息
func GetUserID(c tb.Context) error {
	lang := c.Sender().LanguageCode
	return c.Reply(fmt.Sprintf(i18nGetString("GetUserID", lang), strconv.FormatInt(c.Sender().ID, 10), lang))
}

func ProcessCallback(c tb.Context) error {
	lang := c.Sender().LanguageCode

	cb := c.Callback()
	if cb == nil {
		return nil
	}

	btnDataSplitN := strings.SplitN(strings.TrimSpace(cb.Data), "|", 2)
	if len(btnDataSplitN) != 2 {
		return c.Bot().Respond(cb, &tb.CallbackResponse{
			Text: "Invalid button format",
		})
	}

	btnType := btnDataSplitN[0]
	btnData := btnDataSplitN[1]

	switch {
	case btnType == "genText":
		GenerateText(c, btnData)
		return c.Bot().Respond(cb, &tb.CallbackResponse{
			Text: i18nGetString("Btn_GenInfo", lang),
		})
	case btnType == "claim":
		Claim(c, btnData)
		return c.Bot().Respond(cb, &tb.CallbackResponse{
			Text: i18nGetString("Btn_Claim", lang),
		})

	default:
		return c.Send("test")
	}
}

// Participate 参加抽奖
func Participate(bot *tb.Bot, cb *tb.Callback, text string) error {
	//lang := c.Sender().LanguageCode
	return nil
}

func Claim(c tb.Context, btnData string) error {
	//lang := c.Sender().LanguageCode
	return nil
}

func GenerateText(c tb.Context, btnData string) error {
	lang := c.Sender().LanguageCode
	event, _ := getLotteryInfo(btnData)

	// 奖品、领奖方式
	Rewards := "\n"
	ClaimMethod := ""
	var prizeList []PrizeType
	if len(event.Prizes) > 0 {
		err := json.Unmarshal(event.Prizes, &prizeList)
		if err != nil {
			return c.Reply(i18nGetString("Query_getprizeerror", lang))
		} else {
			for _, p := range prizeList {
				if p.IsNameHidden {
					Rewards += fmt.Sprintf(i18nGetString("Lottery_PrizeString", lang), EscapeMarkdownV2(strconv.FormatInt(int64(p.Amount), 10)), EscapeMarkdownV2(p.HiddenCap))
					ClaimMethod += "\n \\- _" + EscapeMarkdownV2(p.HiddenCap)
				} else {
					Rewards += fmt.Sprintf(i18nGetString("Lottery_PrizeString", lang), EscapeMarkdownV2(strconv.FormatInt(int64(p.Amount), 10)), EscapeMarkdownV2(p.Name))
					ClaimMethod += "\n \\- _" + EscapeMarkdownV2(p.Name)
				}
				method := p.ClaimRewardMethod
				switch {
				case method == "group":
					ClaimMethod += i18nGetString("Lottery_ClaimInGroup", lang)
				case isValidUsername(method):
					ClaimMethod += fmt.Sprintf(i18nGetString("Lottery_ClaimWithUsername", lang), method)
				case method == "direct":
					ClaimMethod += i18nGetString("Lottery_ClaimInBot", lang)
				}
			}
		}
	}

	//参与条件
	condition_participate := ""
	if event.InChannelRequired != 0 {
		condition_participate += fmt.Sprintf(i18nGetString("Lottery_CondiParitChan", lang), event.InChannelRequiredLink)
	}
	if event.InGroupRequired != 0 {
		condition_participate += fmt.Sprintf(i18nGetString("Lottery_CondiParitGrp", lang), event.InGroupRequiredLink)
	}
	if condition_participate == "" {
		condition_participate = i18nGetString("Lottery_CondiParitNil", lang)
	}

	// 开奖方式
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

	//lottery := fmt.Sprintf(i18nGetString("a", lang), event.LotteryDescription)
	lottery := fmt.Sprintf(i18nGetString("Lottery_String", lang), EscapeMarkdownV2(event.LotteryDescription), condition_participate, Rewards, DrawMethod, ClaimMethod)

	btnParticipate := tb.InlineButton{
		Unique: "genText",
		URL:    fmt.Sprintf("https://t.me/%s?start=%s", bot.Me.Username, ("Parti" + event.EventID)),
		Text:   i18nGetString("Btn_Participate", lang),
	}
	markup := &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{btnParticipate},
		},
	}

	return c.Reply(lottery, markdownOption, markup)
}

func GenerateTextCMD(c tb.Context) error {
	lang := c.Sender().LanguageCode
	values := strings.SplitN(c.Text(), " ", 2)
	if len(values) != 2 {
		c.Reply(i18nGetString("Query_contenterror", lang), markdownOption)
		return nil
	}
	return GenerateText(c, values[1])
}
