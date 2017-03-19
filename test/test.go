package main

import (
	"bufio"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
	"TelegramBot/nsuhelp"
)

const myId = 227605930
const botToken = "325933326:AAFWjDWFPKFjAMg9MDr_Av-g643F_UhJmNY"

// SearchWeather Возвращает строку с температурой, в противном случае ошибку.
func parseng() [5] string {
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
	for i, v := range answer{
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

		result[i] = v + "\n\n"+"vk.com"+hrefText[6:len(hrefText)-1]+"\n\n"
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

	for update := range updates {
		if update.Message == nil {
			continue
		}
		a, _ := nsuhelp.GetLatestPosts()
		for _, v := range a {
			bot.Send(tgbotapi.NewMessage(myId, "Новый пост в \"Помогу в НГУ\":\n"+v))
		}
	}
}
