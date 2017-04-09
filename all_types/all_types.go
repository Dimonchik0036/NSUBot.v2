package all_types

import (
	"log"
	"time"
)

// Хранят основную информацию
var AllChatsInfo = make(map[int64]string)
var AllUsersInfo = make(map[int]string)
var AllSchedule = make(map[string][7]string)
var AllLabels = make(map[int]UserGroup)
var AllSubscription = make(map[string]*Subscription)

// Logger - Логер всех событий программы
var Logger *log.Logger
var LoggerFilename string = time.Now().Format("2006-01-02T15-04") + ".txt"

const (
	UsersFilename         = "users_info.txt"
	LabelsFilename        = "labels.txt"
	SubscriptionsFilename = "users_subscriptions.txt"
)

const (
	MyTimeFormat   = "02.01.06 15:04:10"
	MyGroupLabel   = "Моя"
	MaxCountLabel  = 20
	MaxCountSymbol = 64
	Yes            = 1
	No             = 0
)

// Личные данные
const (
	BotToken = "325933326:AAFWjDWFPKFjAMg9MDr_Av-g643F_UhJmNY"
	MyId     = 227605930
)

// UserGroup - Хранит метки пользователя
type UserGroup struct {
	Group   map[string]string
	MyGroup string
}

type UserGroupLabels struct {
	Id     int    `json:"ID"`
	Labels string `json:"Labels"`
}

type UserLabels struct {
	Label string `json:"Label"`
	Group string `json:"Group"`
}

// UserInfo - Хранит данные о пользователе
type UserInfo struct {
	TimeCreate     string `json:"TimeCreate"`
	TimeLastAction string `json:"TimeLastAction"`
	FirstName      string `json:"FirstName"`
	LastName       string `json:"LastName"`
	UserName       string `json:"UserName"`
	ID             int    `json:"ID"`
}

// Subscription - Вся структура хранения данных о подписках
type Subscription struct {
	Name              string      `json:"name"`
	ScreenName        string      `json:"screen_name"`
	Posts             *[]Post     `json:"posts"`
	UserSubscriptions map[int]int `json:"subscriptions"`
}

func (s *Subscription) ChangeSubscriptions(id int) string {
	v, ok := s.UserSubscriptions[id]
	if !ok {
		s.UserSubscriptions[id] = Yes
		return "Вы были подписаны на рассылку " + s.Name + "."
	} else {
		if v != 0 {
			s.UserSubscriptions[id] = No
			return "Вы были отписаны от рассылки " + s.Name + "."
		} else {
			s.UserSubscriptions[id] = Yes
			return "Вы были подписаны на рассылку " + s.Name + "."
		}
	}
}

type Post struct {
	Text     string `json:"text"`
	Href     string `json:"href"`
	Date     int    `json:"date"`
	IsPinned int    `json:"is_pinned"`
}

func (post *Post) String() string {
	return time.Unix(int64(post.Date), 0).Format(MyTimeFormat) + "\n" + post.Href + "\n" + post.Text
}
