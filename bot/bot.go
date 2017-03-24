package main

import (
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/loader"
	"TelegramBot/nsuhelp"
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
var userGroup = make(map[int]customers.UserGroup)
var scheduleMap = make(map[string][7]string)

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

func getHelp(arg string) (text string) {
	switch arg {
	case "setgroup":
		text = "Команда позволяет назначить группу для быстрого доступа.\n" +
			"Например, если ввести \"/setgroup 16211.1\", то при использовании /today или /tomorrow без аргументов, будет показываться расписание группы 16211.1\n\n" +
			"Если ввести \"/setgroup <номер группы>\", то эта группа будет вызываться по умолчанию, " +
			"тоесть можно будет писать /today или /tomorrow без каких либо номеров групп.\n\n" +
			"Команда \"/setgroup <номер группы>  <метка>\" позволяет привязать группу к какой-то метке, " +
			"в качестве метки может выступать любая последовательность символов, не содержащая пробелов.\n" +
			"Чтобы воспрользоваться метками, достаточно ввести \"/today <метка>\" или \"/tomorrow <метка>\"."
	case "today", "tomorrow":
		text = "/today <номер группы | метка> - Показывает расписание занятий на сегодня, пример: \"/today 16211.1\".\n" +
			"/tomorrow <номер группы | метка> - Показывает расписание занятий на завтра, пример: \"/tomorrow 16211.1\".\n\n" +
			"Для вызова этих команд необходимо ввести номер группы. Если воспользоваться командой /setgroup, " +
			"то появится возможность использовать метки вместо номера группы, либо вовсе не писать ничего, " +
			"если добавить свою группу в стандартные (подробнее можно прочитать в \"/help setgroup\")."
	case "weather":
		text = "/weather - Показать температуру воздуха около НГУ."
	case "labels":
		text = "/labels - Показывает записанные метки."
	case "clearlabels":
		text = "/clearlabels - Очищает все метки, кроме стандартной."
	case "feedback":
		text = "/feedback <текст> - Оставить отзыв, который будет услышан."
	case "nsuhelp":
		text = "/nsuhelp - Управление подпиской на рассылку новостей из группы \"Помогу в НГУ\".\n\n" +
			"Позволяет подписаться на рассылку новых новостей из группы \"Помогу в НГУ\"."
	case "secret":
		text = "ACHTUNG! Использование этих команд запрещено на территории РФ. Автор ответственности не несёт, используйте на свой страх и риск. \n\n" +
			"/joke - Показывает бородатый анекдот.\n" +
			"/subjoke - Подписывает на рассылку бородатых анекдотов. Именно их можно получить, используя /joke\n" +
			"/post <ID группы в VK> - Показывает закреплённый и 4 обычных поста из этой группы VK.\n\n" +
			"/creator - Используешь -> ? -> PROFIT!"
	default:
		text = "Список команд:\n" +
			"/help - Показать список команд\n\n" +
			"/weather - Показать температуру воздуха около НГУ\n\n" +
			"/today <номер группы | метка> - Показывает расписание занятий конкретной группы.\n\n" +
			"/tomorrow <номер группы | метка> - Показывает расписание занятий конкретной группы на завтра.\n\n" +
			"/setgroup <номер группы + метка> - Устанавливает номер группы для быстрого доступа.\n\n" +
			"/labels - Показывает записанные метки.\n\n" +
			"/clearlabels - Очищает все метки, кроме стандартной.\n\n" +
			"/nsuhelp - Управление подпиской на рассылку новостей из группы \"Помогу в НГУ\".\n\n" +
			"/feedback <текст> - Оставить отзыв, который будет услышан.\n\n" +
			"/faq - Типичные вопросы и ответы на них.\n\n" +
			"Для подробного описания команд, введите \"/help <команда>\". Например, \"/help setgroup\".\n\n" +
			"P.S. Значёк <|> в расписании показывает, что это двойная пара. Отображение только текущей недели будет добавлено чуть позже."
	}

	return text
}

func processingMessages(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
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
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, getHelp(update.Message.CommandArguments()))
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
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Мой телеграм: @Dimonchik0036\nМой github: github.com/dimonchik0036")
		case "Погода", "weather", "погода", "Weather", "weather_nsu":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, weatherText)
		case "today":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, schedule.PrintSchedule(scheduleMap, userGroup, update.Message.CommandArguments(), 0, update.Message.From.ID))
		case "tomorrow":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, schedule.PrintSchedule(scheduleMap, userGroup, update.Message.CommandArguments(), 1, update.Message.From.ID))
		case "setgroup":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, customers.AddGroupNumber(scheduleMap, userGroup, update.Message.From.ID, update.Message.CommandArguments()))
		case "labels":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, customers.PrintUserLabels(userGroup[update.Message.From.ID].Group))
		case "clearlabels":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, customers.DeletUserLabels(userGroup[update.Message.From.ID]))
		case "joke":
			joke, err := jokes.GetJokes()
			if err == nil {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, joke)
			}
		case "subjoke":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, jokes.ChangeJokeSubscriptions(update.Message.From.ID))
		case "nsuhelp":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, nsuhelp.ChangeSubscriptions(update.Message.From.ID))
		case "post":
			a, err := nsuhelp.GetGroupPost(update.Message.CommandArguments())
			if err == nil {
				if a[0][0] != "" {
					for _, v := range a {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, v[0]+"\n\n"+v[1])
					}
				} else {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Группа не валидна.")
				}
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Группа не валидна.")
			}
		case "feedback":
			if update.Message.CommandArguments() != "" {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо за обратную связь.")
				bot.Send(tgbotapi.NewMessage(myId, update.Message.CommandArguments()+"\n\nОтзыв от:\n"+loader.WriteUsers(users[update.Message.From.ID])))
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы забыли набрать сообщение.")
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

		if update.Message.From.ID == myId {
			if update.Message.Command() == "reset" {
				bot.Send(tgbotapi.NewMessage(myId, "Выключаюсь."))
				logAll.Print("Выключаюсь по приказу.")
				return
			}

			sendMembers(update.Message.Command(), update.Message.CommandArguments(), bot)
		}
	}
}

