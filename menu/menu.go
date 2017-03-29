package menu

import (
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func MessageProcessing(update tgbotapi.Update) (answer tgbotapi.Chattable, err error) {
	if update.CallbackQuery != nil {
		log.Print("CallbackQuery: ", update.CallbackQuery.Data)

		switch update.CallbackQuery.Data {
		case "today", "today_friend":
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Вторник.\n"+
				"ФИТ\n"+
				"Группа 16211.1\n"+
				"1 П.  09:00: Алг.и геом (л), ауд. 402, Чуркин В.А.\n"+
				"2 П. 10:50: Ин.яз. (с), ауд. 1132, Савилова Т.К.\n"+
				"3 П. 12:40: -\n"+
				"4 П. 14:30: физв (л), ауд.\n"+
				"5 П. 16:20: -\n"+
				"6 П. 18:10: -\n"+
				"7 П. 20:00: -")
			markup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_today")))

			msg.ReplyMarkup = &markup
			answer = msg
		case "tomorrow", "tomorrow_friend":
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Вторник.\n"+
				"ФИТ\n"+
				"Группа 16211.1\n"+
				"1 П.  09:00: Алг.и геом (л), ауд. 402, Чуркин В.А.\n"+
				"2 П. 10:50: Ин.яз. (с), ауд. 1132, Савилова Т.К.\n"+
				"3 П. 12:40: -\n"+
				"4 П. 14:30: физв (л), ауд.\n"+
				"5 П. 16:20: -\n"+
				"6 П. 18:10: -\n"+
				"7 П. 20:00: -")
			markup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_tomorrow")))

			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_today":
			text, markup, err := DayMenu(0)
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_tomorrow":
			text, markup, err := DayMenu(1)
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_week":
			text, markup, err := WeekMenu()
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_main":
			text, markup, err := MainMenu()
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_weather":
			_, markup, err := MainMenu()
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, weather())
			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_schedule":
			markup, err := ScheduleMenu()
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, schedule())
			msg.ReplyMarkup = &markup
			answer = msg
		case "menu_subscriptions":
			text, markup, err := SubscriptionsMenu()
			if err != nil {
				break
			}

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
			msg.ReplyMarkup = &markup
			answer = msg
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
			case "Темп.":
				answer = tgbotapi.NewMessage(update.Message.Chat.ID, weather())
			case "Расп.", "menu_schedule":
				markup, err := ScheduleMenu()
				if err != nil {
					break
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, schedule())
				msg.ReplyMarkup = &markup
				answer = msg
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
			case "menu_start", "Меню", "menu":
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
			tgbotapi.NewInlineKeyboardButtonData("Температура", "menu_weather")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Расписание", "menu_schedule")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подписки", "menu_subscriptions")))

	text = "Главное меню"

	return
}

func ScheduleMenu() (markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сегодня", "menu_today"), tgbotapi.NewInlineKeyboardButtonData("Завтра", "menu_tomorrow")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вся неделя", "menu_week")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_main")))

	return
}

func WeekMenu() (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_main")))

	text = "Заглушка"

	return
}

func DayMenu(day int) (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	switch day {
	case 0:
		markup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Моё", "today")),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("*Типо твоего друга*", "today_friend")),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_schedule")))

		text = "Расписание на сегодня."
	case 1:
		markup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Моё", "tomorrow")),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("*Типо твоего друга*", "tomorrow_friend")),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_schedule")))

		text = "Расписание на завтра"
	default:
		err = errors.New("Ошибка нумерации дня.")
	}

	return
}

func MainKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup, err error) {
	keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/Меню")))
	return
}

func SubscriptionsMenu() (text string, markup tgbotapi.InlineKeyboardMarkup, err error) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "menu_main")))

	text = "Заглушка"
	return
}

func weather() string {
	return "Это температура."
}

func schedule() string {
	return "Это расписание."
}
