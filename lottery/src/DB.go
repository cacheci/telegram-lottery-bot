package main

import (
	"fmt"
	"log"
	"time"

	tb "gopkg.in/telebot.v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(conf.Bot.Sqlite), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	if err := db.AutoMigrate(&UserinfoType{}, &LotteryEventType{}, &ParticipationType{}); err != nil {
		log.Fatal("数据表加载失败:", err)
	}

	return db
}

func getLotteryInfo(EventID string) (*LotteryEventType, error) {
	var result LotteryEventType
	RealID, err := feistelGetID(EventID)
	if err != nil {
		return nil, err
	}
	err = DB.Preload("LuckyUsers").Preload("Owner").First(&result, RealID).Error
	if err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func listLottery() ([]LotteryEventType, error) {
	var result []LotteryEventType
	if err := DB.Preload("LuckyUsers").Preload("Owner").Find(&result).Error; err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func deleteEvent(id uint) error {
	return DB.Delete(&LotteryEventType{}, id).Error
}

func createEvent(lottery *LotteryEventType) error {
	if err := DB.Create(lottery).Error; err != nil {
		return err
	}
	return nil
}

func JoinEvent(c tb.Context, user UserinfoType, eventID string) error {
	lang := c.Sender().LanguageCode

	event, err := getLotteryInfo(eventID)
	if err != nil {
		return c.Send(i18nGetString("Query_searcherror", lang))
	}

	// 查看抽奖是否开启
	if event.isCompleted {
		return c.Send(i18nGetString("JoinEvent_EventInactive", lang))
	}

	// 读取设置，判断是否满足加入条件
	conditionstr := ""
	if event.InChannelRequired != 0 {
		// 检查是否是频道成员
		channel := &tb.Chat{ID: event.InChannelRequired}
		isMember, err := c.Bot().ChatMemberOf(channel, c.Sender())
		if err != nil {
			return c.Send("err")
		}

		switch isMember.Role {
		case tb.Member, tb.Administrator, tb.Creator:
		default:
			conditionstr += fmt.Sprintf(i18nGetString("JoinEvent_NotInChannel", lang), event.InChannelRequiredLink)
		}

	}

	if event.InGroupRequired != 0 {
		// 检查是否是群组成员
		group := &tb.Chat{ID: event.InGroupRequired}
		isMember, err := c.Bot().ChatMemberOf(group, c.Sender())
		if err != nil {
			return c.Reply(i18nGetString("BOT_INTERNAL_ERROR", lang))
		}

		switch isMember.Role {
		case tb.Member, tb.Administrator, tb.Creator:
		default:
			conditionstr += fmt.Sprintf(i18nGetString("JoinEvent_NotInGroup", lang), event.InGroupRequiredLink)
		}

	}

	if conditionstr != "" {
		return c.Reply(conditionstr)
	}

	participation := ParticipationType{
		EventID:  event.ID,
		Userid:   user.ID,
		JoinedAt: time.Now().Unix(),
		User:     user,
		UserFKID: user.ID,
	}

	result := DB.FirstOrCreate(&participation, ParticipationType{
		EventID:  event.ID,
		Userid:   user.ID,
		JoinedAt: time.Now().Unix(),
		User:     user,
		UserFKID: user.ID,
	})
	if result.Error != nil {
		return c.Reply(i18nGetString("BOT_INTERNAL_ERROR", lang))
	}

	if result.RowsAffected == 0 {
		return c.Reply(i18nGetString("JoinEvent_AlreadyJoined", lang))
	} else {
		if err := DB.Model(&event).UpdateColumn("Usercount", gorm.Expr("Usercount + ?", 1)).Error; err != nil {
			return c.Reply(i18nGetString("BOT_INTERNAL_ERROR", lang))
		} else {
			JoinUUID, err := getJoinUUID(event.EventID(), user.ID)
			if err != nil {
				return c.Reply(i18nGetString("BOT_INTERNAL_ERROR", lang))
			} else {
				return c.Reply(fmt.Sprintf(i18nGetString("JoinEvent_Success", lang), EscapeMarkdownV2(JoinUUID)), markdownOption)
			}
		}
	}
}
