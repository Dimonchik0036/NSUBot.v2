package menu

import (
	"TelegramBot/customers"
	"TelegramBot/schedule"
	"TelegramBot/subscriptions"
	"TelegramBot/weather"
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"regexp"
)

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
const today_label = "today_label"
const tomorrow_label = "tomorrow_label"
const today_text = "Расписание на сегодня"
const tomorrow_text = "Расписание на завтра"

func MessageProcessing(update tgbotapi.Update) (answer tgbotapi.Chattable, err error) {
	if update.CallbackQuery != nil {
		log.Print("CallbackQuery: ", update.CallbackQuery.Data, " ID: ", update.CallbackQuery.From.ID)

		d, label := ScheduleCheck(update.CallbackQuery.Data)
		if label != "" {
			msg := tgbotapi.NewEditMessageText(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				schedule.PrintSchedule(label, update.CallbackQuery.From.ID, d))

			markup, err := BackDayButton(d)
			if err != nil {
				return nil, err
			}

			msg.ReplyMarkup = &markup
			answer = msg
		} else {
			switch update.CallbackQuery.Data {
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
			case tag_clear_labels:
				_, markup, err := OptionsMenu(tag_main)
				if err != nil {
					break
				}

				text := customers.DeleteUserLabels(update.CallbackQuery.From.ID)

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
				msg.ReplyMarkup = &markup
				answer = msg
			case tag_show_labels:
				_, markup, err := OptionsMenu(tag_main)
				if err != nil {
					break
				}

				text := customers.PrintUserLabels(update.CallbackQuery.From.ID)

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
				msg.ReplyMarkup = &markup
				answer = msg
			case tag_today:
				text, markup, err := DayMenu(tag_schedule, customers.AllLabels[update.CallbackQuery.From.ID], 0)
				if err != nil {
					break
				}

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
				msg.ReplyMarkup = &markup
				answer = msg
			case tag_tomorrow:
				text, markup, err := DayMenu(tag_schedule, customers.AllLabels[update.CallbackQuery.From.ID], 1)
				if err != nil {
					break
				}

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
				msg.ReplyMarkup = &markup
				answer = msg
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
			}
		}
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
		log.Print("Message")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			/*case "weather":
				answer = tgbotapi.NewMessage(update.Message.Chat.ID, weather.CurrentWeather)
			case tag_schedule:
				text, markup, err := ScheduleMenu(oldMenu string)
				if err != nil {
					break
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
				msg.ReplyMarkup = &markup
				answer = msg*/
			case "start":
				markup, err := MainKeyboard()
				if err == nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Приветсвенное сообщение.")
					msg.ReplyMarkup = markup

					answer = msg
				}
			case "keyboard":
				markup, err := MainKeyboard()
				if err == nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Клавиатура активирована.")
					msg.ReplyMarkup = markup

					answer = msg
				}
			case tag_main, "Меню", "menu":
				text, markup, err := MainMenu()
				if err == nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					msg.ReplyMarkup = markup

					answer = msg
				}
			}
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
			tgbotapi.NewInlineKeyboardButtonData("Показать все метки", tag_show_labels)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Очистить все метки", tag_clear_labels)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu)))

	text = "Дополнительные функции"

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

	text = "В разработке Ø"

	return
}

func DayMenu(oldMenu string, labels customers.UserGroup, day int) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	var date string

	switch day {
	case 0:
		text = today_text
		date = today_label
	case 1:
		text = tomorrow_text
		date = tomorrow_label
	default:
		err = errors.New("Ошибка нумерации дня.")
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	if labels.MyGroup != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моё", date+labels.MyGroup)))
	}

	for l, g := range labels.Group {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(l, date+g)))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu),
		tgbotapi.NewInlineKeyboardButtonData(MainButtonText, tag_main)))

	markup.InlineKeyboard = rows
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

func SubscriptionsMenu(oldMenu string) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помогу в НГУ", subscriptions.NsuHelp)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сайт ФИТ НГУ", subscriptions.NsuFit)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BackButtonText, oldMenu)))

	text = "Доступные подписки"
	return
}

func ScheduleCheck(command string) (d int, label string) {
	todayReg, err := regexp.Compile(today_label)
	if err != nil {
		return
	}

	tomorrowReg, err := regexp.Compile(tomorrow_label)
	if err != nil {
		return
	}

	index := todayReg.FindStringIndex(command)
	if len(index) > 0 && index[0] == 0 {
		label = command[index[1]:]
		d = 0

		return
	}

	index = tomorrowReg.FindStringIndex(command)
	if len(index) > 0 && index[0] == 0 {
		label = command[index[1]:]
		d = 0

		return
	}

	return
}
