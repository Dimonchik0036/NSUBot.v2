package subscriptions

import (
	"TelegramBot/all_types"
	"TelegramBot/mymodule"
	"TelegramBot/vkapi"
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
)

const CountPost = 5

var LatestPosts [CountPost][2]string

var ParserGroup = "nsuhelp"

const NsuHelp = "nsuhelp"
const NsuFit = "nsufit"

func GetLatestPosts(groupName string) ([CountPost][2]string, error) {
	var er [CountPost][2]string
	if groupName == "" {
		return er, errors.New("Не задана группа.")
	}

	res, err := http.Get("https://vk.com/" + groupName)
	if err != nil {
		return er, err
	}

	if res.Status != "200 OK" {
		return er, err
	}

	textBody, err := ioutil.ReadAll(bufio.NewReader(res.Body))
	if err != nil {
		return er, err
	}

	textReg, err := regexp.Compile("\"pi_text\">.+?</div>")
	if err != nil {
		return er, err
	}

	infoReg, err := regexp.Compile("\"wi_date\" href.+?>")
	if err != nil {
		return er, err
	}

	anchoredReg, err := regexp.Compile("<div class=\"wi_explain\"")
	if err != nil {
		return er, err
	}

	hrefReg, err := regexp.Compile("href=\".+?\"")
	if err != nil {
		return er, err
	}

	titleReg, err := regexp.Compile("<title>.*</title>")
	if err != nil {
		return er, err
	}

	text := html.UnescapeString(string(textBody))

	titleText := titleReg.FindString(text)
	if titleText == "" {
		return er, errors.New("Отсутствует заголовок.")
	} else {
		titleText = titleText[7 : len(titleText)-8]
	}

	index1Text, index2Text, err := mymodule.SearchBeginEnd(text, "<div class=\"wall_item\"", "<div class=\"wi_buttons\"", CountPost)
	if err != nil {
		return er, err
	}

	var result [CountPost][2]string
	var k int

	if anchoredReg.FindString(text) == "" {
		k = 1
	}

	for f := range index1Text {
		if f+k == CountPost {
			break
		}

		buffer := text[index1Text[f][1]:index2Text[f][0]]
		i := f + k

		result[i][0] = infoReg.FindString(buffer)
		result[i][1] = textReg.FindString(buffer)

		if len(result[i][1]) > 16 {
			result[i][1] = result[i][1][10 : len(result[i][1])-6]
		}

		if len(result[i][1]) > 0 {
			if byte(result[i][1][0]) == ' ' {
				result[i][1] = result[i][1][1:]
			}

			result[i][1], err = mymodule.ChangeSymbol(result[i][1], "\n", "<br/>")
			if err != nil {
				return er, err
			}

			result[i][1], err = mymodule.ChangeSymbol(result[i][1], "", "<.+?>")
			if err != nil {
				return er, err
			}
		}

		hrefText := hrefReg.FindString(result[i][0])

		if len(hrefText) > 7 {
			result[i][0] = titleText + "\nvk.com" + hrefText[6:len(hrefText)-1]
		} else {
			return er, errors.New("Галюн сообщений")
		}
	}

	return result, nil
}

func GetNewPosts() (result [][2]string) {
	p, err := GetLatestPosts(ParserGroup)
	if err != nil {
		return nil
	}

	if len(p) == 0 {
		return nil
	}

	if (p[1][0] == "" && p[0][0] == "") ||
		(p[1][0] == LatestPosts[1][0]) && (p[0][0] == LatestPosts[0][0]) {

		return nil
	}

	for i := len(p) - 1; i > 0; i-- {
		flag := true

		for j := 1; j < len(LatestPosts); j++ {
			if p[i] == LatestPosts[j] {
				flag = false
				break
			}
		}

		if flag {
			result = append(result, p[i])
		}
	}

	if p[0][0] != LatestPosts[0][0] {
		last := p[0]

		if last[0] != "" {
			last[0] += "\nЗакреплённая запись."
			result = append(result, last)
		}
	}

	LatestPosts = p

	return
}

func GetGroupPost(groupName string) ([CountPost][2]string, error) {
	p, err := GetLatestPosts(groupName)
	if err != nil || p[1][0] == "" {
		return p, errors.New("Группа не валидна.")
	}

	return p, err
}

func ChangeDefaultGroup(group string) string {
	_, err := GetGroupPost(group)
	if err == nil {
		ParserGroup = group
		return "Готово."
	} else {
		return "Группа не валидна."
	}
}

func GetPosts(domain string, count int) (posts []all_types.Post, err error) {
	res, err := vkapi.GetWallJson(domain, 0, count, "all")
	if err != nil {
		return
	}

	if res.Items == nil {
		return posts, errors.New("*Item равен nil")
	}

	for _, item := range *res.Items {
		if item.MarkedAsAds != 0 {
			continue
		}

		var post all_types.Post
		post.Text = item.Text
		post.Date = item.Date
		post.IsPinned = post.IsPinned
		post.Href = "https://vk.com/wall" + fmt.Sprint(item.OwnerID) + "_" + fmt.Sprint(item.ID)

		posts = append(posts, post)
	}

	return
}

func AddNewGroupToParse(domain string) (err error) {
	g, err := vkapi.GetGroup(0, domain)
	if err != nil {
		return err
	}

	var sub all_types.Subscription

	sub.Name = g.Name
	sub.ScreenName = domain
	sub.UserSubscriptions = make(map[int]int)

	posts, err := GetPosts(domain, 5)
	if err != nil {
		return
	}

	sub.Posts = &posts

	all_types.AllSubscription[domain] = &sub

	return nil
}
