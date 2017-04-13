package menu

import (
	"TelegramBot/all_types"
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/loader"
	"TelegramBot/schedule"
	"TelegramBot/subscriptions"
	"TelegramBot/weather"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"strconv"
	"time"
)

var queue = make(map[int]queueType)

type queueType struct {
	command  string
	argument string
	button   string
}

// Кнопки возвращения
const (
	BackButtonText = "« Назад"
	MainButtonText = "« В начало"
)

const (
	tag_main               = "menu_main"
	tag_week               = "menu_week"
	tag_schedule           = "menu_schedule"
	tag_weather            = "menu_weather"
	tag_subscriptions      = "menu_subscriptions"
	tag_options            = "menu_options"
	tag_clear_labels       = "clear_labels"
	tag_show_labels        = "show_labels"
	tag_labels             = "menu_labels"
	set_new_group          = "setgroup"
	tag_delete             = "delete"
	tag_schedule_day       = "tag_schedule_day"
	tag_day                = "tag_day"
	tag_keyboard           = "keyboard"
	tag_user_subscriptions = "user_subscriptions"
	set_different_group    = "set_different_group"
	different_day          = "different_day"
	today                  = "today"
	tomorrow               = "tomorrow"
	faq                    = "faq"
	feedback               = "feedback"
)

var FlagToRunner = true

func MessageProcessing(bot *tgbotapi.BotAPI, update tgbotapi.Update) (err error) {
	if update.CallbackQuery != nil {
		return ProcessingCallback(bot, update)
	}

	if update.Message != nil {
		return ProcessingMessage(bot, update)
	}

	if update.InlineQuery != nil {
		all_types.Logger.Print("InlineQuery")
	}

	if update.ChosenInlineResult != nil {
		all_types.Logger.Print("ChosenInlineResult")
	}

	if update.ChannelPost != nil {
		all_types.Logger.Print("ChannelPost")
	}

	return
}

func ProcessingCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) (err error) {
	command, argument := customers.DecomposeQuery(update.CallbackQuery.Data)

	all_types.Logger.Print("[", update.CallbackQuery.From.ID, "] @"+update.CallbackQuery.From.UserName+" "+update.CallbackQuery.From.FirstName+" "+update.CallbackQuery.From.LastName+", MessageID: ", update.CallbackQuery.Message.MessageID, ", Запрос: "+command+" | "+argument)

	switch command {
	case tag_keyboard:
		text := "Не удалось активировать квалиатуру, попробуйсте чуть позже."
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)

		markup, err := MainKeyboard()
		if err == nil {
			msg.Text = "Клавиатура активирована."
			msg.ReplyMarkup = markup
		}

		bot.Send(msg)
	case feedback:
		queue[update.CallbackQuery.From.ID] = queueType{feedback, "", ""}

		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Наберите свой отзыв:"))
	case faq:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, FaqText())

		m := RowButtonBack(tag_options, true)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_user_subscriptions:
		subscriptions.ChangeGroupByDomain(argument, update.CallbackQuery.From.ID)
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Нажмите на группу, если хотите подписаться на рассылку")

		m := UniteMarkup(SubscriptionsMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case all_types.NsuFit:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Скоро")

		m := RowButtonBack(tag_subscriptions, true)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_labels:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Управление метками")
		m := UniteMarkup(LabelsMenu(), RowButtonBack(tag_schedule, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_delete:
		text, markup := StartDeleteLabel(argument, update.CallbackQuery.From.ID)
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)

		msg.ReplyMarkup = &markup

		bot.Send(msg)
	case tag_clear_labels:
		text := customers.DeleteUserLabels(update.CallbackQuery.From.ID)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		m := UniteMarkup(LabelsMenu(), RowButtonBack(tag_schedule, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_show_labels:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, customers.PrintUserLabels(update.CallbackQuery.From.ID))
		m := RowButtonBack(tag_labels, true)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_schedule_day:
		g := ShowLabelsButton(tag_day+" "+argument+" ", update.CallbackQuery.From.ID)
		if len(g.InlineKeyboard) == 0 {
			g = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить метку", set_different_group+" "+tag_schedule_day+" "+argument)))
		}

		markup := UniteMarkup(g, tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ввести другой номер", different_day+" "+argument))),
			RowButtonBack(tag_schedule+" "+argument, true))

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Выберите группу")
		msg.ReplyMarkup = &markup

		bot.Send(msg)
		return
	case tag_day:
		day, group := customers.DecomposeQuery(argument)
		offset := Day(day)

		text, _ := schedule.PrintSchedule(group, offset, update.CallbackQuery.From.ID, false)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		m := RowButtonBack(tag_schedule_day+" "+day, true)

		msg.ReplyMarkup = &m

		bot.Send(msg)
	case different_day:
		queue[update.CallbackQuery.From.ID] = queueType{tag_day, argument + " ", ""}

		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите номер группы"))
	case tag_week:
		g, ok := all_types.AllLabels[update.CallbackQuery.From.ID]
		if !ok || g.MyGroup == "" {
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Вы не указали свою группу")
			m := UniteMarkup(tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить метку", set_different_group+" "+tag_schedule))),
				RowButtonBack(tag_schedule, true))

			msg.ReplyMarkup = &m

			bot.Send(msg)
			return
		}

		var msg tgbotapi.MessageConfig

		days := schedule.GetWeek(g.MyGroup)
		if len(days) > 0 {
			for i := 0; i < 6; i++ {
				bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, days[i]))
			}

			msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Готово")
		} else {
			msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Произошла ошибка, сообщите об этом мне /feedback, если ошибка появляется")
			bot.Send(tgbotapi.NewMessage(all_types.MyId, "Проблема с расписанием на неделю у группы "+g.MyGroup))
		}

		m := UniteMarkup(WeekMenu(), RowButtonBack(tag_schedule, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
		return
	case tag_main:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Главное меню")

		m := MainMenu()
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_options:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Дополнительные функции")

		m := UniteMarkup(OptionsMenu(), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_weather:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, weather.CurrentWeather)

		m := RowButtonBack(tag_main, false)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_schedule:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Расписание")

		m := UniteMarkup(ScheduleMenu(), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_subscriptions:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Нажмите на группу, если хотите подписаться на рассылку")

		m := UniteMarkup(SubscriptionsMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case set_new_group:
		text, markup := AddNewGroup(argument, tag_labels, update.CallbackQuery.From.ID, "")
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)

		if len(markup.InlineKeyboard) > 0 {
			msg.ReplyMarkup = markup
		}

		bot.Send(msg)
	case set_different_group:
		text, markup := AddNewGroup("", argument, update.CallbackQuery.From.ID, "Введите номер своей группы")
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)

		if len(markup.InlineKeyboard) > 0 {
			msg.ReplyMarkup = markup
		}

		bot.Send(msg)
	default:
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"Упс! Произошла ошибка, попробуйсте повторить операцию.")

		m := MainMenu()
		msg.ReplyMarkup = &m

		bot.Send(msg)
	}

	return
}

