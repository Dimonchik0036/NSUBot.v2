package menu

import (
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/loader"
	"TelegramBot/schedule"
	"TelegramBot/subscriptions"
	"TelegramBot/weather"
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"time"
)

//var queue = make(map[int]queueType)
var queue = make(map[int]string)

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
const tag_schedule_day = "tag_schedule_day"
const tag_day = "tag_day"
const schedule_today = "schedule_today"
const schedule_tomorrow = "schedule_tomorrow"
const different_today = "different_today"
const different_tomorrow = "different_tomorrow"
const today = "today"
const tomorrow = "tomorrow"
const faq = "faq"
const feedback = "feedback"
const tag_keyboard = "keyboard"

var FlagToRunner = true

func MessageProcessing(bot *tgbotapi.BotAPI, update tgbotapi.Update) (err error) {
	if update.CallbackQuery != nil {
		return ProcessingCallback(bot, update)
	}

	if update.Message != nil {
		return ProcessingMessage(bot, update)
	}

	if update.InlineQuery != nil {
		loader.Logger.Print("InlineQuery")
	}

	if update.ChosenInlineResult != nil {
		loader.Logger.Print("ChosenInlineResult")
	}

	if update.ChannelPost != nil {
		loader.Logger.Print("ChannelPost")
	}

	return errors.New("Сообщение не прошло обработку.")
}

func ProcessingCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) (err error) {
	loader.Logger.Print("["+update.CallbackQuery.From.UserName+"]"+update.CallbackQuery.From.FirstName+" "+update.CallbackQuery.From.LastName+" ID: ", update.CallbackQuery.From.ID, " CallbackQuery: ", update.CallbackQuery.Data, " ID: ", update.CallbackQuery.From.ID, " MessageID: ", update.CallbackQuery.Message.MessageID)

	/*data := update.CallbackQuery.Data
	q, ok := queue[update.CallbackQuery.From.ID]

	if ok && data != q.oldMenu && data != tag_main && q.command != "" && q.id == update.CallbackQuery.Message.MessageID {
		data = q.command
	}*/

	//queue[update.CallbackQuery.From.ID] = queueType{false, false, "", "", 0}

	command, argument := customers.DecomposeQuery(update.CallbackQuery.Data)

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
		queue[update.CallbackQuery.From.ID] = feedback

		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Наберите свой отзыв:"))
	case faq:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, FaqText())

		m := RowButtonBack(tag_options, true)
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case subscriptions.NsuHelp:
		text := subscriptions.ChangeSubscriptions(update.CallbackQuery.From.ID, "Помогу в НГУ")
		loader.Logger.Print("["+update.CallbackQuery.From.UserName+"]"+update.CallbackQuery.From.FirstName+" "+update.CallbackQuery.From.LastName+" ID: ", update.CallbackQuery.From.ID, " Разультат: "+text)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)

		m := UniteMarkup(SubscriptionsMenu(), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case subscriptions.NsuFit:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Ещё в разработке")

		m := UniteMarkup(SubscriptionsMenu(), RowButtonBack(tag_main, false))
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
			g = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить метку", tag_usergroup)))
		}

		markup := UniteMarkup(g, tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ввести другой номер", tag_schedule_day+" "+argument))),
			RowButtonBack(tag_schedule+" "+argument, true))

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Выберите группу")
		msg.ReplyMarkup = &markup

		bot.Send(msg)
		return
	case tag_day:
		day, group := customers.DecomposeQuery(argument)
		var offset int

		switch day {
		case today:
			offset = 0
		case tomorrow:
			offset = 1
		}

		text, _ := schedule.PrintSchedule(group, offset, update.CallbackQuery.From.ID, false)

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		m := RowButtonBack(tag_schedule_day+" "+day, true)

		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_week:
		for i := 0; i < 7; i++ {

			bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, ""))
		}

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста")

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
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Доступные подписки")

		m := UniteMarkup(SubscriptionsMenu(), RowButtonBack(tag_main, false))
		msg.ReplyMarkup = &m

		bot.Send(msg)
	case tag_usergroup:
		bot.Send(NewUserGroup(update))
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

