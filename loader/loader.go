package loader

import (
	"TelegramBot/all_types"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"log"
	"os"
	"time"
)

func LoadUsersSubscriptions() error {
	userFile, err := os.OpenFile(all_types.SubscriptionsFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	dec := json.NewDecoder(userFile)
	err = dec.Decode(&all_types.AllSubscription)
	if err != nil {
		return err
	}

	err = userFile.Close()

	return err
}

func UpdateUserSubscriptions() error {
	userFile, err := os.OpenFile(all_types.SubscriptionsFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	b, err := json.Marshal(all_types.AllSubscription)
	if err != nil {
		return err
	}

	userFile.Write(b)

	err = userFile.Close()

	return err
}

// LoadLoggers Инициализирует логгеры.
func LoadLoggers() (err error) {
	fileLogger, err := os.OpenFile(all_types.LoggerFilename, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.New("Не удалось открыть файл: " + all_types.LoggerFilename)
	}

	all_types.Logger = log.New(fileLogger, "", log.LstdFlags)
	all_types.Logger.Println("Начинаю.")

	return
}

// LoadUserGroup Загружает данные о запомненных группах.
func LoadUserGroup() error {
	userFile, err := os.OpenFile(all_types.LabelsFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	decUsers := json.NewDecoder(userFile)

	for {
		var u all_types.UserGroupLabels

		if err := decUsers.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		decLabels := json.NewDecoder(bytes.NewReader([]byte(u.Labels)))

		var g all_types.UserGroup
		g.Group = make(map[string]string)

		for {
			var l all_types.UserLabels

			if err := decLabels.Decode(&l); err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if l.Label == all_types.MyGroupLabel {
				g.MyGroup = l.Group
			} else {
				g.Group[l.Label] = l.Group
			}
		}

		all_types.AllLabels[u.Id] = g
	}

	err = userFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func LoadUsersInfo() (err error) {
	userFile, err := os.OpenFile(all_types.UsersFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	dec := json.NewDecoder(userFile)

	err = dec.Decode(&all_types.AllUsersInfo)
	if err != nil {
		return
	}

	err = userFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserInfo() error {
	userFile, err := os.OpenFile(all_types.UsersFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	b, err := json.Marshal(all_types.AllUsersInfo)
	if err != nil {
		return err
	}

	userFile.Write(b)

	err = userFile.Close()

	return err
}

// RefreshUserInfo Возвращает строку с новым пользователем
func RefreshUserInfo(bot *tgbotapi.BotAPI, user tgbotapi.User) {
	_, ok := all_types.AllUsersInfo[user.ID]

	timeNow := time.Now().Format(all_types.MyTimeFormat)

	var u all_types.UserInfo

	if !ok {
		u.TimeCreate = timeNow
	}
	u.TimeLastAction = timeNow
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.ID = user.ID
	u.PermissionToSend = true

	if user.UserName != "" {
		u.UserName = "@" + user.UserName
	}

	all_types.AllUsersInfo[u.ID] = &u

	if !ok {
		bot.Send(tgbotapi.NewMessage(all_types.MyId, "Новый пользователь!\n"+u.String()))
	}
}

func ReloadUserDate(bot *tgbotapi.BotAPI, user tgbotapi.User) error {
	u, ok := all_types.AllUsersInfo[user.ID]
	if !ok {
		RefreshUserInfo(bot, user)
		return errors.New("Не удалось найти пользователя.")
	}

	u.TimeLastAction = time.Now().Format(all_types.MyTimeFormat)
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	if user.UserName != "" {
		u.UserName = "@" + user.UserName
	}

	return nil
}

// LoadChats Загружает данные о чатах.
func LoadChats() error {
	return nil
}

// LoadSchedule Загружает данные о чатах.
func LoadSchedule() error {
	return nil
}
