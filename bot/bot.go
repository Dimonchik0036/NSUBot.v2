package main

import (
	"TelegramBot/all_types"
	"TelegramBot/customers"
	"TelegramBot/loader"
	"TelegramBot/menu"
	"TelegramBot/schedule"
	"TelegramBot/subscriptions"
	"TelegramBot/weather"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"time"
)

// loadAll - Загружает все необходимые данные и возвращает указатель на BotAPI
func loadAll() (bot *tgbotapi.BotAPI) {
	bot, err := tgbotapi.NewBotAPI(all_types.BotToken)
	if err != nil {
		all_types.Logger.Fatal("Бот в отпуске: ", err)
	}

	info, err := schedule.GetAllSchedule("GK")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(all_types.MyId, "Всё плохо с GK"))
		all_types.Logger.Fatal("GK")
	} else {
		all_types.Logger.Print(info)
	}

	info, err = schedule.GetAllSchedule("LK")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(all_types.MyId, "Всё плохо с LK"))
		all_types.Logger.Fatal("LK")
	} else {
		all_types.Logger.Print(info)
	}

	go func() {
		for {
			answer, err := schedule.ParseSchedule("GK")
			if err != nil {
				all_types.Logger.Print(err)
			} else {
				if answer != "" {
					all_types.Logger.Print(answer)
				}
			}

			answer, err = schedule.ParseSchedule("LK")
			if err != nil {
				all_types.Logger.Print(err)
			} else {
				if answer != "" {
					all_types.Logger.Print(answer)
				}
			}

			time.Sleep(all_types.ScheduleDelay)
		}
	}()

	go func() {
		for {
			err := weather.SearchWeather()
			if err != nil {
				all_types.Logger.Print(err)
			}

			time.Sleep(2 * time.Minute)
		}
	}()

	err = loader.LoadUsersInfo()
	if err != nil {
		all_types.Logger.Print(err)
	}

	err = loader.LoadChats()
	if err != nil {
		all_types.Logger.Print(err)
	}

	err = loader.LoadUserGroup()
	if err != nil {
		all_types.Logger.Print(err)
	}

	err = loader.LoadSchedule()
	if err != nil {
		all_types.Logger.Print(err)
	}

	err = loader.LoadUsersSubscriptions()
	if err != nil {
		all_types.Logger.Print(err)
	}

	err = subscriptions.LoadFitNsuFile()
	if err != nil {
		all_types.Logger.Print(err)
	}

	_, err = bot.Send(tgbotapi.NewMessage(all_types.MyId, "Я перезагрузился."))
	if err != nil {
		all_types.Logger.Print("Не смог отправить весточку повелителю.", err)
	}

	CheckDefaultGroup(bot, all_types.NsuHelp)
	CheckDefaultGroup(bot, all_types.NsuSecret)
	CheckDefaultGroup(bot, all_types.NsuLove)
	CheckDefaultGroup(bot, all_types.Nsu24)
	CheckDefaultGroup(bot, all_types.NsuTypical)

	CheckDefaultFit(bot, all_types.News_announc, "Объявления")
	CheckDefaultFit(bot, all_types.News_konf, "Конференции")
	CheckDefaultFit(bot, all_types.News_news, "События")
	CheckDefaultFit(bot, all_types.News_conc, "Конкурсы")

	CheckDefaultFit(bot, all_types.News_admin_prikazy, "Административные приказы")

	CheckDefaultFit(bot, all_types.News_chairs+all_types.News_anksi, "Кафедра систем информатики")
	CheckDefaultFit(bot, all_types.News_chairs+all_types.News_ankks, "Кафедра компьютерных систем")
	CheckDefaultFit(bot, all_types.News_chairs+all_types.News_koinews, "Кафедра общей информатики")
	CheckDefaultFit(bot, all_types.News_chairs+all_types.News_kpvnews, "Кафедра параллельных вычислений")
	CheckDefaultFit(bot, all_types.News_chairs+all_types.News_kktnews, "Кафедра компьютерных технологий")

	go func() {
		for {
			time.Sleep(all_types.DelayUpdate)

			if !menu.FlagToRunner {
				return
			}

			err := loader.UpdateUserInfo()
			if err != nil {
				all_types.Logger.Print(err)
			}

			err = customers.UpdateUserLabels()
			if err != nil {
				all_types.Logger.Print(err)
			}

			err = loader.UpdateUserSubscriptions()
			if err != nil {
				all_types.Logger.Print(err)
			}

			err = subscriptions.RefreshFitNsuFile()
			if err != nil {
				all_types.Logger.Print(err)
			}
		}
	}()

	go func() {
		for {
			for _, v := range all_types.AllSubscription {
				if !v.IsActive {
					continue
				}

				m, err := v.GetAndRefreshLastPosts()
				if err != nil {
					all_types.Logger.Print(err)
					continue
				}

				if len(m) > 0 {
					if v.IsReady {
						for i, ok := range v.UserSubscriptions {
							if ok != 0 {
								for _, post := range m {
									if len(post) > 4500 {
										post = post[:4500] + "...\n\nСлишком длинное сообщение, продолжение доступно по ссылке в начале сообщения."
									}

									bot.Send(tgbotapi.NewMessage(int64(i), v.Name+"\n"+post))
								}
							}
						}
					} else {
						for _, post := range m {
							if len(post) > 4500 {
								post = post[:4500] + "...\n\nСлишком длинное сообщение, продолжение доступно по ссылке в начале сообщения."
							}

							bot.Send(tgbotapi.NewMessage(all_types.MyId, v.Name+"\n"+post))
						}

						bot.Send(tgbotapi.NewMessage(all_types.MyId, "Для меня любимого"))
					}
				}

				time.Sleep(time.Second)
			}

			for _, l := range subscriptions.FitNsuNews {
				if !l.IsActive {
					continue
				}

				m, err := l.GetAndRefreshLastNews()
				if err != nil {
					all_types.Logger.Print(err)
					continue
				}

				if len(m) > 0 {
					for i, ok := range l.Users {
						if ok != 0 {
							for _, post := range m {
								if len(post) > 4500 {
									post = post[:4500] + "...\n\nСлишком длинное сообщение, продолжение доступно по ссылке в начале сообщения."
								}

								bot.Send(tgbotapi.NewMessage(int64(i), l.MainTitle+"\n"+post))
							}
						}
					}
				}
			}

			time.Sleep(all_types.ParseDelay)
		}
	}()

	all_types.Logger.Printf("Бот %s запущен.", bot.Self.UserName)

	return
}