func StartDeleteLabel(argument string, id int) (text string, markup tgbotapi.InlineKeyboardMarkup) {
	text = "Нажмите на метки, которые хотите удалить"

	if argument != "" {
		v := customers.AllLabels[id]

		if argument == v.MyGroup {
			v.MyGroup = ""

			customers.AllLabels[id] = v
		} else {
			delete(customers.AllLabels[id].Group, argument)
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

func NewUserGroup(update tgbotapi.Update) (answer tgbotapi.Chattable) {
	answer = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Если вы хотите добавить свою группу в избранное, то введите её номер.\n\nЕсли вы хотите добавить/изменить метку, то введите номер группы и название метки через пробел:")
	return
}

func ProcessingMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) (err error) {
	loader.Logger.Print("["+update.Message.From.UserName+"]"+update.Message.From.FirstName+" "+update.Message.From.LastName+" ID: ", update.Message.From.ID, " MessageText: ", update.Message.Text, " ID: ", update.Message.From.ID)

	/*command := update.Message.Command()
	arguments := update.Message.CommandArguments()

	q, ok := queue[update.Message.From.ID]
	queue[update.Message.From.ID] = queueType{false, q.showButton, "", "", 0}

	if !update.Message.IsCommand() && ok {
		command = q.command
		arguments = update.Message.Text
	}*/

	var command string
	var argument string

	if update.Message.IsCommand() {
		command = update.Message.Command()
		argument = update.Message.CommandArguments()
	} else {
		command = queue[update.Message.From.ID]
		argument = update.Message.Text
	}

	queue[update.Message.From.ID] = ""

	switch command {
	case feedback:
		if argument != "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо за отзыв!")

			msg.ReplyMarkup = RowButtonBack(tag_options, true)
			bot.Send(msg)

			bot.Send(tgbotapi.NewMessage(loader.MyId, argument+"\n\nОтзыв от: ["+fmt.Sprint(update.Message.From.ID)+"]\n@"+update.Message.From.UserName+"\n"+update.Message.From.LastName+" "+update.Message.From.FirstName))

			return
		}

		queue[update.Message.From.ID] = feedback
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Наберите свой отзыв:"))

		return
	case "creator", "maker", "author", "father", "Creator", "Maker", "Author", "Father":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Мой телеграм: @Dimonchik0036\nМой GitHub: github.com/dimonchik0036"))
	case "reset":
		if update.Message.From.ID == loader.MyId {
			bot.Send(tgbotapi.NewMessage(loader.MyId, "Выключаюсь."))

			go func() {
				FlagToRunner = false
				time.Sleep(5 * time.Second)

				customers.UpdateUserLabels()
				loader.UpdateUserSubscriptions()

				os.Exit(0)
			}()
		}
	case "weather":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, weather.CurrentWeather))
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Привет!\nЯ - твой помощник, сейчас я покажу тебе, что могу. Советую сразу включить /keyboard, чтобы было проще возвращаться к меню.")

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
	case today, "t", "td", tomorrow, "tm", "tom":
		var day int
		switch command {
		case today, "t", "td":
			day = 0
		case tomorrow, "tm", "tom":
			day = 1
		}

		text, _ := schedule.PrintSchedule(argument, day, update.Message.From.ID, false)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

		/*if ok {
			queue[update.Message.From.ID] = queueType{false, false, "", "", 0}

			/*if q.showButton {
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
		}*/

		bot.Send(msg)
	case "setgroup":
		ok, text := customers.AddGroupNumber(schedule.TableSchedule, update.Message.From.ID, argument)
		if ok {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			loader.Logger.Print("["+update.Message.From.UserName+"]"+update.Message.From.FirstName+" "+update.Message.From.LastName+" ID: ", update.Message.From.ID, " Результат: "+text)

			/*if q.showButton {
				msg.ReplyMarkup = UniteMarkup(LabelsMenu(), RowButtonBack(tag_schedule, true))
			}

			queue[update.Message.From.ID] = queueType{false, false, "", "", 0}*/

			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

			if text != "Предел" {
				//queue[update.Message.From.ID] = queueType{true, q.showButton, command, "", 0}
			} else {
				//queue[update.Message.From.ID] = queueType{false, false, "", "", 0}

				msg.Text = "Вы достигли предела меток. Теперь вы можете только очистить список меток, воспользовавшись командой /clearlabels, " +
					"или изменять группы, привязанные к меткам, но не можете добавлять новые."

				/*if q.showButton {
					msg.ReplyMarkup = UniteMarkup(LabelsMenu(), RowButtonBack(tag_schedule, true))
					msg.Text = "Вы достигли предела меток. Теперь вы можете только очистить список меток " +
						"или изменить группы, привязанные к меткам, но не можете добавлять новые."

				}*/
			}

			/*if !q.run {
				text = "Если вы хотите добавить свою группу в избранное, то введите её номер.\n\nЕсли вы хотите добавить/изменить метку, то введите номер группы и название метки через пробел:"
			}*/

			bot.Send(msg)
		}
	case "labels":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, customers.PrintUserLabels(update.Message.From.ID)))
	case "clearlabels":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, customers.DeleteUserLabels(update.Message.From.ID)))
	case "delete":
		delete(customers.AllLabels[update.Message.From.ID].Group, argument)
	case "joke", "j":
		joke, err := jokes.GetJokes()
		if err == nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, joke))
		}
	case "subjoke":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, jokes.ChangeJokeSubscriptions(update.Message.From.ID)))
	case "nsuhelp":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, subscriptions.ChangeSubscriptions(update.Message.From.ID, "Помогу в НГУ")))
	case "faq":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, FaqText()))
	}

	return
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
			tgbotapi.NewInlineKeyboardButtonData("Добавить/изменить метку", tag_usergroup)),
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
	v, ok := customers.AllLabels[id]
	if !ok {
		return
	}

	if v.Group != nil {
		for l := range v.Group {
			markup.InlineKeyboard = append(markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(l, prefix+l)))
		}
	}

	if v.MyGroup != "" {
		markup.InlineKeyboard = append(markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моё", prefix+v.MyGroup)))
	}

	return
}

func MainKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup, err error) {
	keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/Меню")))
	return
}

func SubscriptionsMenu() (markup tgbotapi.InlineKeyboardMarkup) {
	markup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помогу в НГУ", subscriptions.NsuHelp)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сайт ФИТ НГУ", subscriptions.NsuFit)))
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
			"/subjoke - Подписывает на рассылку бородатых анекдотов. Именно их можно получить, используя /joke\n" +
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

		"Если у Вас остались ещё какие-то вопросы, то их можно задать мне @dimonchik0036."
}
