package menu

import (
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/loader"
	"TelegramBot/schedule"
	"TelegramBot/subscriptions"
	"TelegramBot/weather"
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"time"
)

var queue = make(map[int]queueType)

type queueType struct {
	run        bool
	showButton bool
	command    string
	oldMenu    string
	id         int
}

const BackButtonText = "« Назад"
const MainButtonText = "« В начало"

const tag_main = "menu_main"
const tag_today = "menu_today"
const tag_tomorrow = "menu_tomorrow"
const tag_week = "menu_week"
const tag_schedule = "menu_schedule"
const tag_weather = "menu_weather"
const tag_subscriptions = "menu_subscriptions"
const tag_options = "menu_options"
const tag_clear_labels = "clear_labels"
const tag_show_labels = "show_labels"
const tag_labels = "menu_labels"
const tag_usergroup = "tag_usergroup"
const tag_delete = "delete"
const schedule_today = "schedule_today"
const schedule_tomorrow = "schedule_tomorrow"
const different_today = "different_today"
const different_tomorrow = "different_tomorrow"
const today_text = "Расписание на сегодня:"
const tomorrow_text = "Расписание на завтра:"
const today = "today"
const tomorrow = "tomorrow"

var FlagToRunner = true

func ProcessingCallback(update tgbotapi.Update) (answer tgbotapi.Chattable, err error) {
	log.Print("CallbackQuery: ", update.CallbackQuery.Data, " ID: ", update.CallbackQuery.From.ID, " MessageID:", update.CallbackQuery.Message.MessageID)

	data := update.CallbackQuery.Data
	q, ok := queue[update.CallbackQuery.From.ID]

	if ok && data != q.oldMenu && data != tag_main && q.command != "" && q.id == update.CallbackQuery.Message.MessageID {
		data = q.command
	}

	queue[update.CallbackQuery.From.ID] = queueType{false, false, "", "", 0}

	switch data {
	case subscriptions.NsuHelp:
		_, markup, err := SubscriptionsMenu(tag_main)
		if err != nil {
			break
		}

		text := subscriptions.ChangeSubscriptions(update.CallbackQuery.From.ID, "Помогу в НГУ")

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case subscriptions.NsuFit:
		_, markup, err := SubscriptionsMenu(tag_main)
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Ещё в разработке")
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_labels:
		text, markup, err := LabelsMenu(tag_options)
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_delete:
		queue[update.CallbackQuery.From.ID] = queueType{true, true, "delete", tag_labels, update.CallbackQuery.Message.MessageID}
		v := customers.AllLabels[update.CallbackQuery.From.ID]

		if update.CallbackQuery.Data == customers.MyGroupLabel {
			v.MyGroup = ""

			customers.AllLabels[update.CallbackQuery.From.ID] = v
		} else {
			delete(customers.AllLabels[update.CallbackQuery.From.ID].Group, update.CallbackQuery.Data)
		}

		text, markup, err := DeleteMenu(tag_labels, customers.AllLabels[update.CallbackQuery.From.ID])
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_clear_labels:
		_, markup, err := LabelsMenu(tag_labels)
		if err != nil {
			break
		}

		text := customers.DeleteUserLabels(update.CallbackQuery.From.ID)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_show_labels:
		_, markup, err := LabelsMenu(tag_options)
		if err != nil {
			break
		}

		text := customers.PrintUserLabels(update.CallbackQuery.From.ID)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_today, tag_tomorrow:
		var day int
		var weekDay string

		switch data {
		case tag_tomorrow:
			weekDay = different_tomorrow
			day = 1
			queue[update.CallbackQuery.From.ID] = queueType{true, true, schedule_tomorrow, tag_tomorrow, update.CallbackQuery.Message.MessageID}
		case tag_today:
			weekDay = different_today
			day = 0
			queue[update.CallbackQuery.From.ID] = queueType{true, true, schedule_today, tag_today, update.CallbackQuery.Message.MessageID}
		}

		text, markup, err := DayMenu(tag_schedule, customers.AllLabels[update.CallbackQuery.From.ID], day)
		if err != nil {
			break
		}

		lastRow := markup.InlineKeyboard[len(markup.InlineKeyboard)-1]
		markup.InlineKeyboard[len(markup.InlineKeyboard)-1] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ввести самому", weekDay))
		markup.InlineKeyboard = append(markup.InlineKeyboard, lastRow)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case schedule_today, schedule_tomorrow:
		var day int

		switch data {
		case schedule_today:
			day = 0
		case schedule_tomorrow:
			day = 1
		}

		markup, err := BackDayButton(day)
		if err != nil {
			return nil, err
		}

		text, ok := schedule.PrintSchedule(update.CallbackQuery.Data, day, update.CallbackQuery.From.ID, false)
		if ok {
			msg := tgbotapi.NewEditMessageText(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				text)
			msg.ReplyMarkup = &markup

			answer = msg
		} else {
			var d string

			switch update.CallbackQuery.Data {
			case different_today:
				d = today
			case different_tomorrow:
				d = tomorrow
			}

			if d != "" {
				queue[update.CallbackQuery.From.ID] = queueType{true, true, d, tag_schedule, 0}

				answer = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите номер группы:")
			} else {
				text, markup, err := ScheduleMenu(tag_main)
				if err != nil {
					break
				}

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
				msg.ReplyMarkup = &markup

				answer = msg
			}
		}
	case tag_week:
		text, markup, err := WeekMenu(tag_schedule)
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_main:
		text, markup, err := MainMenu()
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_options:
		text, markup, err := OptionsMenu(tag_main)
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_weather:
		_, markup, err := MainMenu()
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, weather.CurrentWeather)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_schedule:
		text, markup, err := ScheduleMenu(tag_main)
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_subscriptions:
		text, markup, err := SubscriptionsMenu(tag_main)
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = &markup

		answer = msg
	case tag_usergroup:
		queue[update.CallbackQuery.From.ID] = queueType{true, true, "setgroup", "", 0}

		answer = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите номер группы и название метки:")
	default:
		_, markup, err := MainMenu()
		if err != nil {
			break
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"Упс! Произошла ошибка, попробуйсте повторить операцию.")
		msg.ReplyMarkup = &markup

		answer = msg
	}

	return
}

