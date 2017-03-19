package loader

import (
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

var UserFileName string = "users_info.txt"
var TimeFormat string = "02.01.06 15:04:10"

// LoadLoggers Инициализирует логгеры.
func LoadLoggers(logAll **log.Logger) (filenameLogAll string, err error) {
	filenameLogAll = time.Now().Format("020106_1504") + ".txt"

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
func LoadUserGroup(userGroup map[int]string) (string, error) {
	userGroup[227605930] = "16211.1" // Создатель

	userGroup[221524772] = "16361.1" //Паша Тырышкин
	userGroup[215065513] = "16207.1" //Рома Терехов
	userGroup[61219035] = "16209.1"  //Женя Макрушин
	userGroup[250493282] = "16211.1" //Юля Красник
	userGroup[238697588] = "16941.2" //George K
	userGroup[172833377] = "15808.1" //Piligrim_hola
	userGroup[149906245] = "15808.1" //Maria Petlina
	userGroup[185802556] = "15809.1" //Яша Филологический
	userGroup[200867264] = "14304.1" //Saint Pilgrimage
	userGroup[258540109] = "13504.1" //Alexey Taratenko
	userGroup[1469626] = "14308.1"   //Iwan 茴_茴
	userGroup[254438520] = "16134.1" //Vladislav Rublev
	userGroup[161872635] = "16209.1" //Кирилл Полушин
	userGroup[204767177] = "13121.1" //Алексей Р.
	userGroup[693712] = "14203.1"    //Николай Березовский
	userGroup[338030847] = "16203.1" //Fedor Pushkov

	return "", nil
}

// LoadUsers Загружает данные о пользователях.
func LoadUsers(users map[int]string) (int, error) {
	userfile, err := os.OpenFile(UserFileName, os.O_RDWR, os.ModePerm)
	if err != nil {
		return 0, err
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
func LoadSchedule(scheduleMap map[string][7]string) error {
	return nil
}
