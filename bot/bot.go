package main

import (
	"TelegramBot/customers"
	"TelegramBot/loader"
	"TelegramBot/menu"
	"TelegramBot/schedule"
	"TelegramBot/subscriptions"
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

// Хранят количество пользователей
var chatsCount int
var usersCount int

// Хранят дату последнего обновления
var gkDate string
var lkDate string

// Рабочие переменные
var timeToStart string

// Логгеры
var logAll *log.Logger

// Личные данные
const botToken = "371494091:AAGndTNOEJpsCO9_CxDuPpa9R025Lxms6UI"

func messageLog(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if (update.Message.Chat.IsGroup() || update.Message.Chat.IsChannel() || update.Message.Chat.IsSuperGroup()) && update.Message.IsCommand() {
		logAll.Printf("[%d] %s",
			update.Message.Chat.ID, "'"+
				update.Message.Chat.Title+"' "+
				update.Message.From.FirstName+" "+
				update.Message.From.LastName+" (@"+
				update.Message.From.UserName+")")

	}
}

func processingUser(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	if update.Message.Chat.Type != "private" {
		_, ok := chats[update.Message.Chat.ID]
		if !ok {
			n := newChat(update.Message.Chat)
			chats[update.Message.Chat.ID] = n

			_, err := bot.Send(tgbotapi.NewMessage(loader.MyId, "Новая чат-сессия!\n"+n))
			if err != nil {
				logAll.Print("newChat:", err)
			}
		}
	}

	m, ok, err := loader.NewUserInfo(users, update)
	if err != nil {
		return err
	}

	if ok {
		bot.Send(tgbotapi.NewMessage(loader.MyId, "Новый пользователь!\n"+m))
		usersCount++
	} else {
		loader.ReloadUserDate(users, update.Message.From.ID)
	}

	return nil
}

func messages(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	processingUser(bot, update)
	messageLog(update)

	err := menu.MessageProcessing(bot, update)
	if err != nil {
		logAll.Print(err)
	}

	sendMembers(bot, update)

	if update.Message == nil {
		return
	}

	var msg tgbotapi.MessageConfig
	var nilMsg bool

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "post":
			a, err := subscriptions.GetGroupPost(update.Message.CommandArguments())
			if err == nil {
				if a[1][0] != "" || a[0][0] != "" {
					if a[0][0] != "" {
						a[0][0] += "\nЗакреплённая запись"
					}

					for i := len(a) - 1; i >= 0; i-- {
						if len(a[i][1]) > 4500 {
							a[i][1] = a[i][1][:4500] + "...\n\n>>> Достигнуто ограничение на размер сообщения, перейдите по ссылке в начале сообщения, если хотите дочитать. <<<"
						}

						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a[i][0]+"\n\n"+a[i][1]))
					}
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Всегда пожалуйста.")
				} else {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Группа не валидна.")
				}
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Группа не валидна.")
			}
		default:
			nilMsg = true
		}

		if !nilMsg {
			_, err := bot.Send(msg)
			if err != nil {
				logAll.Print("Невозможно отправить: ", err)
			}
		}
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
func sendMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil || update.Message.From.ID != loader.MyId {
		return
	}

	var message string

	switch update.Message.Command() {
	case "defaultgroup":
		message = subscriptions.ChangeDefaultGroup(update.Message.CommandArguments())
	case "admin":
		message += "/users <_ | all> - Выводит статистику по пользователям.\n" +
			"/groups <_ | all> - Выводит статистику по каналам.\n" +
			"/setmessage <текст> - Задаёт сообщение, которое будет отображаться вместо погоды.\n" +
			"/sendmelog <data | users | labels | sub> - Присылает файл с логами.\n" +
			"/sendall <текст> - Делает рассылку текста. \n" +
			"/reset - Завершает текущую сессию бота. \n" +
			"/defaultgroup <id группы> - Изменяет отслеживаемую группу."
	case "users":
		if update.Message.CommandArguments() == "all" {

			var count int
			for _, v := range users {
				count++
				message += loader.WriteUsers(v) + "\n\n"

				if (count % 10) == 0 {
					bot.Send(tgbotapi.NewMessage(loader.MyId, message))
					message = ""
				}
			}
		}

		message += "Количество пользователей: " + strconv.Itoa(usersCount)
	case "groups":
		if update.Message.CommandArguments() == "all" {
			for _, v := range chats {
				message += v + "\n\n"
			}

		}

		message += "Количество чатов: " + strconv.Itoa(chatsCount)
	case "sendall":
		if update.Message.CommandArguments() != "" {
			logAll.Print("Рассылаю всем: '" + update.Message.CommandArguments() + "'")

			for i := range users {
				_, err := bot.Send(tgbotapi.NewMessage(int64(i), update.Message.CommandArguments()))
				if err != nil {
					logAll.Print("Что-то пошло не так при рассылке ["+fmt.Sprint(i)+"]", err)
				}
			}
		}

		return
	case "setmessage":
		weather.CurrentWeather = update.Message.CommandArguments()
		logAll.Print("Обновлена строка температуры на: " + weather.CurrentWeather)
		message += "Готово!\n" + "'" + weather.CurrentWeather + "'"
	case "sendmelog":
		if update.Message.CommandArguments() == "data" ||
			update.Message.CommandArguments() == "users" ||
			update.Message.CommandArguments() == "labels" ||
			update.Message.CommandArguments() == "sub" {

			_, err := bot.Send(tgbotapi.NewMessage(loader.MyId, "Отправляю..."))
			if err != nil {
				logAll.Print("Что-то пошло не так при sendmelog", err)
			}

			var name string

			switch update.Message.CommandArguments() {
			case "data":
				name = timeToStart
			case "users":
				name = loader.UserFileName
			case "labels":
				name = customers.LabelsFile
			case "sub":
				name = "users_subscriptions.txt"
			}

			_, err = bot.Send(tgbotapi.NewDocumentUpload(loader.MyId, name))
			if err != nil {
				_, err = bot.Send(tgbotapi.NewMessage(loader.MyId, "Не удалось отправить файл"))
				if err != nil {
					logAll.Print("С отправкой файла всё плохо")
				}

				logAll.Print("Ошибка отправки файла лога:", err)
			}
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(loader.MyId, "Попробуй ещё раз ввести аргументы правильно\n"+
				"'data' - Файл полного лога\n"+
				"'users' - файл с пользователями\n"+
				"'labels' - файл с метками\n"+
				"'sub' - файл с подписками"))
			if err != nil {
				logAll.Print("Что-то пошло не так ", err)
			}
		}

		return
	default:
		return
	}

	_, err := bot.Send(tgbotapi.NewMessage(loader.MyId, message))
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
		logAll.Fatal("Бот в отпуске: ", err)
	}

	bot.Debug = false

	info, err := schedule.GetAllSchedule("GK", &gkDate, &lkDate)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(loader.MyId, "Всё плохо с GK"))
		logAll.Fatal("GK")
	} else {
		logAll.Print(info)
	}

	info, err = schedule.GetAllSchedule("LK", &gkDate, &lkDate)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(loader.MyId, "Всё плохо с LK"))
		logAll.Fatal("LK")
	} else {
		logAll.Print(info)
	}

	go func() {
		for {
			answer, err := schedule.ParseSchedule("GK", &gkDate, &lkDate)
			if err != nil {
				logAll.Print(err)
			} else {
				if answer != "" {
					logAll.Print(answer)
				}
			}

			answer, err = schedule.ParseSchedule("LK", &gkDate, &lkDate)
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
			err := weather.SearchWeather()
			if err != nil {
				logAll.Print(err)
			}

			time.Sleep(time.Minute)
		}
	}()

	logAll.Printf("Бот %s запущен.", bot.Self.UserName)

	_, err = bot.Send(tgbotapi.NewMessage(loader.MyId, "Я перезагрузился."))
	if err != nil {
		logAll.Print("Не смог отправить весточку повелителю.", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logAll.Fatal(err)
	}

	usersCount, err = loader.LoadUsers(users)
	if err != nil {
		log.Fatal(err)
	}

	err = loader.LoadChats(chats)
	if err != nil {
		logAll.Print(err)
	}

	err = loader.LoadUserGroup()
	if err != nil {
		logAll.Print(err)
	}

	err = loader.LoadSchedule()
	if err != nil {
		logAll.Print(err)
	}

	err = loader.LoadUsersSubscriptions()
	if err != nil {
		logAll.Print(err)
	}

	go func() {
		for {
			time.Sleep(7 * time.Minute)

			if !menu.FlagToRunner {
				return
			}

			err := loader.UpdateUserInfo(users)
			if err != nil {
				logAll.Print(err)
			}

			err = customers.UpdateUserLabels()
			if err != nil {
				logAll.Print(err)
			}

			err = loader.UpdateUserSubscriptions()
			if err != nil {
				logAll.Print(err)
			}
		}
	}()

	go func() {
		for a := subscriptions.GetNewPosts(); len(a) == 0 || (a[0][1] == "" && a[1][0] == ""); a = subscriptions.GetNewPosts() {
			time.Sleep(5 * time.Second)
		}

		logAll.Print("Удачно загрузилась парсилка.")

		for {
			a := subscriptions.GetNewPosts()
			if len(a) != 0 {
				if a[0][0] != "" {
					for i, b := range subscriptions.UsersNsuHelp {
						if b != 0 {
							for _, v := range a {
								if len([]byte(v[1])) > 4500 {
									v[1] = string([]byte(v[1][:4500])) + "\n\n>>> Достигнуто ограничение на размер сообщения, перейдите по ссылке в начале сообщения, если хотите дочитать. <<<"
								}
								bot.Send(tgbotapi.NewMessage(int64(i), v[0]+"\n\n"+v[1]))
							}
						}
					}

				}
			}

			time.Sleep(33 * time.Second)
		}
	}()

	for update := range updates {
		go messages(bot, update)
	}
}
