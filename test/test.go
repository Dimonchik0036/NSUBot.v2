package main

import (
	"TelegramBot/nsuhelp"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
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

	nsuhelp.UsersNsuHelp[myId] = true

	go func() {
		for {
			a := nsuhelp.GetNewPosts()
			if len(a) != 0 {
				if a[0][0] != "" {
					for i, b := range nsuhelp.UsersNsuHelp {
						if b {
							for _, v := range a {
								if len(v[1]) > 4500 {
									v[1] = v[1][:4500] + "\n\n>>> Достигнуто ограничение на размер сообщения, перейдите по ссылке в начале сообщения, если хотите дочитать. <<<"
								}
								bot.Send(tgbotapi.NewMessage(int64(i), v[0]+"\n\n"+v[1]))
							}
						}
					}

				}
			}

			time.Sleep(57 * time.Minute)
		}
	}()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "nsuhelp":
				v, ok := nsuhelp.UsersNsuHelp[update.Message.From.ID]
				if !ok {
					nsuhelp.UsersNsuHelp[update.Message.From.ID] = true
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы были подписаны на рассылку."))
				} else {
					if v {
						delete(nsuhelp.UsersNsuHelp, update.Message.From.ID)
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы были отписаны от рассылки."))
					} else {
						nsuhelp.UsersNsuHelp[update.Message.From.ID] = true
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы были подписаны на рассылку."))
					}
				}
			case "check":
				switch update.Message.From.ID {
				case myId:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Чё надо, хозяин?"))
				case 161872635:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Кирилл, эта команда не для тебя!\n\nP.S. Жека пидор."))
				case 61219035:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Жека, не дудось!\n\nP.S. Кирилл пидор."))
				default:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Дай копейку на дошик одмину."))
				}
			case "post":
				a, err := nsuhelp.GetGroupPost(update.Message.CommandArguments())
				if err == nil {
					if a[0][0] != "" {
						for _, v := range a {
							bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, v[0]+"\n\n"+v[1]))
						}
					} else {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Группа не валидна."))
					}
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Группа не валидна."))
				}
			case "default":
				if (update.Message.From.ID == myId) && (update.Message.CommandArguments() != "") {
					_, err := nsuhelp.GetGroupPost(update.Message.CommandArguments())
					if err == nil {
						nsuhelp.DefaulGroup = update.Message.CommandArguments()
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Готово."))
					} else {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нужно больше золота."))
					}

				}
			}
		}
	}
}
