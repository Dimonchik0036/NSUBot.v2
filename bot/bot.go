package main

import (
	"TelegramBot/loader"
	"TelegramBot/schedule"
	"TelegramBot/weather"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

var chats = make(map[int64]string)
var users = make(map[int]string)
var userGroup = make(map[int]string)
var scheduleMap = make(map[string][7]string)

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

	_, ok := scheduleMap[group]
	if !ok {
		group += ".1"
		_, ok = scheduleMap[group]
		if !ok {
			return "Введён некорректный номер группы, попробуйте повторить попытку или воспользоваться /help и /faq для помощи."
		}
	}

	userGroup[id] = group

	return "Группа '" + group + "' успешно назначена, нажмите на /today или /tomorrow для проверки правильности выбора."
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

	v, ok := scheduleMap[name]
	if !ok {
		name += ".1"
		v, ok = scheduleMap[name]
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

	info, err := schedule.GetAllSchedule(scheduleMap, "GK", &gkDate, &lkDate)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(myId, "Всё плохо с GK"))
		logAll.Panic("GK")
	} else {
		logAll.Print(info)
	}

	info, err = schedule.GetAllSchedule(scheduleMap, "LK", &gkDate, &lkDate)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(myId, "Всё плохо с LK"))
		logAll.Panic("LK")
	} else {
		logAll.Print(info)
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

	go func() {
		for {
			answer, err := schedule.ParseSchedule(scheduleMap, &gkDate, &lkDate)
			if err != nil {
				logAll.Print(err)
			} else {
				if answer != "" {
					logAll.Print(answer)
				}
			}

			time.Sleep(5 * time.Minute)
		}
	}()

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