func MessageProcessing(update tgbotapi.Update) (answer tgbotapi.Chattable, err error) {
	if update.CallbackQuery != nil {
		answer, err = ProcessingCallback(update)
		return
	}

	if update.InlineQuery != nil {
		log.Print("InlineQuery")
	}

	if update.ChosenInlineResult != nil {
		log.Print("ChosenInlineResult")
	}

	if update.ChannelPost != nil {
		log.Print("ChannelPost")
	}

	if update.Message != nil {
		log.Print("MessageText: ", update.Message.Text, " ID:", update.Message.From.ID)

		command := update.Message.Command()
		arguments := update.Message.CommandArguments()

		if !update.Message.IsCommand() {
			command = queue[update.Message.From.ID].command
			arguments = update.Message.Text
		}

		q := queue[update.Message.From.ID]
		queue[update.Message.From.ID] = queueType{false, q.showButton, "", "", 0}

		switch command {
		case "creator", "maker", "author", "father", "Creator", "Maker", "Author", "Father":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID, "Мой телеграм: @Dimonchik0036\nМой GitHub: github.com/dimonchik0036")
		case "reset":
			if update.Message.From.ID == loader.MyId {
				answer = tgbotapi.NewMessage(loader.MyId, "Выключаюсь.")

				go func() {
					FlagToRunner = false
					time.Sleep(5 * time.Second)

					customers.UpdateUserLabels()
					loader.UpdateUserSubscriptions()

					os.Exit(0)
				}()
			}
		case "weather":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID, weather.CurrentWeather)
		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Привет!\nЯ - твой помощник, сейчас я покажу тебе, что могу.\n\n"+GetHelp(""))

			markup, err := MainKeyboard()
			if err == nil {
				msg.ReplyMarkup = markup
			}

			answer = msg
		case "help", "h":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, GetHelp(arguments))

			answer = msg
		case "keyboard", "k":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			if arguments == "off" {
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

			answer = msg
		case "Меню", "menu":
			text, markup, err := MainMenu()
			if err == nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
				msg.ReplyMarkup = markup

				answer = msg
			}
		case today, "t", "td", tomorrow, "tm", "tom":
			var day int
			switch command {
			case today, "t", "td":
				day = 0
			case tomorrow, "tm", "tom":
				day = 1
			}

			text, ok := schedule.PrintSchedule(arguments, day, update.Message.From.ID, false)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

			if ok {
				queue[update.Message.From.ID] = queueType{false, false, "", "", 0}

				if q.showButton {
					markup, err := BackDayButton(day)
					if err == nil {
						msg.ReplyMarkup = markup
					}
				}
			} else {
				queue[update.Message.From.ID] = queueType{true, q.showButton, command, "", 0}

				if !q.run {
					msg.Text = "Введите номер группы:"
				}
			}

			answer = msg
		case "setgroup":
			ok, text := customers.AddGroupNumber(schedule.TableSchedule, update.Message.From.ID, arguments)
			if ok {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

				if q.showButton {
					_, markup, err := LabelsMenu(tag_options)
					if err == nil {
						msg.ReplyMarkup = markup
					}
				}

				queue[update.Message.From.ID] = queueType{false, false, "", "", 0}

				answer = msg
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

				if text != "Предел" {
					queue[update.Message.From.ID] = queueType{true, q.showButton, command, "", 0}
				} else {
					queue[update.Message.From.ID] = queueType{false, false, "", "", 0}

					msg.Text = "Вы достигли предела меток. Теперь Вы можете только очистить список меток, воспользовавшись командой /clearlabels, " +
						"или изменять группы, привязанные к меткам, но не можете добавлять новые."

					if q.showButton {
						_, markup, err := LabelsMenu(tag_options)
						if err == nil {
							msg.ReplyMarkup = markup
							msg.Text = "Вы достигли предела меток. Теперь Вы можете только очистить список меток " +
								"или изменить группы, привязанные к меткам, но не можете добавлять новые."
						}
					}
				}

				if !q.run {
					text = "Введите номер группы и название метки:"
				}

				answer = msg
			}
		case "labels":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID, customers.PrintUserLabels(update.Message.From.ID))
		case "clearlabels":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID, customers.DeleteUserLabels(update.Message.From.ID))
		case "delete":
			delete(customers.AllLabels[update.Message.From.ID].Group, arguments)
		case "joke", "j":
			joke, err := jokes.GetJokes()
			if err == nil {
				answer = tgbotapi.NewMessage(update.Message.Chat.ID, joke)
			}
		case "subjoke":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID, jokes.ChangeJokeSubscriptions(update.Message.From.ID))
		case "nsuhelp":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID, subscriptions.ChangeSubscriptions(update.Message.From.ID, "Помогу в НГУ"))
		case "faq":
			answer = tgbotapi.NewMessage(update.Message.Chat.ID,
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
		}

	}

	if answer == nil {
		return nil, errors.New("Сообщение не прошло обработку.")
	}

	return
}

