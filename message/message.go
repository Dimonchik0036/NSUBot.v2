package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const myId = 227605930
const botToken = "371494091:AAGndTNOEJpsCO9_CxDuPpa9R025Lxms6UI"

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return
	}

	bot.Send(tgbotapi.NewMessage(myId, "Запущен"))
	/*_, err = bot.Send(tgbotapi.NewMessage(myId, "В течение суток были проведены технические работы, приношу свои извинения, если были перебои в работе."))
	if err != nil {
		log.Print(err)
	}

	_, err = bot.Send(tgbotapi.NewMessage(245647624, "В течение суток были проведены технические работы, приношу свои извинения, если были перебои в работе."))
	if err != nil {
		log.Print(err)
	}
	_, err = bot.Send(tgbotapi.NewMessage(142080444, "В течение суток были проведены технические работы, приношу свои извинения, если были перебои в работе."))
	if err != nil {
		log.Print(err)
	}

	_, err = bot.Send(tgbotapi.NewMessage(337845911, "В течение суток были проведены технические работы, приношу свои извинения, если были перебои в работе."))
	if err != nil {
		log.Print(err)
	}


	_, err = bot.Send(tgbotapi.NewMessage(268902362, "В течение суток были проведены технические работы, приношу свои извинения, если были перебои в работе."))
	if err != nil {
		log.Print(err)
	}

	_, err = bot.Send(tgbotapi.NewMessage(196705683, "В течение суток были проведены технические работы, приношу свои извинения, если были перебои в работе."))
	if err != nil {
		log.Print(err)
	}*/
	/*u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ведутся технические работы, прошу прощения за доставленные неудобства."))
			log.Println("[", update.CallbackQuery.From.ID, "] @"+update.CallbackQuery.From.UserName+" "+update.CallbackQuery.From.FirstName+" "+update.CallbackQuery.From.LastName+", MessageID: ", update.CallbackQuery.Message.MessageID, ", Запрос: "+update.CallbackQuery.Data)
		}

		if update.Message != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ведутся технические работы, прошу прощения за доставленные неудобства."))
			log.Println("[", update.Message.From.ID, "] @"+update.Message.From.UserName+" "+update.Message.From.FirstName+" "+update.Message.From.LastName+", Команда: "+update.Message.Text)
		}
	}*/
}
