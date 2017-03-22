package main

import (
	"TelegramBot/nsuhelp"
	"bufio"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

const myId = 227605930
const botToken = "325933326:AAFWjDWFPKFjAMg9MDr_Av-g643F_UhJmNY"

// SearchWeather Возвращает строку с температурой, в противном случае ошибку.
func parseng() [5]string {
	var er [5]string
	res, err := http.Get("https://vk.com/nsuhelp?offset=10&own=1#posts")
	if err != nil {
		return er
	}

	if res.Status != "200 OK" {
		return er
	}

	textBody, err := ioutil.ReadAll(bufio.NewReader(res.Body))
	if err != nil {
		return er
	}

	mess, err := regexp.Compile("\"pi_text\">.+?</div>")
	wi_info, err := regexp.Compile("\"wi_date\" href.+?>")
	hrefReg, err := regexp.Compile("href=\".+?\"")

	text := html.UnescapeString(string(textBody))
	print(text)

	answer := mess.FindAllString(text, -1)
	wi_answer := wi_info.FindAllString(text, -1)

	brReg, err := regexp.Compile("<br/>")
	if err != nil {
		return er
	}

	trashReg, err := regexp.Compile("<.+?>")
	if err != nil {
		return er
	}
	var result [5]string
	for i, v := range answer {
		v = v[10 : len(v)-6]
		if byte(v[0]) == ' ' {
			v = v[1:]
		}
		for index := brReg.FindStringIndex(v); len(index) > 0; index = brReg.FindStringIndex(v) {
			v = v[:index[0]] + "\n" + v[index[1]:]
		}
		for index := trashReg.FindStringIndex(v); len(index) > 0; index = trashReg.FindStringIndex(v) {
			v = v[:index[0]] + v[index[1]:]
		}
		hrefText := hrefReg.FindString(wi_answer[i])

		result[i] = v + "\n\n" + "vk.com" + hrefText[6:len(hrefText)-1] + "\n\n"
	}

	return result
}

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

	go func() {
		for {
			a := nsuhelp.GetNewPosts()
			if len(a) != 0 {
				for i, b := range nsuhelp.UsersNsuHelp {
					if b {
						for _, v := range a {
							bot.Send(tgbotapi.NewMessage(int64(i), "Новый пост в \"Помогу в НГУ\":\n\n"+v))
						}
					}
				}
			}

			time.Sleep(30 * time.Second)
		}
	}()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "nsuhelpon":
				_, ok := nsuhelp.UsersNsuHelp[update.Message.From.ID]
				if !ok {
					nsuhelp.UsersNsuHelp[update.Message.From.ID] = true
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы были подписаны на рассылку."))
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже подписаны на рассылку."))
				}
			case "nsuhelpoff":
				_, ok := nsuhelp.UsersNsuHelp[update.Message.From.ID]
				if !ok {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не были подписаны на рассылку."))
				} else {
					delete(nsuhelp.UsersNsuHelp, update.Message.From.ID)
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы были отписаны от рассылки."))
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
			}
		}
	}
}
