package main

import (
	"TelegramBot/menu"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

const myId = 227605930
const botToken = "325933326:AAFWjDWFPKFjAMg9MDr_Av-g643F_UhJmNY"

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

	for update := range updates {
		msg, err := menu.MessageProcessing(update)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Print(msg)
		bot.Send(msg)
	}
}
