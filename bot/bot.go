package main

import (
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/loader"
	"TelegramBot/schedule"
	"TelegramBot/weather"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

// Хранят основную информацию
var chats = make(map[int64]string)
var users = make(map[int]string)
var userGroup = make(map[int]string)
var scheduleMap = make(map[string][7]string)
var anekdotsBase = make(map[int]bool)

// Хранят количество пользователей
var chatsCount int
var usersCount int = 31

// Хранят дату последнего обновления
var gkDate string
var lkDate string

// Рабочие переменные
var weatherText string = "Погода временно недоступна, попробуйте чуть позднее."
var timeToStart string

// Логгеры
var logAll *log.Logger

// Личные данные
const myId = 227605930
const botToken = "371494091:AAGndTNOEJpsCO9_CxDuPpa9R025Lxms6UI"

func SendAnekdotsAll(bot *tgbotapi.BotAPI) error {
	for {
		joke, err := jokes.GetAnekdots()
		if err == nil {
			for i, v := range anekdotsBase {
				if v {
					bot.Send(tgbotapi.NewMessage(int64(i), joke))
				}
			}
		}

		time.Sleep(time.Minute * 30)
	}
}

// newChat Возвращает строку с новым каналом
func newChat(chat *tgbotapi.Chat) string {
	message := "Ник: @" + chat.UserName +
		"\nИмя: " + chat.FirstName +
		"\nФамилия: " + chat.LastName +
		"\nЗаголовок: " + chat.Title +
		"\nID: " + fmt.Sprintf("%d", chat.ID) +
		"\nТип: " + chat.Type

	chatsCount++

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
			"/sendmelog <data> - Присылает файл с логами.\n" +
			"/sendall <текст> - Делает рассылку текста. \n" +
			"/reset - Завершает текущую сессию бота."
	case "users":
		if arg == "all" {
			for _, v := range users {
				message += loader.WriteUsers(v) + "\n\n"
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
		if arg == "data" {
			_, err := bot.Send(tgbotapi.NewMessage(myId, "Отправляю..."))
			if err != nil {
				logAll.Print("Что-то пошло не так при sendmelog", err)
			}

			var name string

			switch arg {
			case "data":
				name = timeToStart
			}
			_, err = bot.Send(tgbotapi.NewDocumentUpload(myId, name))
			if err != nil {
				_, err = bot.Send(tgbotapi.NewMessage(myId, "Не удалось отправить файл."))
				if err != nil {
					logAll.Print("С отправкой файла всё плохо.")
				}

				logAll.Print("Ошибка отправки файла лога:", err)
			}
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(myId, "Попробуй ещё раз ввести аргументы правильно:\n"+
				"'data' - Файл полного лога."))
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

	timeToStart, err = loader.LoadLoggers(&logAll)
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
			answer, err := schedule.ParseSchedule(scheduleMap, "GK", &gkDate, &lkDate)
			if err != nil {
				logAll.Print(err)
			} else {
				if answer != "" {
					logAll.Print(answer)
				}
			}

			answer, err = schedule.ParseSchedule(scheduleMap, "LK", &gkDate, &lkDate)
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

	go SendAnekdotsAll(bot)

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
	usersCount, err = loader.LoadUsers(users)
	loader.LoadChats(chats)
	loader.LoadSchedule(scheduleMap)

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

		m, ok, err := loader.NewUserInfo(users, &update)
		if err != nil {
			logAll.Print(err)
		}

		if ok {
			bot.Send(tgbotapi.NewMessage(myId, "Новый пользователь!\n"+m))
			usersCount++
		} else {
			loader.ReloadUserDate(users, update.Message.From.ID)
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
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, schedule.PrintSchedule(scheduleMap, userGroup, update.Message.CommandArguments(), 0, update.Message.From.ID))
			case "tomorrow":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, schedule.PrintSchedule(scheduleMap, userGroup, update.Message.CommandArguments(), 1, update.Message.From.ID))
			case "setgroup":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, customers.AddGroupNumber(scheduleMap, userGroup, update.Message.From.ID, update.Message.CommandArguments()))
			case "joke":
				joke, err := jokes.GetAnekdots()
				if err == nil {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, joke)
				}
			case "jokeon":
				anekdotsBase[update.Message.From.ID] = true
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы согласились на рассылку анекдотов.")
			case "jokeoff":
				anekdotsBase[update.Message.From.ID] = false
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы отказались от рассылки анекдотов.")
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
				if update.Message.Command() == "reset" {
					bot.Send(tgbotapi.NewMessage(myId, "Выключаюсь."))
					logAll.Print("Выключаюсь по приказу.")
					break
				}

				sendMembers(update.Message.Command(), update.Message.CommandArguments(), bot)
			}
		}
	}
}
