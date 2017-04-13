package all_types

import (
	"TelegramBot/vkapi"
	"errors"
	"fmt"
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

// Хранят количество пользователей
var ChatsCount int
var UsersCount int

const (
	NsuHelp   = "nsuhelp"
	NsuLove   = "lovensu"
	NsuSecret = "secretnsu"

	NsuFit = "nsufit"

	DelayUpdate   = time.Minute * 7
	ParseDelay    = time.Second * 31
	ScheduleDelay = time.Minute * 5
)

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
	MaxCountPosts  = 5
	Yes            = 1
	No             = 0
)

// Личные данные
const (
	BotToken = "371494091:AAGndTNOEJpsCO9_CxDuPpa9R025Lxms6UI"
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
	IsActive          bool        `json:"is_active"`
	UserSubscriptions map[int]int `json:"subscriptions"`
	Posts             *[]Post     `json:"posts"`
}

// Post - Структура, заполненная данными с поста
type Post struct {
	Text     string `json:"text"`
	Href     string `json:"href"`
	Date     int    `json:"date"`
	IsPinned int    `json:"is_pinned"`
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

func (s *Subscription) GetNewPosts() (posts []Post, err error) {
	posts, err = GetPosts(s.ScreenName, MaxCountPosts)
	return
}

func (s *Subscription) GetAndRefreshLastPosts() (message []string, err error) {
	newPosts, err := s.GetNewPosts()
	if err != nil {
		return
	}

	if s.Posts == nil {
		for i := 0; i < len(newPosts); i++ {
			message = append(message, newPosts[i].String())
		}

		s.Posts = &newPosts
		return
	}

	for i := 0; i < len(newPosts); i++ {
		flag := true

		for j := 0; j < len(*s.Posts); j++ {
			if newPosts[i].Href == (*s.Posts)[j].Href {
				flag = false
				break
			}
		}

		if flag {
			message = append(message, newPosts[i].String())
		}
	}

	s.Posts = &newPosts

	return
}

func GetPosts(domain string, count int) (posts []Post, err error) {
	res, err := vkapi.GetWallJson(domain, 0, count, "all")
	if err != nil {
		return
	}

	if res.Items == nil {
		return posts, errors.New("*Item равен nil")
	}

	for _, item := range *res.Items {
		if item.MarkedAsAds != 0 {
			continue
		}

		var post Post
		post.Text = item.Text
		post.Date = item.Date
		post.IsPinned = item.IsPinned
		post.Href = "https://vk.com/wall" + fmt.Sprint(item.OwnerID) + "_" + fmt.Sprint(item.ID)

		posts = append(posts, post)
	}

	return
}

func (post *Post) String() string {
	if post.IsPinned == 1 {
		return "Закреплённая запись\n" + time.Unix(int64(post.Date), 0).Format(MyTimeFormat) + "\n" + post.Href + "\n\n" + post.Text
	} else {
		return time.Unix(int64(post.Date), 0).Format(MyTimeFormat) + "\n" + post.Href + "\n\n" + post.Text
	}
}