func Day(day string) int {
	switch day {
	case today:
		return 0
	case tomorrow:
		return 1
	default:
		return 0
	}
}

func AddNewGroup(argument string, back string, id int, myText string) (text string, markup tgbotapi.InlineKeyboardMarkup) {
	if argument == "" {
		if myText == "" {
			text = "Если вы хотите добавить свою группу в избранное, то введите её номер.\n\nЕсли вы хотите добавить/изменить метку, то введите номер группы и название метки через пробел:"
		} else {
			text = myText
		}
		queue[id] = queueType{set_new_group, "", back}
		return
	}

	var check int

	check, text = customers.AddGroupNumber(id, argument)

	switch check {
	case 0:
		queue[id] = queueType{set_new_group, "", back}
	case 1:
		markup = RowButtonBack(back, true)
		return
	case 2:
		text = "Вы достигли предела меток"
		markup = RowButtonBack(back, true)
	}

	return
}

func StartDeleteLabel(argument string, id int) (text string, markup tgbotapi.InlineKeyboardMarkup) {
	text = "Нажмите на метки, которые хотите удалить"

	if argument != "" {
		v := all_types.AllLabels[id]

		if argument == v.MyGroup {
			v.MyGroup = ""

			all_types.AllLabels[id] = v
		} else {
			delete(all_types.AllLabels[id].Group, argument)
		}
	}

	g := ShowLabelsButton(tag_delete+" ", id)

	if len(g.InlineKeyboard) > 0 {
		markup = UniteMarkup(g, RowButtonBack(tag_labels, true))
	} else {
		text = "Список меток пуст"
		markup = RowButtonBack(tag_labels, true)
	}

	return
}

func ProcessingMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) (err error) {
	var command string
	var argument string
	var button string

	if update.Message.IsCommand() {
		command = update.Message.Command()
		argument = update.Message.CommandArguments()
	} else {
		q := queue[update.Message.From.ID]
		command = q.command
		argument = q.argument + update.Message.Text
		button = q.button
	}

	all_types.Logger.Print("[", update.Message.From.ID, "] @"+update.Message.From.UserName+" "+update.Message.From.FirstName+" "+update.Message.From.LastName+", Команда: "+command, " | "+argument)

	queue[update.Message.From.ID] = queueType{"", "", ""}

	switch command {
	case feedback:
		if argument != "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо за отзыв!")

			msg.ReplyMarkup = RowButtonBack(tag_options, true)
			bot.Send(msg)

			bot.Send(tgbotapi.NewMessage(all_types.MyId, argument+"\n\nОтзыв от: ["+fmt.Sprint(update.Message.From.ID)+"]\n@"+update.Message.From.UserName+"\n"+update.Message.From.LastName+" "+update.Message.From.FirstName))

			return
		}

		queue[update.Message.From.ID] = queueType{feedback, "", ""}
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Наберите свой отзыв:"))

		return
	case "creator", "maker", "author", "father", "Creator", "Maker", "Author", "Father":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я в телеграм: @Dimonchik0036\nЯ на GitHub: github.com/dimonchik0036\nЯ в VK: vk.com/dimonchik0036"))
	case "weather":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, weather.CurrentWeather))
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Привет!\nЯ - твой помощник, сейчас я покажу тебе, что могу. Советую сразу включить /keyboard, чтобы было проще возвращаться к меню.\nЕщё есть полезные советы /help и /faq.")

		msg.ReplyMarkup = MainMenu()

		bot.Send(msg)
	case "help", "h":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, GetHelp(argument)))
	case "keyboard", "k":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if argument == "off" {
			msg.Text = "Клавиатура отключена."
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		} else {
			markup, err := MainKeyboard()
			if err == nil {
				msg.Text = "Клавиатура активирована."
				msg.ReplyMarkup = markup
			} else {
				msg.Text = "Не удалось активировать квалиатуру, попробуйсте чуть позже."
			}
		}

		bot.Send(msg)
	case "Меню", "menu":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Главное меню")
		msg.ReplyMarkup = MainMenu()

		bot.Send(msg)
	case tag_day:
		day, group := customers.DecomposeQuery(argument)
		offset := Day(day)

		text, ok := schedule.PrintSchedule(group, offset, update.Message.From.ID, true)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

		if !ok {
			queue[update.Message.From.ID] = queueType{tag_day, day + " ", ""}
		} else {
			msg.ReplyMarkup = RowButtonBack(tag_schedule_day+" "+day, true)
		}

		bot.Send(msg)
	case set_new_group:
		if button == "" {
			button = tag_labels
		}

		text, markup := AddNewGroup(argument, button, update.Message.From.ID, "")

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

		if len(markup.InlineKeyboard) > 0 {
			msg.ReplyMarkup = markup
		}

		bot.Send(msg)
	case "labels":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, customers.PrintUserLabels(update.Message.From.ID)))
	case "clearlabels":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, customers.DeleteUserLabels(update.Message.From.ID)))
	case "delete":
		delete(all_types.AllLabels[update.Message.From.ID].Group, argument)
	case "joke", "j":
		joke, err := jokes.GetJokes()
		if err == nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, joke))
		}
	case "faq":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, FaqText()))
		return
	}

	if update.Message.From.ID == all_types.MyId {
		adminMessage(bot, update.Message.Chat.ID, command, argument)
	}

	return
}

