package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"os/exec"
	"time"
)

const myId = 227605930

func main() {
	file, err := os.OpenFile("logStart.txt", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	myLogger := log.New(file, "", log.LstdFlags)

	bot, err := tgbotapi.NewBotAPI("371494091:AAGndTNOEJpsCO9_CxDuPpa9R025Lxms6UI")
	if err != nil {
		myLogger.Panic("Бот в отпуске:", err)
	}

	bot.Debug = false

	for {
		myLogger.Print("Начинаю запуск...")
		_, err = bot.Send(tgbotapi.NewMessage(myId, "Запускаю бота..."))
		if err != nil {
			myLogger.Print("Не отправить сообщение боту.")
		}

		cmd := exec.Command("./TelegramBot")

		err := cmd.Start()
		if err != nil {
			myLogger.Println("Запусе не удался.")

			_, err = bot.Send(tgbotapi.NewMessage(myId, "Запуск не удался."))
			if err != nil {
				myLogger.Print("Не отправить сообщение боту.")
			}
		}

		myLogger.Print("Вторая фаза запуска...")
		err = cmd.Wait()
		if err != nil {
			myLogger.Println("Бот пал в бою.")

			_, err = bot.Send(tgbotapi.NewMessage(myId, "Бот пал в неравном бою."))
			if err != nil {
				myLogger.Print("Не отправить сообщение боту.")
			}
		}

		time.Sleep(time.Second * 30)
	}
}
