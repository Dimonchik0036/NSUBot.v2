package jokes

import (
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

func GetAnekdots() (string, error) {
	res, err := http.Get("https://www.anekdot.ru/random/anekdot/")
	if err != nil {
		return "", err
	}

	if res.Status != "200 OK" {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	jokeReg, err := regexp.Compile("class=\"text\">.*?</div>")
	if err != nil {
		return "", err
	}

	textBody := html.UnescapeString(string(body))

	text := jokeReg.FindString(textBody)

	if len(text) > 13 {
		text = text[13 : len(text)-6]
	}

	brReg, err := regexp.Compile("<br>")
	if err != nil {
		return "", err
	}

	for index := brReg.FindStringIndex(text); len(index) > 0; index = brReg.FindStringIndex(text) {
		text = text[:index[0]] + "\n" + text[index[1]:]
	}

	jokeFile, err := os.OpenFile("jokes.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err == nil {
		jokeFile.WriteString(">>>>>>\n" + text + "\n\n")
		jokeFile.Close()
	}

	return text, nil
}
