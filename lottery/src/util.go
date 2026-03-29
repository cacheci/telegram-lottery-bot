package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/telebot.v4"
	tb "gopkg.in/telebot.v4"
)

func isAdmin(id int64) bool {
	for _, adminid := range conf.Admin.Admin {
		if adminid == id {
			return true
		}
	}
	return false
}

func isSuperAdmin(id int64) bool {
	for _, sadminid := range conf.Admin.Superadmin {
		if sadminid == id {
			return true
		}
	}
	return false
}

func EscapeMarkdownV2(text string) string {
	special := "_*[]()~`>#+-=|{}.!"
	for _, ch := range special {
		text = strings.ReplaceAll(text, string(ch), "\\"+string(ch))
	}
	return text
}

func SafetyDisplaynameInput(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch {
		// Zalgo
		case unicode.Is(unicode.Mn, r): // Nonspacing Mark
			return -1
		case unicode.Is(unicode.Me, r): // Enclosing Mark
			return -1

		// 控制符号
		case r >= 0x200B && r <= 0x200F:
			return -1
		case r >= 0x202A && r <= 0x202E:
			return -1
		case r >= 0x2060 && r <= 0x206F:
			return -1
		case r == 0xFEFF:
			return -1
		case r <= 0x1F:
			return -1
		case r == 0x7F:
			return -1
		case r >= 0x80 && r <= 0x9F:
			return -1

		default:
			return r
		}
	}, name)
	return strings.TrimSpace(cleaned)
}

func isValidUsername(username string) bool {
	matched, _ := regexp.MatchString(`^@[a-zA-Z][a-zA-Z0-9_]*$`, username)
	return matched
}

// Event.EventID 用于直接获取字符串ID
func (t LotteryEventType) EventID() string {
	x := uint32(t.ID)
	l := uint16(x >> 16)
	r := uint16(x & 0xFFFF)
	key := conf.Bot.key

	for i := 0; i < 4; i++ {
		f := uint16((uint32(r)*key + uint32(i)*0x9e37) & 0xFFFF)
		l, r = r, l^f
	}

	return fmt.Sprintf("%08x", (uint32(l)<<16)|uint32(r))
}

// feistelGetID 用于将 EventID 转回 ID
func feistelGetID(EventID string) (uint, error) {
	if len(EventID) != 8 {
		return 0, fmt.Errorf("Bad EventID input: %s", EventID)
	}

	v, err := strconv.ParseUint(EventID, 16, 32)
	if err != nil {
		log.Printf("Bad EventID input")
		return 0, fmt.Errorf("Bad EventID input: %s", EventID)
	}

	x := uint32(v)
	l := uint16(x >> 16)
	r := uint16(x & 0xFFFF)
	key := conf.Bot.key

	for i := 3; i >= 0; i-- {
		f := uint16((uint32(l)*key + uint32(i)*0x9e37) & 0xFFFF)
		l, r = r^f, l
	}

	return uint((uint32(l) << 16) | uint32(r)), nil
}

func getJoinUUID(EventID string, userid int64) (string, error) {

	if len(EventID) != 8 {
		return "", fmt.Errorf("Bad EventID input: %s", EventID)
	}
	eventID, err := strconv.ParseUint(EventID, 16, 32)
	if err != nil {
		return "", fmt.Errorf("Bad EventID input: %s", err)
	}

	d := make([]byte, 16)
	binary.BigEndian.PutUint32(d[0:4], uint32(eventID))
	binary.BigEndian.PutUint64(d[4:12], uint64(userid))
	binary.BigEndian.PutUint32(d[12:16], 0)
	checksum := crc32.ChecksumIEEE(d)
	binary.BigEndian.PutUint32(d[12:16], checksum)

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		d[0:4], d[4:6], d[6:8], d[8:10], d[10:16]), nil
}

func isMemberstatReadable(c tb.Context, chatID int64) bool {
	lang := c.Sender().LanguageCode

	chat := &tb.Chat{ID: chatID}
	SeifID, _ := strconv.ParseInt(strings.Split(conf.Bot.Token[:1], ":")[0], 10, 64)
	member, err := bot.ChatMemberOf(chat, &telebot.User{ID: SeifID})
	if err != nil {
		c.Reply(i18nGetString("BOT_INTERNAL_ERROR", lang))
		return false
	}

	switch member.Role {
	case tb.Creator, tb.Administrator, tb.Member:
		// Bot 是群主、管理员或普通成员 → 可读
		return true
	case tb.Restricted, tb.Left, tb.Kicked:
		return false
	default:
		return false
	}
}
