package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
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

func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func UUIDShort() string {
	return fmt.Sprintf("%s-%s", RandomHex(4), RandomHex(2))
}

func UUIDLong(ShortID string) string {
	if ShortID == "" {
		ShortID = UUIDShort()
	}
	return fmt.Sprintf("%s-%s-%s-%s", ShortID, RandomHex(2), RandomHex(2), RandomHex(6))
}

func EscapeMarkdownV2(text string) string {
	special := "_*[]()~`>#+-=|{}.!"
	for _, ch := range special {
		text = strings.ReplaceAll(text, string(ch), "\\"+string(ch))
	}
	return text
}

func isValidUsername(username string) bool {
	matched, _ := regexp.MatchString(`^@[a-zA-Z][a-zA-Z0-9_]*$`, username)
	return matched
}
