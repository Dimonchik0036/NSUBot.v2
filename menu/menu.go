package menu

import (
	"TelegramBot/all_types"
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/loader"
	/*"TelegramBot/schedule"*/
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
	tag_support            = "menu_support"
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
	tag_fit                = "menu_fit"
	set_different_group    = "set_different_group"
	different_day          = "different_day"
	today                  = "today"
	tomorrow               = "tomorrow"
	faq                    = "faq"
	help                   = "help"
	feedback               = "feedback"
	vote                   = "vote"
	voteYes                = "yes"
	voteNo                 = "no"
	voteFit                = "fit"
)

type Vote struct {
	UserId int

	// 0 - Not voted
	// 1 - Yes
	// -1 - No
	// 2... Other
	Answer int
	Text   string
}

var voteArray []Vote

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

	/*if loader.ReloadUserDate(bot, *update.CallbackQuery.From) != nil {
		all_types.Logger.Print("Не удалось найти пользователя")
	}*/

	switch command {
	case vote:
		u, ok := all_types.AllUsersInfo[update.CallbackQuery.From.ID]
		if ok {
			bot.Send(tgbotapi.NewMessage(all_types.MyId, "Вердикт: "+argument+"\n"+u.String()))
		} else {
			bot.Send(tgbotapi.NewMessage(all_types.MyId, "Пользователь: "+argument+"\n\nСтранный юзер "+fmt.Sprint(update.Message.From.ID)))
		}

		switch argument {
		case voteYes:
			voteArray = append(voteArray, Vote{update.CallbackQuery.From.ID, 1, ""})
		case voteNo:
			voteArray = append(voteArray, Vote{update.CallbackQuery.From.ID, -1, ""})
		case voteFit:
			voteArray = append(voteArray, Vote{update.CallbackQuery.From.ID, 2, ""})
		default:
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Опрос уже окончен.")

			m := UniteMarkup(RowButtonBack(tag_main, false))
			msg.ReplyMarkup = &m
			return nil
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Спасибо за участие в опросе")
		m := UniteMarkup(RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		_, err := bot.Send(msg)
		if err != nil {
			all_types.Logger.Println(update.Message.From.ID, err)
		}
	case tag_fit:
		var m tgbotapi.InlineKeyboardMarkup

		if argument != "" {
			firstPar, secondPar := customers.DecomposeQuery(argument)
			if firstPar == all_types.News_chairs {
				if secondPar != "" {
					subscriptions.ChangeUserFit(firstPar+secondPar, update.CallbackQuery.From.ID)
				}

				m = UniteMarkup(ChairsMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_fit, true))
			} else {
				subscriptions.ChangeUserFit(argument, update.CallbackQuery.From.ID)

				m = UniteMarkup(FitMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_subscriptions, true))
			}
		} else {
			m = UniteMarkup(FitMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_subscriptions, true))
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Нажмите на раздел, если хотите подписаться на рассылку")
		msg.ReplyMarkup = &m

		bot.Send(msg)
		return
	case tag_keyboard:
		text := "Не удалось активировать квалиатуру, попробуйте чуть позже."
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)

		markup, err := MainKeyboard()
		if err == nil {
			msg.Text = "Клавиатура активирована."
			msg.ReplyMarkup = markup
		}

		bot.Send(msg)
	case tag_support:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Поддержка")

		m := UniteMarkup(SupportMenu(), RowButtonBack(tag_options, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case help:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, GetHelp(""))

		m := UniteMarkup(RowButtonBack(tag_support, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case feedback:
		queue[update.CallbackQuery.From.ID] = queueType{feedback, "", ""}

		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Наберите свой отзыв:"))
	case faq:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, FaqText())

		m := RowButtonBack(tag_support, true)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_user_subscriptions:
		if argument != "" {
			subscriptions.ChangeGroupByDomain(argument, update.CallbackQuery.From.ID)
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Нажмите на группу, если хотите подписаться на рассылку")

		m := UniteMarkup(VkGroupMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_subscriptions, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case all_types.NsuFit:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Скоро")

		m := RowButtonBack(tag_subscriptions, true)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	/*case tag_labels:
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
		msg.ParseMode = "HTML"
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
				m := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, days[i])
				m.ParseMode = "HTML"
				bot.Send(m)
			}

			msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Готово")
		} else {
			msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Произошла ошибка, сообщите об этом мне /feedback, если ошибка появляется")
			bot.Send(tgbotapi.NewMessage(all_types.MyId, "Проблема с расписанием на неделю у группы "+g.MyGroup))
		}

		m := UniteMarkup(WeekMenu(), RowButtonBack(tag_schedule, true))
		msg.ReplyMarkup = &m

		bot.Send(msg)
		return*/
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
	/*case tag_schedule:
	msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Расписание")

	m := UniteMarkup(ScheduleMenu(), RowButtonBack(tag_main, false))
	msg.ReplyMarkup = &m

	bot.Send(msg)*/
	case tag_subscriptions:
		if argument == all_types.NewsBot {
			subscriptions.ChangeBotSubscriptions(update.CallbackQuery.From.ID)
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Нажмите на группу, если хотите подписаться на рассылку")

		m := UniteMarkup(SubscriptionsMenu(update.CallbackQuery.From.ID), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	/*case set_new_group:
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

		bot.Send(msg)*/
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
	/*var button string*/

	if update.Message.IsCommand() {
		command = update.Message.Command()
		argument = update.Message.CommandArguments()
	} else {
		q := queue[update.Message.From.ID]
		command = q.command
		argument = q.argument + update.Message.Text
		/*button = q.button*/
	}

	all_types.Logger.Print("[", update.Message.From.ID, "] @"+update.Message.From.UserName+" "+update.Message.From.FirstName+" "+update.Message.From.LastName+", Команда: "+command, " | "+argument)

	/*if loader.ReloadUserDate(bot, *update.Message.From) != nil {
		all_types.Logger.Print("Не удалось найти пользователя")
	}*/

	queue[update.Message.From.ID] = queueType{"", "", ""}

	switch command {
	case "cansel":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Отменил все свои дела.\nZzzz..."))
		return
	case feedback:
		if argument != "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо за отзыв!")

			msg.ReplyMarkup = RowButtonBack(tag_support, true)
			bot.Send(msg)

			bot.Send(tgbotapi.NewMessage(all_types.MyId, argument+"\n\nОтзыв от: ["+fmt.Sprint(update.Message.From.ID)+"]\n@"+update.Message.From.UserName+"\n"+update.Message.From.LastName+" "+update.Message.From.FirstName))

			return
		}

		queue[update.Message.From.ID] = queueType{feedback, "", ""}
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Наберите свой отзыв:"))

		return
	case "botnews":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, subscriptions.ChangeBotSubscriptions(update.Message.From.ID)))
	case "creator", "maker", "author", "father", "Creator", "Maker", "Author", "Father":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я в телеграм: @Dimonchik0036\nЯ на GitHub: github.com/dimonchik0036\nЯ в VK: vk.com/dimonchik0036"))
	case "weather":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, weather.CurrentWeather))
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Привет!\n\nЯ - твой помощник, сейчас я покажу тебе, что могу. Советую сразу включить /keyboard, "+
				"чтобы было проще возвращаться к меню.\n\n"+
				"Ещё есть полезные советы /help и /faq.\n\n")

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
	/*case tag_day:
		day, group := customers.DecomposeQuery(argument)
		offset := Day(day)

		text, ok := schedule.PrintSchedule(group, offset, update.Message.From.ID, true)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ParseMode = "HTML"
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
		delete(all_types.AllLabels[update.Message.From.ID].Group, argument)*/
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
			"/sendallall <текст> - Делает рассылку текста абсолютно всем, игнорируя ограничение\n"+
			"/resetallusersub <YES> - Выставляет всем пользователям флаг на сообщения\n"+
			"/reset - Завершает текущую сессию бота\n"+
			"/addnewgs <domain> - Добавляет группу к парсу\n"+
			"/delgroup <domain> - Удаляет группу из списка парсинга\n"+
			"/deluser <id> - Удаляет пользователя\n"+
			"/showgl - Показывает данные по подпискам\n"+
			"/changeus <domain + id> - Изменяет подписку пользователю\n"+
			"/activateg <domain> - Разрешает/запрещает парсинг группы\n"+
			"/activatesend <domain> - Разрешает/запрещает рассылку группы\n"+
			"/statg <domain> - Выводит статистику пользователей по этой группе\n"+
			"/sendbyid <id + text> - Отправляет сообщение пользователю\n"+
			"/statsub - Отправляет количество пользователей, подписанных на новости бота\n"+
			"/addfit <href + title> - Добавляет раздел новостей\n"+
			"/changefit <href + id> - Подписывает пользователя на обновления\n"+
			"/delfit <href> - Удаляет группу фита\n"+
			"/showfit - Показывает группы\n"+
			"/fitactiv <href> - Активирует/деактивирует раздел\n"+
			"/fitstat <href> - Показывает статистику раздела\n\n"+
			"/showVoteStat < _ | all > - Показывает стату голосования\n"))
		return
	case "startVote":
		all_types.Logger.Print("Начинаю опрос")

		for i := range all_types.AllUsersInfo {
			m := tgbotapi.NewMessage(int64(i), "Опрос:\n<b>Хотели бы вы получать новости с сайта своего факультета?</b>")
			m.ReplyMarkup = VoteMenu()
			m.ParseMode = "HTML"

			_, err := bot.Send(m)
			if err != nil {
				all_types.Logger.Print("Что-то пошло не так при рассылке ["+fmt.Sprint(i)+"]", err)
			}
		}
	case "showVoteStat":
		if len(voteArray) == 0 {
			bot.Send(tgbotapi.NewMessage(where, "Ещё никто не проголосовал"))
			return
		}

		if argument != "all" {
			var yes, no, fit int
			for _, v := range voteArray {
				switch v.Answer {
				case 1:
					yes++
				case -1:
					no++
				case 2:
					fit++
				}
			}

			all := 50
			var yesStr, noStr, fitStr string
			for i := 0; i < yes*all/(fit+yes+no); i++ {
				yesStr += "*"
			}
			for i := 0; i < no*all/(fit+yes+no); i++ {
				noStr += "*"
			}
			for i := 0; i < fit*all/(fit+yes+no); i++ {
				fitStr += "*"
			}

			m := tgbotapi.NewMessage(where, "Опрос:\nВсего: "+fmt.Sprint(yes+no+fit)+"\n"+
				"Yes: "+fmt.Sprint(yes)+"\n"+yesStr+"\n"+
				"No: "+fmt.Sprint(no)+"\n"+noStr+"\n"+
				"Fit: "+fmt.Sprint(fit)+"\n"+fitStr+"\n")

			_, err := bot.Send(m)
			if err != nil {
				all_types.Logger.Print("Что-то пошло не так при рассылке мне", err)
			}
		} else {
			var message string
			for i, v := range voteArray {
				u, ok := all_types.AllUsersInfo[v.UserId]
				if !ok {
					all_types.Logger.Println("Не робит голос ", v.UserId)
					continue
				}

				message += u.String() + "\nГолос: "
				switch v.Answer {
				case 1:
					message += "yes"
				case -1:
					message += "no"
				case 2:
					message += "fit"
				}

				message += "\n\n"

				if ((i + 1) % 15) == 0 {
					bot.Send(tgbotapi.NewMessage(where, message))
					message = ""
				}
			}

			if message != "" {
				bot.Send(tgbotapi.NewMessage(where, message))
			}
		}
	case "resetallusersub":
		if argument == "YES" {
			for _, u := range all_types.AllUsersInfo {
				u.PermissionToSend = true
			}

			bot.Send(tgbotapi.NewMessage(where, "Готово, спамер"))
		} else {
			bot.Send(tgbotapi.NewMessage(where, "Будь осторожен"))
		}
		return
	case "changefit":
		href, sId := customers.DecomposeQuery(argument)
		id, err := strconv.Atoi(sId)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, "Ошибка перевода числа"))
			return
		}

		bot.Send(tgbotapi.NewMessage(where, subscriptions.ChangeUserFit(href, id)))
	case "delfit":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.DeleteFitNews(argument)))
	case "addfit":
		href, title := customers.DecomposeQuery(argument)
		bot.Send(tgbotapi.NewMessage(where, subscriptions.AddNewNewsList(href, title)))
	case "reset":
		bot.Send(tgbotapi.NewMessage(where, "Выключаюсь."))

		go func() {
			FlagToRunner = false
			time.Sleep(5 * time.Second)

			customers.UpdateUserLabels()
			loader.UpdateUserInfo()
			loader.UpdateUserSubscriptions()
			subscriptions.RefreshFitNsuFile()

			bot.Send(tgbotapi.NewMessage(where, "Вырубай проц"))

			os.Exit(0)
		}()

		return
	case "statsub":
		var count int
		for _, u := range all_types.AllUsersInfo {
			if u.PermissionToSend {
				count++
			}
		}

		bot.Send(tgbotapi.NewMessage(where, "Подписано на бота: "+fmt.Sprint(count)))
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
		m := tgbotapi.NewMessage(int64(id), text)
		m.ParseMode = "HTML"

		_, err = bot.Send(m)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, "При отправке произошла ошибка: "+err.Error()))
		} else {
			bot.Send(tgbotapi.NewMessage(where, "Успешно"))
		}

		return
	case "users":
		var message string

		if argument == "all" {
			var count int
			for _, v := range all_types.AllUsersInfo {
				count++
				message += v.String() + "\n" + "Рассылка: " + fmt.Sprint(v.PermissionToSend) + "\n\n"

				if (count % 15) == 0 {
					bot.Send(tgbotapi.NewMessage(where, message))
					message = ""
				}
			}
		}

		message += "Количество пользователей: " + strconv.Itoa(len(all_types.AllUsersInfo))

		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "groups":
		var message string
		if argument == "all" {
			for _, v := range all_types.AllChatsInfo {
				message += v + "\n\n"
			}

		}

		message += "Количество чатов: " + strconv.Itoa(len(all_types.AllChatsInfo))

		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "sendall":
		if argument != "" {
			all_types.Logger.Print("Рассылаю всем: '" + argument + "'")

			for i, u := range all_types.AllUsersInfo {
				if !u.PermissionToSend {
					continue
				}
				m := tgbotapi.NewMessage(int64(i), argument+"\n\nВы всегда можете отписаться от получения информации обновлений через /botnews")
				m.ParseMode = "HTML"

				_, err := bot.Send(m)
				if err != nil {
					all_types.Logger.Print("Что-то пошло не так при рассылке ["+fmt.Sprint(i)+"]", err)
				}
			}
		}

		return
	case "sendallall":
		if argument != "" {
			all_types.Logger.Print("Рассылаю всем: '" + argument + "'")

			for i := range all_types.AllUsersInfo {
				m := tgbotapi.NewMessage(int64(i), argument)
				m.ParseMode = "HTML"

				_, err := bot.Send(m)
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

		return
	case "showgl":
		g := subscriptions.ShowAllGroups()
		var message string
		for _, m := range g {
			message += m + "\n\n"
		}

		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "showfit":
		g := subscriptions.ShowAllFitNewsGroup()
		var message string
		for _, m := range g {
			message += m + "\n\n"
		}

		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "changeus":
		domain, sid := customers.DecomposeQuery(argument)
		id, err := strconv.Atoi(sid)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(where, "Ошибка перевода числа"))
			return
		}

		message := subscriptions.ChangeGroupById(domain, id)
		bot.Send(tgbotapi.NewMessage(where, message))

		return
	case "activateg":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.ChangeGroupActivity(argument)))
		return
	case "activatesend":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.GroupReady(argument)))
		return
	case "fitactiv":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.ChangeFitNewsActivity(argument)))
		return
	case "delgroup":
		bot.Send(tgbotapi.NewMessage(where, subscriptions.DeleteGroup(argument)))
		return
	case "deluser":
		bot.Send(tgbotapi.NewMessage(where, customers.DeleteUser(argument)))
		return
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
	case "fitstat":
		m := subscriptions.ShowAllFitUsersGroup(argument)
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

func VoteMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", vote+" "+voteYes)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Нет", vote+" "+voteNo)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Я с ФИТа, уже работает", vote+" "+voteFit)))
	return
}

func ChairsMenu(id int) (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_chairs+all_types.News_anksi, id)+"Кафедра систем информатики", tag_fit+" "+all_types.News_chairs+" "+all_types.News_anksi)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_chairs+all_types.News_ankks, id)+"Кафедра компьютерных систем", tag_fit+" "+all_types.News_chairs+" "+all_types.News_ankks)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_chairs+all_types.News_koinews, id)+"Кафедра общей информатики", tag_fit+" "+all_types.News_chairs+" "+all_types.News_koinews)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_chairs+all_types.News_kpvnews, id)+"Кафедра параллельных вычислений", tag_fit+" "+all_types.News_chairs+" "+all_types.News_kpvnews)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_chairs+all_types.News_kktnews, id)+"Кафедра компьютерных технологий", tag_fit+" "+all_types.News_chairs+" "+all_types.News_kktnews)))
	return
}

func FitMenu(id int) (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Кафедры", tag_fit+" "+all_types.News_chairs)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_announc, id)+"Объявления", tag_fit+" "+all_types.News_announc),
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_news, id)+"События", tag_fit+" "+all_types.News_news)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_konf, id)+"Конференции", tag_fit+" "+all_types.News_konf),
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_conc, id)+"Конкурсы", tag_fit+" "+all_types.News_conc)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckFit(all_types.News_admin_prikazy, id)+"Административные приказы", tag_fit+" "+all_types.News_admin_prikazy)))
	return
}

func MainMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Температура", tag_weather)),
		/*tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Расписания", tag_schedule)),*/
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Подписки", tag_subscriptions)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Дополнительно", tag_options)))

	return
}

func OptionsMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Включить клавиатуру", tag_keyboard)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Поддержка", tag_support)))

	return
}

func SupportMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("FAQ", faq)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подсказки", help)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", feedback)))

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
		return "☑️ "
	}
}

func CheckNews(id int) string {
	u, ok := all_types.AllUsersInfo[id]
	if !ok {
		return "⚠️"
	}

	if u.PermissionToSend {
		return "☑️ "
	} else {
		return ""
	}
}

func CheckFit(href string, id int) string {
	l, ok := subscriptions.FitNsuNews[href]
	if !ok {
		return "⚠️"
	}

	u, ok := l.Users[id]
	if !ok {
		return ""
	}

	if u == 0 {
		return ""
	} else {
		return "☑️ "
	}
}

func VkGroupMenu(id int) (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuSecret, id)+"Подслушано НГУ", tag_user_subscriptions+" "+all_types.NsuSecret)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuLove, id)+"Признавашки НГУ", tag_user_subscriptions+" "+all_types.NsuLove)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuHelp, id)+"Помогу в НГУ", tag_user_subscriptions+" "+all_types.NsuHelp)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.NsuTypical, id)+"Типичный НГУ*", tag_user_subscriptions+" "+all_types.NsuTypical)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckSub(all_types.Nsu24, id)+"Официальная группа НГУ", tag_user_subscriptions+" "+all_types.Nsu24)))
	return
}

