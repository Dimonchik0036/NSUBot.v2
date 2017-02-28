package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

func main() {
	file, err := os.OpenFile(time.Now().Format("020106_1504")+".txt", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Print(err)
	}

	title, err := regexp.Compile("class=\"appdate\".*</p>")
	if err != nil {
		log.Print(err)
	}

	re, err := regexp.Compile("<td.*e..9key-.*</td>")
	if err != nil {
		log.Print(err)
	}

	eset, err := regexp.Compile("e..")
	if err != nil {
		log.Print(err)
	}

	esetCode, err := regexp.Compile("....-....-....-....-....")
	if err != nil {
		log.Print(err)
	}

	res, err := http.Get("http://trialeset.ru/")
	if err != nil {
		log.Print(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}

	err = res.Body.Close()
	if err != nil {
		log.Print(err)
	}

	text := re.FindAllString(string(body), -1)

	code := title.FindString(string(body))
	code = code[:len(code)-4] + "\n"
	code = code[16:] + "\n"

	for _, v := range text {
		code += eset.FindString(v) + ": " + esetCode.FindString(v) + "\n"
	}

	file.WriteString(code + "\n")

	res, err = http.Get("http://trialeset.ru/eset-mobile-security")
	if err != nil {
		log.Print(err)
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}

	err = res.Body.Close()
	if err != nil {
		log.Print(err)
	}

	text = esetCode.FindAllString(string(body), -1)
	code = title.FindString(string(body))
	code = code[:len(code)-4] + "\n"
	code = code[16:]

	code += "Ключи для мобильного приложения:\n\n"

	for _, v := range text {
		code += v + "\n"
	}

	file.WriteString(code)
}
