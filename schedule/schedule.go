package schedule

import (
	"TelegramBot/customers"
	"errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

var TableSchedule = make(map[string][7]string)

// searchFacultyName Вытаскивает из текста название факультета.
func searchFacultyName(text string) (string, error) {
	titleRegexp, err := regexp.Compile("<H1>.*</H1>")
	if err != nil {
		return "", errors.New("Не удалось создать правило для regexp")
	}

	facRegexp, err := regexp.Compile(".*>.*</A>")
	if err != nil {
		return "", errors.New("Не удалось создать правило для regexp")
	}

	facNameRegexp, err := regexp.Compile(">.*<")
	if err != nil {
		return "", errors.New("Не удалось создать правило для regexp")
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

	return facName + "\n" + titleText, nil
}

// getGroupSchedule Загружает расписание группы.
func getGroupSchedule(name string, group string) error {
	res, err := http.Get("http://old.nsu.ru/education/schedule/Html_" + group + "/Groups/" + name)
	if err != nil {
		return err
	}

	if res.Status != "200 OK" {
		return errors.New("Статус страницы не верен: " + res.Status)
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return err
	}

	textIn := html.UnescapeString(string(textEsc))

	title, err := searchFacultyName(textIn)
	if err != nil {
		return err
	}

	nameRegexp, err := regexp.Compile("[0-9]+[a-zA-Z0-9а-яА-Я][.][0-9]*")
	if err != nil {
		return err
	}

	groupTitle := nameRegexp.FindString(title)
	if groupTitle == "" {
		return errors.New("Ошибка титула.")
	}

	text := []byte(textIn)

	blocksRegexp, err := regexp.Compile("</TR>[^><]")
	if err != nil {
		return err
	}

	beginRegexp, err := regexp.Compile("<TD>")
	if err != nil {
		return err
	}

	endRegexp, err := regexp.Compile("</TD>")
	if err != nil {
		return err
	}

	n := blocksRegexp.FindAllIndex(text, -1)
	if len(n) < 8 {
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
	if err != nil {
		return err
	}

	doubleDay, err := regexp.Compile("<HR")
	if err != nil {
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
					text += string(v[1:len(v)-1]) + ", "
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

	TableSchedule[groupTitle] = message

	return nil
}

// GetAllSchedule Заполняет расписание.
func GetAllSchedule(group string, gkDate *string, lkDate *string) (info string, err error) {
	res, err := http.Get("http://old.nsu.ru/education/schedule/Html_" + group + "/Groups/")
	if err != nil {
		return "", errors.New("Расписание временно недоступно.")
	}

	if res.Status != "200 OK" {
		return "", errors.New("Страница работает некорректно: " + res.Status)
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		return "", errors.New("Не удалось отформатировать страницу.")
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", errors.New("Не удалось считать содержаие body.")
	}

	res.Body.Close()

	text := html.UnescapeString(string(textEsc))

	data, err := regexp.Compile("[0-9a-zA-Z-]+ [0-9:]{5}")
	if err != nil {
		return "", err
	}

	date := data.FindString(text)

	hrefRegexp, err := regexp.Compile(">[0-9a-z]*_[0-9]*[.]htm")
	if err != nil {
		return "", err
	}

	hrefK := hrefRegexp.FindAllString(text, -1)

	var mess [7]string
	for i := 0; i < 7; i++ {
		mess[i] = "Не удалось загрузить расписание, сообщите об этом."
	}

	for _, v := range hrefK {
		err = getGroupSchedule(v[1:], group)
		if err != nil {
			info += group + " "
			TableSchedule[v[1:]] = mess
		}
	}

	if group == "GK" {
		info = "GK " + date + " " + info
		*gkDate = date
	} else {
		info = "LK " + date + " " + info
		*lkDate = date
	}

	return info, nil
}

// ParseSchedule Проверяет расписание на изменение.
func ParseSchedule(group string, gkDate *string, lkDate *string) (info string, err error) {
	res, err := http.Get("http://old.nsu.ru/education/schedule/Html_" + group + "/Groups/")
	if err != nil {
		return "", err
	}

	if res.Status != "200 OK" {
		return "", errors.New("Ошибка статуса страницы: " + res.Status)
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", err
	}

	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	text := html.UnescapeString(string(textEsc))

	data, err := regexp.Compile("[0-9a-zA-Z-]+ [0-9:]{5}")
	if err != nil {
		return "", err
	}

	date := data.FindString(text)

	if date != "" {
		if (group == "GK") && (*gkDate != date) {
			mess, err := GetAllSchedule("GK", gkDate, lkDate)
			if err == nil {
				info = "GK " + date + " " + mess
				*gkDate = date
			}
		} else {
			if (group == "LK") && (*lkDate != date) {
				mess, err := GetAllSchedule("LK", gkDate, lkDate)
				if err == nil {
					info = "LK " + date + " " + mess
					*lkDate = date
				}
			}
		}
	}

	return info, nil
}

// PrintSchedule Возвращает расписание.
func PrintSchedule(group string, offset int, id int) string {
	if len(group) > 16 {
		return "Введите корректный номер группы."
	}

	if group == "" {
		group = customers.AllLabels[id].MyGroup
	} else {
		defaultGroup, ok := customers.AllLabels[id].Group[group]
		if ok {
			group = defaultGroup
		}
	}

	v, ok := TableSchedule[group]
	if !ok {
		group += ".1"
		v, ok = TableSchedule[group]
		if !ok {
			return "Неверно задан номер группы. Воспользуйтесь /help или /faq для помощи."
		}
	}

	var textDay [7]string

	textDay[0] = "Понедельник."
	textDay[1] = "Вторник."
	textDay[2] = "Среда."
	textDay[3] = "Четверг."
	textDay[4] = "Пятница."
	textDay[5] = "Суббота."
	textDay[6] = "Воскресенье."

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

	return textDay[(number+offset)%7] + "\n" + v[(number+offset)%7]
}
