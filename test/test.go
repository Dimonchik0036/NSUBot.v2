package main

import (
	"TelegramBot/loader"
	"TelegramBot/menu"
	"TelegramBot/schedule"
	"TelegramBot/weather"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"time"
)

const myId = 227605930
const botToken = "325933326:AAFWjDWFPKFjAMg9MDr_Av-g643F_UhJmNY"

var gkDate string
var lkDate string

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return
	}

	_, err = bot.Send(tgbotapi.NewMessage(myId, "Я перезагрузился."))
	if err != nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return
	}

	loader.LoadUserGroup()

	schedule.GetAllSchedule("GK", &gkDate, &lkDate)
	schedule.GetAllSchedule("LK", &gkDate, &lkDate)

	go func() {
		for {
			weather.SearchWeather()
			time.Sleep(time.Minute)
		}
	}()

	for update := range updates {
		go func() {
			msg, err := menu.MessageProcessing(update)
			if err != nil {
				log.Print(err)
				return
			}

			bot.Send(msg)
		}()
	}
}
