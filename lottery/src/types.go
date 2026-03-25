package main

import (
	"gorm.io/datatypes"
)

const TableLottery = "lotteries"
const TableLuckyUser = "lucky_users"
const TablePartner = "partners"
const LotteryPrefix = "lottery:"
const DescPrefix = "descriptions"

// BotConfig 机器人配置
type BotConfig struct {
	HTTP struct {
		Proxy   string `yaml:"proxy"`
		Timeout int8   `yaml:"timeout"`
		Api     string `yaml:"tbapi"`
	} `yaml:"http"`
	Bot struct {
		Token        string `yaml:"token"`
		Poller       int8   `yaml:"poller"`
		Sqlite       string `yaml:"DB"`
		Timezone     string `yaml:"tz"`
		OSS          string `yaml:"OSS"`
		FallbackLang string `yaml:"fallbacklang"`
	} `yaml:"bot"`
	Admin struct {
		Admin      []int64 `yaml:"admin"`
		Superadmin []int64 `yaml:"superadmin"`
	} `yaml:"admin"`
}

// LotteryEventType 抽奖活动类型
type LotteryEventType struct {
	ID                 uint `gorm:"primaryKey"`
	EventID            string
	Owner              UserinfoType `gorm:"foreignKey:OwnerFKID;references:ID"`
	OwnerFKID          int64
	LotteryDescription string
	Prizes             datatypes.JSON
	InGroupRequired    int64
	InChannelRequired  int64
	Created            int64
	DeadlineEnabled    bool
	DrawDeadline       int64
	MemberCountEnabled bool
	DrawMemberCount    int64
	Usercount          int64
	Completed          int64
	isCompleted        bool
	LuckyUsers         []UserinfoType `gorm:"many2many:lottery_event_lucky_users"`
}

// PrizeType 奖品信息类型
type PrizeType struct {
	Name         string `json:"name"`
	Amount       int    `json:"amount"`
	IsNameHidden bool   `json:"isnamehidden"`
	HiddenCap    string `json:"hiddencap"`
}

// CreateEventType 创建抽奖类型
type CreateEventType struct {
	LotteryDescription string         `json:"LotteryDescription"`
	Prizes             datatypes.JSON `json:"prizes"`
	InGroupRequired    int64          `json:"InGroupRequired,omitempty"`
	InChannelRequired  int64          `json:"InChannelRequired,omitempty"`
	DeadlineEnabled    bool           `json:"DeadlineEnabled"`
	DrawDeadline       int64          `json:"DrawDeadline,omitempty"`
	MemberCountEnabled bool           `json:"MemberCountEnabled"`
	DrawMemberCount    int64          `json:"DrawMemberCount,omitempty"`
}

// UserinfoType 用户信息类型
type UserinfoType struct {
	Username    string
	Displayname string
	ID          int64 `gorm:"primaryKey;autoIncrement:false"`
}

// Participant 参与抽奖类型
type ParticipantType struct {
	ID       uint   `gorm:"primaryKey"`
	EventID  string `gorm:"uniqueIndex:idx_event_user"`
	Userid   int64  `gorm:"uniqueIndex:idx_event_user"`
	JoinedAt int64
	User     UserinfoType `gorm:"foreignKey:UserFKID;references:ID"`
	UserFKID int64
}

// i18n
type I18nFile struct {
	Languages map[string]map[string]string `json:"languages"`
}
