package nsuhelp

import (
	"bufio"
	"errors"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
)

const CountPost = 5

var LatestPosts [CountPost][2]string
var UsersNsuHelp = make(map[int]bool)

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

	index1Reg, err := regexp.Compile("<div class=\"wall_item\"")
	if err != nil {
		return er, err
	}

	index2Reg, err := regexp.Compile("<div class=\"wi_buttons\"")
	if err != nil {
		return er, err
	}

	hrefReg, err := regexp.Compile("href=\".+?\"")
	if err != nil {
		return er, err
	}

	spaceReg, err := regexp.Compile("<.+?>")
	if err != nil {
		return er, err
	}

	brReg, err := regexp.Compile("<br/>")
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

	index1Text := index1Reg.FindAllStringIndex(text, CountPost)
	index2Text := index2Reg.FindAllStringIndex(text, CountPost)

	//log.Print(index1Text)
	//log.Print(index2Text)

	if len(index1Text) != CountPost || len(index2Text) != CountPost {
		return er, errors.New("Мало постов.")
	}

	var result [CountPost][2]string

	for i := 0; i < CountPost; i++ {
		buffer := text[index1Text[CountPost-i-1][1]:index2Text[CountPost-i-1][0]]

		result[i][0] = infoReg.FindString(buffer)
		result[i][1] = textReg.FindString(buffer)

		//log.Print(i, result[i][0]+"\n"+result[i][1])

		if len(result[i][1]) > 16 {
			result[i][1] = result[i][1][10 : len(result[i][1])-6]
		}

		if len(result[i][1]) > 0 {
			if byte(result[i][1][0]) == ' ' {
				result[i][1] = result[i][1][1:]
			}

			for index := brReg.FindStringIndex(result[i][1]); len(index) > 0; index = brReg.FindStringIndex(result[i][1]) {
				result[i][1] = result[i][1][:index[0]] + "\n" + result[i][1][index[1]:]
			}

			for index := spaceReg.FindStringIndex(result[i][1]); len(index) > 0; index = spaceReg.FindStringIndex(result[i][1]) {
				result[i][1] = result[i][1][:index[0]] + result[i][1][index[1]:]
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
	p, err := GetLatestPosts("nsuhelp")
	if err != nil {
		return nil
	}

	if (p[0][0] == LatestPosts[0][0]) && (p[CountPost-1][0] == LatestPosts[CountPost-1][0]) {
		return nil
	}

	var index int
	for ; (index < CountPost) && (p[0][0] != LatestPosts[index][0]); index++ {
	}
	//log.Print("Индекс: ", index)

	result = p[:index]

	if index < CountPost && (p[CountPost-1][0] != LatestPosts[CountPost-1][0]) {
		result = append(result, p[CountPost-1])
	}

	LatestPosts = p

	return
}

func GetGroupPost(groupName string) [CountPost][2]string {
	p, err := GetLatestPosts(groupName)
	if err != nil || p[0][0] == "" {
		p[0][0] = "Не удалось сделать запрос в эту группу."
		return p
	}

	return p
}
