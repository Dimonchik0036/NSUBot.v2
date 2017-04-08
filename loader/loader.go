package loader

import (
	"TelegramBot/types"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"log"
	"os"
	"time"
)

func LoadUsersSubscriptions() error {
	userFile, err := os.OpenFile(types.SubscriptionsFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	decUsers := json.NewDecoder(userFile)

	for {
		var s types.Subscriptions

		if err := decUsers.Decode(&s); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch s.Group {
		case "nsuhelp":
			types.UsersNsuHelp[s.Id] = s.Selection
		}
	}

	err = userFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserSubscriptions() error {
	userFile, err := os.OpenFile(types.SubscriptionsFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	var s types.Subscriptions
	for i, v := range types.UsersNsuHelp {
		s.Id = i
		s.Selection = v
		s.Group = "nsuhelp"

		b, err := json.Marshal(s)
		if err != nil {
			return err
		}

		b = append(b, '\n')

		userFile.Write(b)
	}

	err = userFile.Close()

	return err
}

// LoadLoggers Инициализирует логгеры.
func LoadLoggers() (err error) {
	fileLogger, err := os.OpenFile(types.LoggerFilename, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.New("Не удалось открыть файл: " + types.LoggerFilename)
	}

	types.Logger = log.New(fileLogger, "", log.LstdFlags)
	types.Logger.Println("Начинаю.")

	return
}

func WriteUsers(mess string) string {
	var u types.UserInfo

	err := json.Unmarshal([]byte(mess), &u)
	if err != nil {
		return "Ошибочка."
	}

	return convertUserInfo(u)
}

// LoadUserGroup Загружает данные о запомненных группах.
func LoadUserGroup() error {
	userFile, err := os.OpenFile(types.LabelsFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	decUsers := json.NewDecoder(userFile)

	for {
		var u types.UserGroupLabels

		if err := decUsers.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		decLabels := json.NewDecoder(bytes.NewReader([]byte(u.Labels)))

		var g types.UserGroup
		g.Group = make(map[string]string)

		for {
			var l types.UserLabels

			if err := decLabels.Decode(&l); err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if l.Label == types.MyGroupLabel {
				g.MyGroup = l.Group
			} else {
				g.Group[l.Label] = l.Group
			}
		}

		types.AllLabels[u.Id] = g
	}

	err = userFile.Close()
	if err != nil {
		return err
	}

	return nil
}

// LoadUsers Загружает данные о пользователях.
func LoadUsers() (int, error) {
	userFile, err := os.OpenFile(types.UsersFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return 0, nil
	}

	var countUsers int

	dec := json.NewDecoder(userFile)

	for {
		var u types.UserInfo

		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return countUsers, err
		}

		info, err := json.Marshal(u)
		if err != nil {
			continue
		}

		types.AllUsersInfo[u.ID] = string(info)
		countUsers++
	}

	err = userFile.Close()
	if err != nil {
		return countUsers, err
	}

	return countUsers, nil
}

func UpdateUserInfo() error {
	userFile, err := os.OpenFile(types.UsersFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	for _, v := range types.AllUsersInfo {
		userFile.WriteString(v + "\n")
	}

	err = userFile.Close()

	return err
}

// NewUserInfo Возвращает строку с новым пользователем
func NewUserInfo(update tgbotapi.Update) (string, bool, error) {
	if update.Message == nil {
		return "", false, errors.New("Не сообщение")
	}

	_, ok := types.AllUsersInfo[update.Message.From.ID]
	if ok {
		return "", false, nil
	}

	timeNow := time.Now().Format(types.MyTimeFormat)

	u := types.UserInfo{
		timeNow,
		timeNow,
		update.Message.From.FirstName,
		update.Message.From.LastName,
		update.Message.From.UserName,
		update.Message.From.ID}

	info, err := json.Marshal(u)
	if err != nil {
		return "", true, err
	}

	types.AllUsersInfo[u.ID] = string(info)

	userFile, err := os.OpenFile(types.UsersFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return convertUserInfo(u), true, err
	}

	userFile.WriteString(string(info) + "\n")

	err = userFile.Close()
	if err != nil {
		return convertUserInfo(u), true, err
	}

	return convertUserInfo(u), true, nil
}

func convertUserInfo(u types.UserInfo) string {
	info := u.UserName

	if info == "" {
		info = u.FirstName + " " + u.LastName
	} else {
		info = "@" + info
	}

	return "ID: " + fmt.Sprintf("%d", u.ID) + "\n" + info + "\nLast action: " + u.TimeLastAction
}

func ReloadUserDate(id int) error {
	info, ok := types.AllUsersInfo[id]
	if !ok {
		return errors.New("Не удалось найти пользователя.")
	}

	var u types.UserInfo

	err := json.Unmarshal([]byte(info), &u)
	if err != nil {
		return errors.New("Не удалось расшифровать данные.")
	}

	u.TimeLastAction = time.Now().Format(types.MyTimeFormat)

	res, err := json.Marshal(u)
	if err != nil {
		return err
	}

	types.AllUsersInfo[id] = string(res)

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
