package main

import (
	"encoding/json"
	"os"
)

var i18nMap map[string]map[string]string

// 载入 i18n 配置
func LoadI18n() error {
	data, err := os.ReadFile("../assets/strings.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &i18nMap)
	if err != nil {
		return err
	}

	return nil
}

// 获取 i18n 字段
func i18nGetString(key string, lang string) string {
	if m, ok := i18nMap[lang]; ok {
		if val, ok := m[key]; ok {
			return val
		}
	}
	if m, ok := i18nMap[conf.Bot.FallbackLang]; ok {
		if val, ok := m[key]; ok {
			return val
		}
	}
	return key
}
