package main

import (
	"errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var schedule = make(map[string][7]string)

func parseTitle(text string) string {
	titleRegexp, err := regexp.Compile("<H1>.*</H1>")
	if err != nil {
		log.Print("Не удалось создать правило для regexp", err)
		return ""
	}

	facRegexp, err := regexp.Compile(".*>.*</A>")
	if err != nil {
		log.Print("Не удалось создать правило для regexp", err)
		return ""
	}

	facNameRegexp, err := regexp.Compile(">.*<")
	if err != nil {
		log.Print("Не удалось создать правило для regexp", err)
		return ""
	}

	titleText := titleRegexp.FindString(text)

	if len(titleText) > 5 {
		titleText = titleText[4 : len(titleText)-5]
	} else {
		titleText = ""
	}

	facText := facRegexp.FindAllString(text, 2)

	var facName string

	if len(facText) > 1 {
		facName = facNameRegexp.FindString(facText[1])
		if len(facName) > 3 {
			facName = facName[1 : len(facName)-1]
		} else {
			facName = ""
		}
	}

	return facName + "\n" + titleText
}

func parseTable(name string, group string) error {
	res, err := http.Get("http://old.nsu.ru/education/schedule/Html_" + group + "/Groups/" + name)
	if err != nil {
		log.Print("Не удалось загрузить страницу:", err)
		return err
	}

	if res.Status != "200 OK" {
		log.Println("Статус страницы не верен")
		return errors.New("Статус страницы не верен.")
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		log.Print("Ошибка чтения страницы:", err)
		return err
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Print("Ошибка чтения страницы:", err)
		return err
	}

	textIn := html.UnescapeString(string(textEsc))

	title := parseTitle(textIn)

	nameRegexp, err := regexp.Compile("[0-9]+[a-zA-Z0-9а-яА-Я][.][0-9]*")
	if err != nil {
		log.Print("Не смог сделать regexp:", err)
		return err
	}

	groupTitle := nameRegexp.FindString(title)
	if groupTitle == "" {
		log.Print("Ошибка титула")
		return errors.New("Ошибка титула")
	}

	text := []byte(textIn)

	blocksRegexp, err := regexp.Compile("</TR>[^><]")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
		return err
	}

	beginRegexp, err := regexp.Compile("<TD>")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
		return err
	}

	endRegexp, err := regexp.Compile("</TD>")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
		return err
	}

	n := blocksRegexp.FindAllIndex(text, -1)
	if len(n) < 8 {
		log.Print("Неверное количество блоков")
		return errors.New("Неверное количество блоков")
	}

	var table [7][]byte
	for i := 0; i < 7; i++ {
		table[i] = text[n[i][1]:n[i+1][1]]
	}

	var tableDay [7][7][]byte
	for k := 0; k < 7; k++ {

		begin := beginRegexp.FindAllIndex(table[k], -1)
		end := endRegexp.FindAllIndex(table[k], -1)
		end = end[:] //Что за?!?!

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
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
		return err
	}

	doubleDay, err := regexp.Compile("<HR")
	if err != nil {
		log.Print("Не удалось создать правило для regexp")
		return err
	}

	for i := 0; i < 7; i++ {
		for j := 0; j < 7; j++ {
			result := words.FindAll(tableDay[i][j], -1)
			resultIndex := words.FindAllIndex(tableDay[i][j], -1)

			doubleDayIndex := doubleDay.FindIndex(tableDay[i][j])
			doubleDayFlag := false

			if len(doubleDayIndex) > 0 {
				doubleDayFlag = true
			}

			var text string
			var symbol string

			for i, v := range result {
				if doubleDayFlag && (resultIndex[i][0] > doubleDayIndex[0]) {
					if i == 0 {
						text += "-"
					}

					text += " <|> "
					text +=  string(v[1:len(v)-1])
					doubleDayFlag = false
				} else {
					text += symbol + string(v[1:len(v)-1])
					symbol = ", "
				}
			}

			if len(doubleDayIndex) > 0 {
				if doubleDayFlag {
					text += " <|> -"
				}
			} else {
				if text == "" {
					text = "-"
				}
			}

			tableDay[i][j] = []byte(text)
		}
	}
	var message [7]string
	for number := 0; number < 7; number++ {
		message[number] = title + "\n" +
			"1 П.  9:00: " + string(tableDay[0][number]) + "\n" +
			"2 П. 10:50: " + string(tableDay[1][number]) + "\n" +
			"3 П. 12:40: " + string(tableDay[2][number]) + "\n" +
			"4 П. 14:30: " + string(tableDay[3][number]) + "\n" +
			"5 П. 16:20: " + string(tableDay[4][number]) + "\n" +
			"6 П. 18:10: " + string(tableDay[5][number]) + "\n" +
			"7 П. 20:00: " + string(tableDay[6][number]) + "\n"
	}

	schedule[groupTitle] = message

	return nil
}

func main() {
	parseTable("16342_1.htm", "GK")

	print(schedule["16342.1"][0])
}