func SubscriptionsMenu(id int) (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Группы VK", tag_user_subscriptions)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Новости ФИТ", tag_fit)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CheckNews(id)+"Новости об обновлении бота", tag_subscriptions+" "+all_types.NewsBot)))
	return
}

func GetHelp(arg string) (text string) {
	switch arg {
	/*case "setgroup":
		text = "Раздел управления меток находится /menu » Расписание » Управление метками."
	case today, tomorrow:
		text = "Достаточно в /menu выбрать пункт Расписание и далее следовать по зову сердца."*/
	case "secret":
		text = "ACHTUNG! Использование этих команд запрещено на территории РФ. Автор ответственности не несёт, используйте на свой страх и риск. \n\n" +
			"/joke - Показывает бородатый анекдот.\n" +
			//"/post <ID группы в VK> - Показывает закреплённый и 4 обычных поста из этой группы VK.\n\n" +
			"/creator - Используешь » ? » PROFIT!"
	default:
		text = "Подсказки по использованию Помощника:\n\n" +
			/*"Если вы интересуетесь расписанием занятий, то вам будет удобно добавить группы в избранное (далее метки), " +
			"это позволит вызывать расписание без особых усилий.\n" +
			"Раздел управления меток находится /menu » Расписание » Управление метками.\n\n" +*/
			"Расписание вернётся в сентябре, после начала учебного семестра.\n\n" +
			"Ответы на дополнительные вопросы можно получить через /faq.\n\n" +
			"Подписаться на новости об обвновлениях бота можно через /botnews или в /menu » Подписки"
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