func SendJokesAll(bot *tgbotapi.BotAPI) error {
	for {
		joke, err := jokes.GetJokes()
		if err == nil {
			for i, v := range jokes.JokeBase {
				if v != 0 {
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
	case "defaultgroup":
		message = nsuhelp.ChangeDefaultGroup(arg)
	case "help":
		message += "/users <_ | all> - Выводит статистику по пользователям.\n" +
			"/groups <_ | all> - Выводит статистику по каналам.\n" +
			"/setmessage <текст> - Задаёт сообщение, которое будет отображаться вместо погоды.\n" +
			"/sendmelog <data | users | labels> - Присылает файл с логами.\n" +
			"/sendall <текст> - Делает рассылку текста. \n" +
			"/reset - Завершает текущую сессию бота. \n" +
			"/defaultgroup <id группы> - Изменяет отслеживаемую группу."
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
		if arg == "data" || arg == "users" || arg == "labels" {
			_, err := bot.Send(tgbotapi.NewMessage(myId, "Отправляю..."))
			if err != nil {
				logAll.Print("Что-то пошло не так при sendmelog", err)
			}

			var name string

			switch arg {
			case "data":
				name = timeToStart
			case "users":
				name = loader.UserFileName
			case "labels":
				name = customers.LabelsFile
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
				"'data' - Файл полного лога.\n"+
				"'users' - файл с пользователями.\n"+
				"'labels' - файл с метками."))
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
		logAll.Fatal("Бот в отпуске: ", err)
	}

	bot.Debug = false

	info, err := schedule.GetAllSchedule(scheduleMap, "GK", &gkDate, &lkDate)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(myId, "Всё плохо с GK"))
		logAll.Fatal("GK")
	} else {
		logAll.Print(info)
	}

	info, err = schedule.GetAllSchedule(scheduleMap, "LK", &gkDate, &lkDate)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(myId, "Всё плохо с LK"))
		logAll.Fatal("LK")
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

	logAll.Printf("Бот %s запущен.", bot.Self.UserName)

	_, err = bot.Send(tgbotapi.NewMessage(myId, "Я перезагрузился."))
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

	loader.LoadChats(chats)
	loader.LoadUserGroup(scheduleMap, userGroup)
	loader.LoadSchedule(scheduleMap)
	loader.LoadUsersSubscriptions()

	go SendJokesAll(bot)

	go func() {
		for {
			time.Sleep(7 * time.Minute)

			loader.UpdateUserInfo(users)
			customers.UpdateUserLabels(userGroup)
			loader.UpdateUserSubscriptions()
		}
	}()

	go func() {
		for {
			a := nsuhelp.GetNewPosts()
			if len(a) != 0 {
				if a[0][0] != "" {
					for i, b := range nsuhelp.UsersNsuHelp {
						if b != 0 {
							for _, v := range a {
								if len(v[1]) > 4500 {
									v[1] = v[1][:4500] + "\n\n>>> Достигнуто ограничение на размер сообщения, перейдите по ссылке в начале сообщения, если хотите дочитать. <<<"
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
		if update.Message == nil {
			continue
		}

		go processingMessages(bot, update)
	}
}
