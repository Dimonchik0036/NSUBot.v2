package loader

import (
	"TelegramBot/customers"
	"TelegramBot/jokes"
	"TelegramBot/subscriptions"
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

type UserInfo struct {
	TimeCreate     string `json:"TimeCreate"`
	TimeLastAction string `json:"TimeLastAction"`
	FirstName      string `json:"FirstName"`
	LastName       string `json:"LastName"`
	UserName       string `json:"UserName"`
	ID             int    `json:"ID"`
}

type Subscriptions struct {
	Id        int    `json:"ID"`
	Group     string `json:"Group"`
	Selection int    `json:"Selection"`
}

var UserFileName string = "users_info.txt"
var TimeFormat string = "02.01.06 15:04:10"

func LoadUsersSubscriptions() error {
	userFile, err := os.OpenFile(subscriptions.FileUsersSubscriptions, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	decUsers := json.NewDecoder(userFile)

	for {
		var s Subscriptions

		if err := decUsers.Decode(&s); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch s.Group {
		case "nsuhelp":
			subscriptions.UsersNsuHelp[s.Id] = s.Selection
		case "jokes":
			jokes.JokeBase[s.Id] = s.Selection
		}
	}

	err = userFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserSubscriptions() error {
	userFile, err := os.OpenFile(subscriptions.FileUsersSubscriptions, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	var s Subscriptions
	for i, v := range subscriptions.UsersNsuHelp {
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

	for i, v := range jokes.JokeBase {
		s.Id = i
		s.Selection = v
		s.Group = "jokes"

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
func LoadLoggers(logAll **log.Logger) (filenameLogAll string, err error) {
	filenameLogAll = time.Now().Format("2006-01-02T15-04") + ".txt"

	fileLoggerAll, err := os.OpenFile(filenameLogAll, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return "", errors.New("Не удалось открыть файл: " + filenameLogAll)
	}

	*logAll = log.New(fileLoggerAll, "", log.LstdFlags)
	(*logAll).Println("Начинаю.")

	return
}

func WriteUsers(mess string) string {
	var u UserInfo

	err := json.Unmarshal([]byte(mess), &u)
	if err != nil {
		return "Ошибочка."
	}

	return convertUserInfo(u)
}

// LoadUserGroup Загружает данные о запомненных группах.
func LoadUserGroup() error {
	userfile, err := os.OpenFile(customers.LabelsFile, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}

	decUsers := json.NewDecoder(userfile)

	for {
		var u customers.UserGroupLabels

		if err := decUsers.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		decLabels := json.NewDecoder(bytes.NewReader([]byte(u.Labels)))

		var g customers.UserGroup
		g.Group = make(map[string]string)

		for {
			var l customers.UserLabels

			if err := decLabels.Decode(&l); err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if l.Label == customers.MyGroupLabel {
				g.MyGroup = l.Group
			} else {
				g.Group[l.Label] = l.Group
			}
		}

		customers.AllLabels[u.Id] = g
	}

	err = userfile.Close()
	if err != nil {
		return err
	}

	return nil
}

// LoadUsers Загружает данные о пользователях.
func LoadUsers(users map[int]string) (int, error) {
	userfile, err := os.OpenFile(UserFileName, os.O_RDWR, os.ModePerm)
	if err != nil {
		return 0, nil
	}

	var countUsers int

	dec := json.NewDecoder(userfile)

	for {
		var u UserInfo

		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return countUsers, err
		}

		info, err := json.Marshal(u)
		if err == nil {
			users[u.ID] = string(info)
		}

		countUsers++
	}

	err = userfile.Close()
	if err != nil {
		return countUsers, err
	}

	return countUsers, nil
}

func UpdateUserInfo(users map[int]string) error {
	userFile, err := os.OpenFile(UserFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	for _, v := range users {
		userFile.WriteString(v + "\n")
	}

	err = userFile.Close()

	return err
}

// NewUserInfo Возвращает строку с новым пользователем
func NewUserInfo(users map[int]string, update *tgbotapi.Update) (string, bool, error) {
	_, ok := users[update.Message.From.ID]
	if ok {
		return "", false, nil
	}

	timeNow := time.Now().Format(TimeFormat)

	u := UserInfo{
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

	users[u.ID] = string(info)

	userFile, err := os.OpenFile(UserFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
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

func convertUserInfo(u UserInfo) string {
	info := u.UserName

	if info == "" {
		info = u.FirstName + " " + u.LastName
	} else {
		info = "@" + info
	}

	return "ID: " + fmt.Sprintf("%d", u.ID) + "\n" + info + "\nLast action: " + u.TimeLastAction
}

func ReloadUserDate(users map[int]string, id int) error {
	info, ok := users[id]
	if !ok {
		return errors.New("Не удалось найти пользователя.")
	}

	var u UserInfo

	err := json.Unmarshal([]byte(info), &u)
	if err != nil {
		return errors.New("Не удалось расшифровать данные.")
	}

	u.TimeLastAction = time.Now().Format(TimeFormat)

	res, err := json.Marshal(u)
	if err == nil {
		users[id] = string(res)
	}

	return nil
}

// LoadChats Загружает данные о чатах.
func LoadChats(chats map[int64]string) error {
	return nil
}

// LoadSchedule Загружает данные о чатах.
func LoadSchedule() error {
	return nil
}
