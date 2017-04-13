package loader

import (
	"TelegramBot/all_types"
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

func WriteUsers(mess string) string {
	var u all_types.UserInfo

	err := json.Unmarshal([]byte(mess), &u)
	if err != nil {
		return "Ошибочка."
	}

	return convertUserInfo(u)
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

// LoadUsers Загружает данные о пользователях.
func LoadUsers() (int, error) {
	userFile, err := os.OpenFile(all_types.UsersFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return 0, nil
	}

	var countUsers int

	dec := json.NewDecoder(userFile)

	for {
		var u all_types.UserInfo

		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return countUsers, err
		}

		info, err := json.Marshal(u)
		if err != nil {
			continue
		}

		all_types.AllUsersInfo[u.ID] = string(info)
		countUsers++
	}

	err = userFile.Close()
	if err != nil {
		return countUsers, err
	}

	return countUsers, nil
}

func UpdateUserInfo() error {
	userFile, err := os.OpenFile(all_types.UsersFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	for _, v := range all_types.AllUsersInfo {
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

	_, ok := all_types.AllUsersInfo[update.Message.From.ID]
	if ok {
		return "", false, nil
	}

	timeNow := time.Now().Format(all_types.MyTimeFormat)

	u := all_types.UserInfo{
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

	all_types.AllUsersInfo[u.ID] = string(info)

	userFile, err := os.OpenFile(all_types.UsersFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
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

func convertUserInfo(u all_types.UserInfo) string {
	info := u.UserName

	if info == "" {
		info = u.FirstName + " " + u.LastName
	} else {
		info = "@" + info
	}

	return "ID: " + fmt.Sprintf("%d", u.ID) + "\n" + info + "\nLast action: " + u.TimeLastAction
}

func ReloadUserDate(id int) error {
	info, ok := all_types.AllUsersInfo[id]
	if !ok {
		return errors.New("Не удалось найти пользователя.")
	}

	var u all_types.UserInfo

	err := json.Unmarshal([]byte(info), &u)
	if err != nil {
		return errors.New("Не удалось расшифровать данные.")
	}

	u.TimeLastAction = time.Now().Format(all_types.MyTimeFormat)

	res, err := json.Marshal(u)
	if err != nil {
		return err
	}

	all_types.AllUsersInfo[id] = string(res)

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
