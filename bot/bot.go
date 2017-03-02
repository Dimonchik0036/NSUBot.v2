package main

import (
	"TelegramBot/loader"
	"TelegramBot/weather"
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var chats = make(map[int64]string)
var users = make(map[int]string)
var userGroup = make(map[int]string)
var schedule = make(map[string][7]string)

var chatsCount int
var usersCount int

var gkDate string
var lkDate string

var weatherText string = "Погода временно недоступна, попробуйте чуть позднее."
var logFileName string
var timeToStart string
var logUsers *log.Logger
var logAll *log.Logger

const myId = 227605930
const botToken = "371494091:AAGndTNOEJpsCO9_CxDuPpa9R025Lxms6UI"

func defaultUserSchedule(id int, group string) string {
	if group == "" {
		return "Вы не ввели номер группы."
	}

	if len(group) > 16 {
		return "Слишком много символов."
	}

	_, ok := schedule[group]
	if !ok {
		group += ".1"
		_, ok = schedule[group]
		if !ok {
			return "Введён некорректный номер группы, попробуйте повторить попытку или воспользоваться /help и /faq для помощи."
		}
	}

	userGroup[id] = group

	return "Группа '" + group + "' успешно назначена, нажмите на /today или /tomorrow для проверки правильности выбора."
}

func parseSchedule() {
	for {
		res, err := http.Get("http://www.nsu.ru/education/schedule/Html_GK/Groups/")
		if err != nil {
			logAll.Print("Не удалось загрузить страницу:", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		if res.Status != "200 OK" {
			logAll.Print("Чёт пошло не так")
			time.Sleep(time.Minute * 5)
			continue
		}

		bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
		if err != nil {
			logAll.Print("Ошибка чтения страницы:", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		textEsc, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			logAll.Print("Ошибка чтения страницы:", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		res.Body.Close()

		text := html.UnescapeString(string(textEsc))

		data, err := regexp.Compile("[0-9a-zA-Z-]+ [0-9:]{5}")
		if err != nil {
			logAll.Print("Не удалось создать правило для regexp")
			time.Sleep(time.Minute * 5)
			continue
		}

		date := data.FindString(text)

		if (date != "") && (gkDate != date) {
			err = scheduleNSU("GK")
			if err == nil {
				gkDate = date
			}
		}

		res, err = http.Get("http://www.nsu.ru/education/schedule/Html_LK/Groups/")
		if err != nil {
			logAll.Print("Не удалось загрузить страницу:", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		if res.Status != "200 OK" {
			logAll.Print("Чёт пошло не так")
			time.Sleep(time.Minute * 5)
			continue
		}

		bodyReader, err = charset.NewReader(res.Body, res.Header.Get("Content-Type"))
		if err != nil {
			logAll.Print("Ошибка чтения страницы:", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		textEsc, err = ioutil.ReadAll(bodyReader)
		if err != nil {
			logAll.Print("Ошибка чтения страницы:", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		res.Body.Close()

		text = html.UnescapeString(string(textEsc))

		data, err = regexp.Compile("[0-9a-zA-Z-]+ [0-9:]{5}")
		if err != nil {
			logAll.Print("Не удалось создать правило для regexp")
			time.Sleep(time.Minute * 5)
			continue
		}

		date = data.FindString(text)

		if (date != "") && (lkDate != date) {
			err = scheduleNSU("LK")
			if err == nil {
				lkDate = date
			}
		}

		time.Sleep(time.Minute * 5)
	}
}

func printSchedule(name string, offset int, id int) string {
	if len(name) > 16 {
		return "Введите корректный номер группы."
	}

	if name == "" {
		n, ok := userGroup[id]
		if ok {
			name = n
		}
	}

	v, ok := schedule[name]
	if !ok {
		name += ".1"
		v, ok = schedule[name]
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

func scheduleNSU(group string) error {
	res, err := http.Get("http://www.nsu.ru/education/schedule/Html_" + group + "/Groups/")
	if err != nil {
		logAll.Print("Не удалось загрузить страницу:", err)
		return errors.New("Расписание временно недоступно.")
	}

	if res.Status != "200 OK" {
		logAll.Print("Чёт пошло не так")
		return errors.New("Не удалось найти такую группу.")
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		logAll.Print("Ошибка чтения страницы:", err)
		return errors.New("Упс, что-то пошло не так.")
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		logAll.Print("Ошибка чтения страницы:", err)
		return errors.New("Упс, что-то пошло не так.")
	}

	res.Body.Close()

	text := html.UnescapeString(string(textEsc))

	data, err := regexp.Compile("[0-9a-zA-Z-]+ [0-9:]{5}")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp")
		return errors.New("Не удалось создать regexp data")
	}

	date := data.FindString(text)

	hrefRegexp, err := regexp.Compile(">[0-9a-z]*_[0-9]*[.]htm")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp")
		return errors.New("Не удалось создать regexp data")
	}

	hrefK := hrefRegexp.FindAllString(text, -1)

	var mess [7]string
	for i := 0; i < 7; i++ {
		mess[i] = "Не удалось загрузить расписание, сообщите об этом."
	}

	for _, v := range hrefK {
		err = parseTable(v[1:], group)
		if err != nil {
			logAll.Print("Не удалось выставить расписание: " + group)
			schedule[v[1:]] = mess
		}
	}

	if group == "GK" {
		logAll.Print(date)
		gkDate = date
	} else {
		logAll.Print(date)
		lkDate = date
	}

	return nil
}

// parseTable
func parseTable(name string, group string) error {
	res, err := http.Get("http://old.nsu.ru/education/schedule/Html_" + group + "/Groups/" + name)
	if err != nil {
		logAll.Print("Не удалось загрузить страницу:", err)
		return err
	}

	if res.Status != "200 OK" {
		logAll.Println("Статус страницы не верен")
		return errors.New("Статус страницы не верен.")
	}

	bodyReader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		logAll.Print("Ошибка чтения страницы:", err)
		return err
	}

	textEsc, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		logAll.Print("Ошибка чтения страницы:", err)
		return err
	}

	textIn := html.UnescapeString(string(textEsc))

	title := parseTitle(textIn)

	nameRegexp, err := regexp.Compile("[0-9]+[a-zA-Z0-9а-яА-Я][.][0-9]*")
	if err != nil {
		logAll.Print("Не смог сделать regexp:", err)
		return err
	}

	groupTitle := nameRegexp.FindString(title)
	if groupTitle == "" {
		logAll.Print("Ошибка титула")
		return errors.New("Ошибка титула")
	}

	text := []byte(textIn)

	blocksRegexp, err := regexp.Compile("</TR>[^><]")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp")
		return err
	}

	beginRegexp, err := regexp.Compile("<TD>")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp")
		return err
	}

	endRegexp, err := regexp.Compile("</TD>")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp")
		return err
	}

	n := blocksRegexp.FindAllIndex(text, -1)
	if len(n) < 8 {
		logAll.Print("Неверное количество блоков")
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
		logAll.Print("Не удалось создать правило для regexp")
		return err
	}

	doubleDay, err := regexp.Compile("<HR")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp")
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

	schedule[groupTitle] = message

	return nil
}

// parseTitle
func parseTitle(text string) string {
	titleRegexp, err := regexp.Compile("<H1>.*</H1>")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp", err)
		return ""
	}

	facRegexp, err := regexp.Compile(".*>.*</A>")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp", err)
		return ""
	}

	facNameRegexp, err := regexp.Compile(">.*<")
	if err != nil {
		logAll.Print("Не удалось создать правило для regexp", err)
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

// newChat Возвращает строку с новым каналом
func newChat(chat *tgbotapi.Chat) string {
	message := "Ник: @" + chat.UserName +
		"\nИмя: " + chat.FirstName +
		"\nФамилия: " + chat.LastName +
		"\nЗаголовок: " + chat.Title +
		"\nID: " + fmt.Sprintf("%d", chat.ID) +
		"\nТип: " + chat.Type

	logUsers.Println("\n'Чат'\n" + message + "\n")
	chatsCount++

	return message
}

// newUser Возвращает строку с новым пользователем
func newUser(update *tgbotapi.Update) string {
	message := "Ник: @" + update.Message.From.UserName +
		"\nИмя: " + update.Message.From.FirstName +
		"\nФамилия: " + update.Message.From.LastName +
		"\nID: " + fmt.Sprintf("%d", update.Message.From.ID)

	logUsers.Println("\n'Пользователь'\n" + message + "\n")

	usersCount++

	return message
}

// sendMembers Отправляет статистику по пользователям
func sendMembers(commands string, arg string, bot *tgbotapi.BotAPI) {
	var message string

	switch commands {
	case "help":
		message += "/users <_|all> - Выводит статистику по пользователям.\n" +
			"/groups <_|all> - Выводит статистику по каналам.\n" +
			"/setmessage <текст> - Задаёт сообщение, которое будет отображаться вместо погоды.\n" +
			"/sendmelog <data|log> - Присылает файл с логами.\n" +
			"/sendall <текст> - Делает рассылку текста. "
	case "users":
		if arg == "all" {
			for _, v := range users {
				message += v
				message += "\n\n"
			}
		}

		message += "Количество пользователей: " + strconv.Itoa(usersCount)
	case "groups":
		if arg == "all" {
			for _, v := range chats {
				message += v + "\n\n"
			}

		}

		message += "Количество чатов: " + strconv.Itoa(chatsCount)
	case "sendall":
		if arg != "" {
			logAll.Print("Рассылаю всем: '" + arg + "'")

			for i := range users {
				_, err := bot.Send(tgbotapi.NewMessage(int64(i), arg))
				if err != nil {
					logAll.Print("Что-то пошло не так при рассылке ["+string(i)+"]", err)
				}
			}
		}

		return
	case "setmessage":
		weatherText = arg
		logAll.Print("Обновлена строка температуры на: " + weatherText)
		message += "Готово!\n" + "'" + weatherText + "'"
	case "sendmelog":
		if arg == "data" || arg == "log" {
			_, err := bot.Send(tgbotapi.NewMessage(myId, "Отправляю..."))
			if err != nil {
				logAll.Print("Что-то пошло не так при sendmelog", err)
			}

			var name string

			switch arg {
			case "data":
				name = timeToStart
			case "log":
				name = logFileName
			}
			_, err = bot.Send(tgbotapi.NewDocumentUpload(myId, name))
			if err != nil {
				_, err = bot.Send(tgbotapi.NewMessage(myId, "Не удаловь отправить файл."))
				if err != nil {
					logAll.Print("С отправкой файла всё плохо.")
				}

				logAll.Print("Ошибка отправки файла лога:", err)
			}
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(myId, "Попробуй ещё раз ввести аргументы правильно:\n"+
				"'data' - Файл полного лога.\n"+
				"'log' - Файл с пользователями."))
			if err != nil {
				logAll.Print("Что-то пошло не так", err)
			}
		}

		return

	default:
		return
	}

	_, err := bot.Send(tgbotapi.NewMessage(myId, message))
	if err != nil {
		logAll.Print("Ошибка отправки сообщения - комманды:", err)
	}
}

func main() {
	var err error

	logFileName, timeToStart, err = loader.InitLoggers(logUsers, logAll)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logAll.Panic("Бот в отпуске:", err)
	}

	bot.Debug = false

	err = scheduleNSU("GK")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(myId, "Всё плохо с GK"))
		logAll.Panic("GK")
	}

	err = scheduleNSU("LK")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(myId, "Всё плохо с LK"))
		logAll.Panic("LK")
	}

	go func() {
		for {
			answer, err := weather.SearchWeather()
			if err != nil {
				logAll.Print(err)
			} else {
				weatherText = answer
			}

			time.Sleep(time.Minute)
		}
	}()

	go parseSchedule()

	logAll.Printf("Бот %s запущен.", bot.Self.UserName)

	_, err = bot.Send(tgbotapi.NewMessage(myId, "Я перезагрузился."))
	if err != nil {
		logAll.Print("Не смог отправить весточку повелителю.", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logAll.Panic(err)
	}

	loader.LoadUserGroup(userGroup)
	loader.LoadUsers(users)
	loader.LoadChats(chats)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.Type != "private" {
			_, ok := chats[update.Message.Chat.ID]
			if !ok {
				n := newChat(update.Message.Chat)
				chats[update.Message.Chat.ID] = n

				_, err := bot.Send(tgbotapi.NewMessage(myId, "Новая чат-сессия!\n"+n))
				if err != nil {
					logAll.Print("newChat:", err)
				}
			}
		}

		_, ok := users[update.Message.From.ID]
		if !ok {
			n := newUser(&update)
			users[update.Message.From.ID] = n

			_, err := bot.Send(tgbotapi.NewMessage(myId, "Новый пользователь!\n"+n))
			if err != nil {
				logAll.Print("newUser:", err)
			}
		}

		if update.Message.Chat.IsGroup() || update.Message.Chat.IsChannel() || update.Message.Chat.IsSuperGroup() {
			logAll.Printf("[%d] %s",
				update.Message.Chat.ID, "'"+
					update.Message.Chat.Title+"' "+
					update.Message.From.FirstName+" "+
					update.Message.From.LastName+" (@"+
					update.Message.From.UserName+")")

		}

		logAll.Printf("[%d] %s: %s",
			update.Message.From.ID,
			update.Message.From.FirstName+" "+
				update.Message.From.LastName+" (@"+
				update.Message.From.UserName+")",
			update.Message.Text)

		var msg tgbotapi.MessageConfig
		var nilMsg bool

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID,
					"Привет!\nЯ - бот, который способен показать температуру около НГУ и расписание занятий.\n"+
						"Рекомендую воспользоваться командой /help, чтобы узнать все возможности. Если возникнут вопросы, то можно воспользоваться /faq.")
			case "help":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID,
					"Список команд:\n"+
						"/help - Показать список команд\n\n"+
						"/weather - Показать температуру воздуха около НГУ\n\n"+
						"/today <номер группы> - Показывает расписание занятий конкретной группы, пример: /today 16211.1\n\n"+
						"/tomorrow <номер группы> - Показывает расписание занятий конкретной группы на завтра, пример: /tomorrow 16211.1\n\n"+
						"/setgroup <номер группы> - Устанавливает номер группы для быстрого доступа. Например, если ввести /setgroup 16211.1,"+
						" то при использовании /today или /tomorrow без аргументов, будет показываться расписание группы 16211.1\n\n"+
						"/faq - Типичные вопросы и ответы на них.\n\n"+
						"P.S. Значёк <|> показывает, что это двойная пара. Отображение только текущей недели будет добавлено чуть позже.")
			case "faq":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID,
					"Q: Можно ли пользоваться командой /today и /tomorrow и не вводить номер группы каждый раз?\n"+
						"A: Да, можно. Для этого необходимо воспользоваться командой /setgroup.\n\n"+

						"Q: Как установить номер группы для быстрого доступа?\n"+
						"A: Необходимо ввести /setgroup <номер группы>, где <номер группы> - это желаемый номер группы(треугольные скобки писать не нужно).\n"+
						"Пример: /setgroup 16211.1\n\n"+

						"Q: Можно ли посмотреть расписание, если не работает официальный сайт с расписанием?\n"+
						"A: Да, можно.\n\n"+

						"Q: Как часто обновляется расписание?\n"+
						"A: Сразу же после изменений в официальном расписании.\n\n"+

						"Если у Вас остались ещё какие-то вопросы, то их можно задать мне @dimonchik0036.")
			case "creator", "maker", "author", "father", "Creator", "Maker", "Author", "Father":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "@Dimonchik0036\ngithub.com/dimonchik0036")
			case "Погода", "weather", "погода", "Weather", "weather_nsu":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, weatherText)
			case "today":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, printSchedule(update.Message.CommandArguments(), 0, update.Message.From.ID))
			case "tomorrow":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, printSchedule(update.Message.CommandArguments(), 1, update.Message.From.ID))
			case "setgroup":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, defaultUserSchedule(update.Message.From.ID, update.Message.CommandArguments()))
			default:
				nilMsg = true

			}

			if !nilMsg {
				_, err := bot.Send(msg)
				if err != nil {
					logAll.Print("Command:", err)
				}
			}

			if update.Message.From.ID == myId {
				sendMembers(update.Message.Command(), update.Message.CommandArguments(), bot)
			}
		}
	}
}
