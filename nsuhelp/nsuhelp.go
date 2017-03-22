package nsuhelp

import (
	"bufio"
	"errors"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
)

var LatestPosts [5]string
var UsersNsuHelp = make(map[int]bool)

func GetLatestPosts() ([5]string, error) {
	var er [5]string
	res, err := http.Get("https://vk.com/nsuhelp")
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

	hrefReg, err := regexp.Compile("href=\".+?\"")
	if err != nil {
		return er, err
	}

	spaceReg, err := regexp.Compile("<.+?>")
	if err != nil {
		return er, err
	}

	text := html.UnescapeString(string(textBody))

	postText := textReg.FindAllString(text, 5)
	postHref := infoReg.FindAllString(text, 5)

	brReg, err := regexp.Compile("<br/>")
	if err != nil {
		return er, err
	}

	var result [5]string
	for i, v := range postText {
		if len(v) > 16 {
			v = v[10 : len(v)-6]
		} else {
			return er, errors.New("Галюн сообщений")
		}

		if byte(v[0]) == ' ' {
			v = v[1:]
		}

		for index := brReg.FindStringIndex(v); len(index) > 0; index = brReg.FindStringIndex(v) {
			v = v[:index[0]] + "\n" + v[index[1]:]
		}

		for index := spaceReg.FindStringIndex(v); len(index) > 0; index = spaceReg.FindStringIndex(v) {
			v = v[:index[0]] + v[index[1]:]
		}

		hrefText := hrefReg.FindString(postHref[i])

		if len(hrefText) > 7 {
			result[i] = v + "\n\n" + "vk.com" + hrefText[6:len(hrefText)-1] + "\n\n"
		} else {
			return er, errors.New("Галюн сообщений")
		}
	}

	return result, nil
}

func GetNewPosts() []string {
	p, err := GetLatestPosts()
	if err != nil {
		return nil
	}

	if p[0] == LatestPosts[0] {
		return nil
	}

	var index int
	for ; (index < 5) && (p[0] != LatestPosts[index]); index++ {
	}

	LatestPosts = p

	return p[:index]
}