// sendMembers Отправляет статистику по пользователям
func adminMessage(bot *tgbotapi.BotAPI, where int64, command string, argument string) {
	switch command {
	case "admin":
		bot.Send(tgbotapi.NewMessage(where, "/users <_ | all> - Выводит статистику по пользователям\n"+
			"/groups <_ | all> - Выводит статистику по каналам\n"+
			"/setmessage <текст> - Задаёт сообщение, которое будет отображаться вместо погоды\n"+
			"/sendmelog <data | users | labels | sub> - Присылает файл с логами\n"+
			"/sendall <текст> - Делает рассылку текста\n"+
			"/reset - Завершает текущую сессию бота\n"+
			"/addnewgs <domain> - Добавляет группу к парсу\n"+
			"/delgroup <domain> - Удаляет группу из списка парсинга\n"+
			"/showgl - Показывает данные по подпискам\n"+
			"/changeus <domain + id> - Изменяет подписку пользователю\n"+
			"/activateg <domain> - Разрешает/запрещает парсинг группы\n"+
			"/statg <domain> - Выводит статистику пользователей по этой группе\n"+
			"/sendbyid <id + text> - Отправляет сообщение пользователю."))
		return
	case "reset":
		bot.Send(tgbotapi.NewMessage(where, "Выключаюсь."))

		go func() {
			FlagToRunner = false
			time.Sleep(5 * time.Second)

			customers.UpdateUserLabels()
			loader.UpdateUserInfo()
			loader.UpdateUserSubscriptions()

			os.Exit(0)
		}()

		return
	case "sendbyid":
		sId, text := customers.DecomposeQuery(argument)
		id, err := strconv.Atoi(sId)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, "Ошибка перевода числа"))
			return
		}

		if text == "" {
			bot.Send(tgbotapi.NewMessage(where, "Текст отсутствует"))
			return
		}

		_, err = bot.Send(tgbotapi.NewMessage(int64(id), text))
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, "При отправке произошла ошибка: "+err.Error()))
		}

		return
	case "users":
		var message string

		if argument == "all" {
			var count int
			for _, v := range all_types.AllUsersInfo {
				count++
				message += loader.WriteUsers(v) + "\n\n"

				if (count % 15) == 0 {
					bot.Send(tgbotapi.NewMessage(where, message))
					message = ""
				}
			}
		}

		message += "Количество пользователей: " + strconv.Itoa(all_types.UsersCount)

		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "groups":
		var message string
		if argument == "all" {
			for _, v := range all_types.AllChatsInfo {
				message += v + "\n\n"
			}

		}

		message += "Количество чатов: " + strconv.Itoa(all_types.ChatsCount)

		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "sendall":
		if argument != "" {
			all_types.Logger.Print("Рассылаю всем: '" + argument + "'")

			for i := range all_types.AllUsersInfo {
				_, err := bot.Send(tgbotapi.NewMessage(int64(i), argument))
				if err != nil {
					all_types.Logger.Print("Что-то пошло не так при рассылке ["+fmt.Sprint(i)+"]", err)
				}
			}
		}

		return
	case "setmessage":
		weather.CurrentWeather = argument
		all_types.Logger.Print("Обновлена строка температуры на: " + weather.CurrentWeather)

		bot.Send(tgbotapi.NewMessage(where, "Готово!\n"+"'"+weather.CurrentWeather+"'"))

		return
	case "sendmelog":
		if argument == "data" ||
			argument == "users" ||
			argument == "labels" ||
			argument == "sub" {

			_, err := bot.Send(tgbotapi.NewMessage(where, "Отправляю..."))
			if err != nil {
				all_types.Logger.Print("Что-то пошло не так при sendmelog", err)
			}

			var name string

			switch argument {
			case "data":
				name = all_types.LoggerFilename
			case "users":
				name = all_types.UsersFilename
			case "labels":
				name = all_types.LabelsFilename
			case "sub":
				name = all_types.SubscriptionsFilename
			}

			_, err = bot.Send(tgbotapi.NewDocumentUpload(where, name))
			if err != nil {
				_, err = bot.Send(tgbotapi.NewMessage(where, "Не удалось отправить файл"))
				if err != nil {
					all_types.Logger.Print("С отправкой файла всё плохо")
				}

				all_types.Logger.Print("Ошибка отправки файла лога:", err)
			}
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(where, "Попробуй ещё раз ввести аргументы правильно\n"+
				"'data' - Файл полного лога\n"+
				"'users' - файл с пользователями\n"+
				"'labels' - файл с метками\n"+
				"'sub' - файл с подписками"))
			if err != nil {
				all_types.Logger.Print("Что-то пошло не так ", err)
			}
		}

		return
	case "addnewgs":
		err := subscriptions.AddNewGroupToParse(argument)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, err.Error()))
		} else {
			bot.Send(tgbotapi.NewMessage(where, "Выполнено"))
		}
	case "showgl":
		g := subscriptions.ShowAllGroups()
		var message string
		for _, m := range g {
			message += m + "\n"
		}

		bot.Send(tgbotapi.NewMessage(where, message))
	case "changeus":
		domain, sid := customers.DecomposeQuery(argument)
		id, err := strconv.Atoi(sid)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, "Ошибка перевода числа"))
			return
		}

		message := subscriptions.ChangeGroupById(domain, id)
		bot.Send(tgbotapi.NewMessage(where, message))
	case "activateg":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.ChangeGroupActivity(argument)))
	case "delgroup":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.DeleteGroup(argument)))
	case "statg":
		m := subscriptions.ShowAllUsersGroup(argument)
		var count int
		var message string

		for _, v := range m {
			count++
			message += v + "\n"

			if (count % 20) == 0 {
				bot.Send(tgbotapi.NewMessage(where, message))
				message = ""
			}
		}

		bot.Send(tgbotapi.NewMessage(where, message+"\nВсего пользователей: "+strconv.Itoa(count)))

		return
	}
}

