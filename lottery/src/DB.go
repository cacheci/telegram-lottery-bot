package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(conf.Bot.Sqlite), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	if err := db.AutoMigrate(&UserinfoType{}, &LotteryEventType{}, &ParticipantType{}); err != nil {
		log.Fatal("数据表加载失败:", err)
	}

	return db
}

func getLotteryInfo(EventID string) (*LotteryEventType, error) {
	var result LotteryEventType
	err := DB.Preload("LuckyUsers").Preload("Owner").Where("event_id = ?", EventID).First(&result).Error
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