func MainMenu() (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Температура", tag_weather)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Расписания", tag_schedule)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подписки", tag_subscriptions)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Дополнительно", tag_options)))

	text = "Главное меню"

	return
}

func OptionsMenu(oldMenu string) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление метками", tag_labels)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu)))

	text = "Дополнительные функции"

	return
}

func LabelsMenu(oldMenu string) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Показать все метки", tag_show_labels)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить/изменить метку", tag_usergroup)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить метку", tag_delete)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Очистить все метки", tag_clear_labels)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu), tgbotapi.NewInlineKeyboardButtonData(MainButtonText, tag_main)))

	text = "Управление метками"

	return
}

func ScheduleMenu(oldMenu string) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("На сегодня", tag_today), tgbotapi.NewInlineKeyboardButtonData("На завтра", tag_tomorrow)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("На всю неделю", tag_week)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu)))

	text = "Расписание"

	return
}

func WeekMenu(oldMenu string) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu), tgbotapi.NewInlineKeyboardButtonData(MainButtonText, tag_main)))

	text = "В следующем обновлении."

	return
}

func DayMenu(oldMenu string, labels customers.UserGroup, day int) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	switch day {
	case 0:
		text = today_text
	case 1:
		text = tomorrow_text
	default:
		err = errors.New("Ошибка нумерации дня.")
	}

	markup.InlineKeyboard = ShowLabelsButton(oldMenu, labels, true)

	return
}

func ShowLabelsButton(oldMenu string, labels customers.UserGroup, group bool) (rows [][]tgbotapi.InlineKeyboardButton) {
	if group {
		if labels.MyGroup != "" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моё", labels.MyGroup)))
		}

		for l, g := range labels.Group {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(l, g)))
		}
	} else {
		if labels.MyGroup != "" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моё", customers.MyGroupLabel)))
		}

		for l := range labels.Group {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(l, l)))
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu),
		tgbotapi.NewInlineKeyboardButtonData(MainButtonText, tag_main)))

	return
}

func BackDayButton(d int) (markup tgbotapi.InlineKeyboardMarkup, err error) {
	var row []tgbotapi.InlineKeyboardButton

	switch d {
	case 0:
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(BackButtonText, tag_today))
	case 1:
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(BackButtonText, tag_tomorrow))
	default:
		err = errors.New("Ошибка генерации кнопки дня.")
		return
	}

	row = append(row, tgbotapi.NewInlineKeyboardButtonData(MainButtonText, tag_main))

	markup = tgbotapi.NewInlineKeyboardMarkup(row)
	return
}

func MainKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup, err error) {
	keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/Меню")))
	return
}

func DeleteMenu(oldMenu string, labels customers.UserGroup) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup.InlineKeyboard = ShowLabelsButton(oldMenu, labels, false)
	text = "Нажмите на метку, которую хотите удалить:"
	return
}

func SubscriptionsMenu(oldMenu string) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помогу в НГУ", subscriptions.NsuHelp)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сайт ФИТ НГУ", subscriptions.NsuFit)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu)))

	text = "Доступные подписки:"
	return
}

func GetHelp(arg string) (text string) {
	switch arg {
	case "setgroup":
		text = "Команда позволяет назначить группу для быстрого доступа.\n" +
			"Например, если ввести \"/setgroup 16211.1\", то при использовании /today или /tomorrow без аргументов, будет показываться расписание группы 16211.1\n\n" +
			"Если ввести \"/setgroup <номер группы>\", то эта группа будет вызываться по умолчанию, " +
			"тоесть можно будет писать /today или /tomorrow без каких либо номеров групп.\n\n" +
			"Команда \"/setgroup <номер группы>  <метка>\" позволяет привязать группу к какой-то метке, " +
			"в качестве метки может выступать любая последовательность символов, не содержащая пробелов.\n" +
			"Чтобы воспрользоваться метками, достаточно ввести \"/today <метка>\" или \"/tomorrow <метка>\"."
	case today, tomorrow:
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