func RowButtonBack(href string, main bool) tgbotapi.InlineKeyboardMarkup {
	var row []tgbotapi.InlineKeyboardButton

	if href != "" {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(BackButtonText, href))
	}

	if main {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(MainButtonText, tag_main))
	}

	return tgbotapi.NewInlineKeyboardMarkup(row)
}

func UniteMarkup(markups ...tgbotapi.InlineKeyboardMarkup) (markup tgbotapi.InlineKeyboardMarkup) {
	for _, m := range markups {
		for _, v := range m.InlineKeyboard {
			markup.InlineKeyboard = append(markup.InlineKeyboard, v)
		}
	}

	return
}

func MainMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Температура", tag_weather)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Расписания", tag_schedule)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подписки", tag_subscriptions)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Дополнительно", tag_options)))

	return
}

func OptionsMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Включить клавиатуру", tag_keyboard)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", feedback)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("FAQ", faq)))

	return
}

func LabelsMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Показать все метки", tag_show_labels)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить/изменить метку", set_new_group)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить метку", tag_delete)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Очистить все метки", tag_clear_labels)))

	return
}

func ScheduleMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("На сегодня", tag_schedule_day+" "+today), tgbotapi.NewInlineKeyboardButtonData("На завтра", tag_schedule_day+" "+tomorrow)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("На всю неделю", tag_week)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление метками", tag_labels)))

	return
}

func WeekMenu() (markup tgbotapi.InlineKeyboardMarkup) {

	return
}

func ShowLabelsButton(prefix string, id int) (markup tgbotapi.InlineKeyboardMarkup) {
	v, ok := all_types.AllLabels[id]
	if !ok {
		return
	}

	if v.Group != nil {
		for l := range v.Group {
			markup.InlineKeyboard = append(markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(l, prefix+l)))
		}
	}

	if v.MyGroup != "" {
		markup.InlineKeyboard = append(markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моя группа", prefix+v.MyGroup)))
	}

	return
}

func MainKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup, err error) {
	keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/Меню")))
	return
}

func CheckSub(domain string, id int) string {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		return "⚠️"
	}

	u, ok := s.UserSubscriptions[id]
	if !ok {
		return ""
	}

	if u == 0 {
		return ""
	} else {
		return "☑️"
	}
}

func SubscriptionsMenu(id int) (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuSecret, id)+"Подслушано НГУ", tag_user_subscriptions+" "+all_types.NsuSecret)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuLove, id)+"Признавашки НГУ", tag_user_subscriptions+" "+all_types.NsuLove)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuHelp, id)+"Помогу в НГУ", tag_user_subscriptions+" "+all_types.NsuHelp)))
	return
}

func GetHelp(arg string) (text string) {
	switch arg {
	case "setgroup":
		text = "Раздел управления меток находится /menu » Расписание » Управление метками."
	case today, tomorrow:
		text = "Достаточно в /menu выбрать пункт Расписание и далее следовать по зову сердца."
	case "secret":
		text = "ACHTUNG! Использование этих команд запрещено на территории РФ. Автор ответственности не несёт, используйте на свой страх и риск. \n\n" +
			"/joke - Показывает бородатый анекдот.\n" +
			"/post <ID группы в VK> - Показывает закреплённый и 4 обычных поста из этой группы VK.\n\n" +
			"/creator - Используешь » ? » PROFIT!"
	default:
		text = "Подсказки по использованию Помощника:\n\n" +
			"Если вы интересуетесь расписанием занятий, то вам будет удобно добавить группы в избранное (далее метки), " +
			"это позволит вызывать расписание без особых усилий.\n" +
			"Раздел управления меток находится /menu » Расписание » Управление метками.\n\n" +
			"Ответы на дополнительные вопросы можно получить через /faq."
	}

	return text
}

func FaqText() string {
	return "Для тех, кому /help мало.\n\n" +
		"Q: Как установить метку на свою группу?\n" +
		"A: Воспользоваться /menu » Расписание » Управление метками » Добавить метку.\n" +
		"После чего ввести номер своей группы.\n\n" +

		"Q: Как установить несколько меток для разных групп?\n" +
		"A: Воспользоваться /menu » Расписание » Управление метками » Добавить метку.\n" +
		"Далее ввести номер группы и желаемое название метки, после этого она появится в списке.\n\n" +

		"Q: Можно ли посмотреть расписание, если не работает официальный сайт с расписанием?\n" +
		"A: Да, можно.\n\n" +

		"Q: Как часто обновляется расписание?\n" +
		"A: Сразу же после изменений в официальном расписании.\n\n" +

		"Если у Вас остались ещё какие-то вопросы, то со мной можно связаться через контакты /author."
}