func CheckDefaultGroup(bot *tgbotapi.BotAPI, domain string) {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		err := subscriptions.AddNewGroupToParse(domain)
		if err != nil {
			all_types.Logger.Print(err)
		}

		bot.Send(tgbotapi.NewMessage(all_types.MyId, "Отсутствует: "+domain))
		all_types.Logger.Print("Отсутствует: " + domain)
	} else {
		if !s.IsActive {
			bot.Send(tgbotapi.NewMessage(all_types.MyId, "Бездействует: "+domain))
			all_types.Logger.Print("Бездействует: " + domain)
		}
	}
}

func CheckDefaultFit(bot *tgbotapi.BotAPI, href string, title string) {
	s, ok := subscriptions.FitNsuNews[href]
	if !ok {
		answer := subscriptions.AddNewNewsList(href, title)

		bot.Send(tgbotapi.NewMessage(all_types.MyId, answer))
		all_types.Logger.Print(answer)
	} else {
		if !s.IsActive {
			bot.Send(tgbotapi.NewMessage(all_types.MyId, "Бездействует: "+href))
			all_types.Logger.Print("Бездействует: " + href)
		}
	}
}

func messageLog(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if (update.Message.Chat.IsGroup() || update.Message.Chat.IsChannel() || update.Message.Chat.IsSuperGroup()) && update.Message.IsCommand() {
		all_types.Logger.Printf("[%d] %s",
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
		_, ok := all_types.AllChatsInfo[update.Message.Chat.ID]
		if !ok {
			n := newChat(update.Message.Chat)
			all_types.AllChatsInfo[update.Message.Chat.ID] = n

			_, err := bot.Send(tgbotapi.NewMessage(all_types.MyId, "Новая чат-сессия!\n"+n))
			if err != nil {
				all_types.Logger.Print("newChat:", err)
			}
		}
	}

	loader.NewUserInfo(bot, *update.Message.From)

	return nil
}

func messages(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	processingUser(bot, update)
	messageLog(update)

	err := menu.MessageProcessing(bot, update)
	if err != nil {
		all_types.Logger.Print(err)
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

	return message
}

func main() {
	err := loader.LoadLoggers()
	if err != nil {
		log.Fatal(err)
	}

	bot := loadAll()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		all_types.Logger.Fatal(err)
	}

	for update := range updates {
		go messages(bot, update)
	}
}
