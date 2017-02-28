package main

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
	"os"
)

func parseTitle(text string) string {
	titleRegexp, err := regexp.Compile("<H1>.*</H1>")
	facRegexp, err := regexp.Compile(".*>.*</A>")
	facNameRegexp, err := regexp.Compile(">.*<")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
	}

	titleText := titleRegexp.FindString(text)

	if len(titleText) > 5 {
		titleText = titleText[4 : len(titleText)-5]
	} else {
		titleText = "Неизвестная группа"
	}

	facText := facRegexp.FindAllString(text, 2)

	var facName string

	if len(facText) > 1 {
		facName = facNameRegexp.FindString(facText[1])
		if len(facName) > 3 {
			facName = facName[1 : len(facName)-1]
		} else {
			facName = "Неизвестный факультет"
		}
	}

	return facName + "\n" + titleText
}

func parseTable(name string, group string) string {

	res, err := http.Get("http://old.nsu.ru/education/schedule/Html_"+group+"/Groups/"+name)
	if err != nil {
		log.Fatal("Не удалось загрузить страницу:", err)
	}
	println(res.Status, name)


	if res.Status != "200 OK" {
		log.Println("Ошибочка")
		return ""
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal("Ошибка чтения страницы:", err)
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Fatal("Ошибка чтения страницы:", err)
	}

	textIn := html.UnescapeString(string(textEsc))

	title := parseTitle(textIn)
	text := []byte(textIn)

	blocksRegexp, err := regexp.Compile("</TR>[^><]")
	beginRegexp, err := regexp.Compile("<TD>")
	endRegexp, err := regexp.Compile("</TD>")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
	}

	n := blocksRegexp.FindAllIndex(text, -1)
	//fmt.Print(string(text))
	//fmt.Print(n)
	var table [7][]byte
	for i := 0; i < 7; i++ {
		table[i] = text[n[i][1]:n[i+1][1]]
	}

	var tableDay [7][7][]byte
	for k := 0; k < 7; k++ {

		begin := beginRegexp.FindAllIndex(table[k], -1)
		end := endRegexp.FindAllIndex(table[k], -1)
		end = end[:]

		var count, index int
		for i := 1; i < len(begin); i++ {
			if begin[i][1] > end[i][1] {
				tableDay[k][count] = []byte(">" + string(table[k][begin[index][1]:end[i][0]]))
				if end[i][0]-begin[index][1] == 2 {
					tableDay[k][count] = []byte("-")
				}
				index = i
				count++
			}
		}

		if count != 7 {
			tableDay[k][count] = []byte("-")
		}
	}

	words, err := regexp.Compile(">[а-яА-Я][^a-zA-Z]+?<")

	for i := 0; i < 7; i++ {
		for j := 0; j < 7; j++ {
			result := words.FindAll(tableDay[i][j], -1)
			var text string
			var symbol string
			for _, v := range result {
				text += symbol + string(v[1:len(v)-1])
				symbol = ", "
			}
			if text == "" {
				text = "Пусто"
			}

			tableDay[i][j] = []byte(text)
		}
	}
	var number int

	switch time.Now().Weekday().String() {
	case "Monday":
		number = 0
	case "Tuesday":
		number = 1
	case "Wednesday":
		number = 2
	case "Thursday":
		number = 3
	case "Friday":
		number = 4
	case "Saturday":
		number = 5
	case "Sunday":
		number = 6
	}

	return title+"\n"+
		"9:00: " + string(tableDay[0][number]) + "\n" +
		"10:50: " + string(tableDay[1][number]) + "\n" +
		"12:40: " + string(tableDay[2][number]) + "\n" +
		"14:30: " + string(tableDay[3][number]) + "\n" +
		"16:20: " + string(tableDay[4][number]) + "\n" +
		"18:10: " + string(tableDay[5][number]) + "\n" +
		"20:00: " + string(tableDay[6][number]) + "\n"
}

func main() {
	res, err := http.Get("http://www.nsu.ru/education/schedule/Html_GK/Groups/")
	if err != nil {
		log.Fatal("Не удалось загрузить страницу:", err)
	}

	println(res.Status)

	if res.Status != "200 OK" {
		log.Println("Ошибочка")
		return
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal("Ошибка чтения страницы:", err)
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Fatal("Ошибка чтения страницы:", err)
	}

	text := html.UnescapeString(string(textEsc))

	data, err := regexp.Compile("[0-9a-zA-Z-]+ [0-9:]{5}")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
	}

	println(data.FindString(text))

	hrefRegexp, err := regexp.Compile(">[0-9a-z]*_[0-9]*[.]htm")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
	}

	hrefK := hrefRegexp.FindAllString(text, -1)

	schedule := make(map[string]string)
	for _, v := range hrefK {
		schedule[v[1:]] = parseTable(v[1:], "GK")
	}

	print(len(hrefK))

	file, err := os.OpenFile("result.txt", os.O_CREATE | os.O_RDWR, os.ModePerm);
	for _, v := range schedule {
		file.WriteString(v+"\n\n")
	}
}
