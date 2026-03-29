package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-yaml/yaml"
	tb "gopkg.in/telebot.v4"
)

var bot *tb.Bot
var conf *BotConfig

// 读取配置文件
func readConfig() {
	yamlFile, err := os.ReadFile("../data/config.yaml")

	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

// 创建机器人进程
func createBot() *tb.Bot {
	httpTimeout := conf.HTTP.Timeout
	transport := &http.Transport{}

	var err error
	loc, err = time.LoadLocation(conf.Bot.Timezone)
	if err != nil {
		loc = time.Local
	}

	// HTTP 代理
	if conf.HTTP.Proxy != "" {
		transport.Proxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse(conf.HTTP.Proxy)
		}
	}

	// 自定义 botapi
	t_api := "https://api.telegram.org"
	if conf.HTTP.Api != "" {
		t_api = conf.HTTP.Api
	}

	// 设定超时
	client := &http.Client{
		Timeout:   time.Duration(httpTimeout) * time.Second,
		Transport: transport,
	}

	// 设定 bot
	bot, err := tb.NewBot(tb.Settings{
		Token: conf.Bot.Token,
		Poller: &tb.LongPoller{
			Timeout: time.Duration(conf.Bot.Poller) * time.Second,
		},
		URL:     t_api,
		Client:  client,
		Verbose: false,
	})

	if err != nil {
		log.Fatal(err)
		return nil
	}
	return bot
}

func main() {
	readConfig()
	LoadI18n()
	DB = initDB()
	bot = createBot()

	// bot.Handle(tb.OnText, ListenCreated)
	bot.Handle("/start", Start)
	bot.Handle("/help", Help)
	bot.Handle("/about", About)
	bot.Handle("/id", GetUserID)
	//	bot.Handle("/joined", JoinedQuery)
	bot.Handle("/create", CreateLottery)    //adminonly
	bot.Handle("/gentext", GenerateTextCMD) //adminonly
	bot.Handle("/query", QueryLottery)      //adminonly
	bot.Handle("/list", ListLottery)        //adminonly
	//	bot.Handle("/draw", DrawLottery)     //adminonly
	//	bot.Handle("/cancel", CancelLottery)   //adminonly
	bot.Handle("/delete", DeleteLottery) //adminonly
	//	bot.Handle("/lucky_list", LuckyList) //adminonly
	bot.Handle(tb.OnCallback, ProcessCallback)

	log.Println("bot 启动成功!")

	bot.Start()
}
