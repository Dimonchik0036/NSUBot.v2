package vkapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	API_METHOD_URL = "https://api.vk.com/method/"
	VERSION        = "5.63"
	PHOTO          = "photo"
	POST           = "post"
)

type Group struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	IsClosed   int    `json:"is_closed"`
	Type       string `json:"type"`
}

type Items struct {
	ID          int           `json:"id"`            // ID записи
	FromID      int           `json:"from_id"`       // ID автора
	OwnerID     int           `json:"owner_id"`      // ID владельца стены
	SignedId    int           `json:"signed_id"`     //идентификатор автора, если запись была опубликована от имени сообщества и подписана пользователем;
	Date        int           `json:"date"`          // Дата размещения записи в unixtime
	MarkedAsAds int           `json:"marked_as_ads"` // Содержит ли рекламу
	PostType    string        `json:"post_type"`     // Тип записи (post, copy, reply, postpone, suggest)
	Text        string        `json:"text"`          // Текст поста
	Attachments []Attachments `json:"attachments"`
	IsPinned    int           `json:"is_pinned"` //информация о том, что запись закреплена.
}

type Attachments struct {
	Type  string `json:"type"`
	Photo *Photo `json:"photo"`
}

type Photo struct {
	ID        int    `json:"id"`       // Идентификатор фотографии
	PostID    int    `json:"post_id"`  // ID поста, к которому прикреплена фотография
	AlbumID   int    `json:"album_id"` //идентификатор альбома, в котором находится фотография
	OwnerID   int    `json:"owner_id"` //идентификатор владельца фотографии.
	UserID    int    `json:"user_id"`  //идентификатор пользователя, загрузившего фото (если фотография размещена в сообществе). Для фотографий, размещенных от имени сообщества, user_id = 100.
	Text      string `json:"text"`     //текст описания фотографии.
	Date      int    `json:"date"`     //дата добавления в формате Unixtime.
	Photo75   string `json:"photo_75"`
	Photo130  string `json:"photo_130"`
	Photo604  string `json:"photo_604"`
	Photo807  string `json:"photo_807"`
	Photo1280 string `json:"photo_1280"`
	Photo2560 string `json:"photo_2560"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	AccessKey string `json:"access_key"` // ключ доступа фотографии
}

type Error struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type ApiVK struct {
	Responses *Response `json:"response"`
	Error     *Error    `json:"error"`
}

type Response struct {
	Count int      `json:"count"` // Количество постов
	Items *[]Items `json:"items"` // Записи со стены
}

func GetWallJson(domain string, offset int, count int, filter string) (*Response, error) {
	switch filter {
	case "all", "owner", "others":
		break
	default:
		return nil, errors.New("Неверное значение фильтра")
	}

	res, err := http.Get(API_METHOD_URL + "wall.get?extended=1&domain=" + domain + "&offset=" + fmt.Sprintf("%d", offset) + "&count=" + fmt.Sprintf("%d", count) + "&filter=" + filter + "&v=" + VERSION)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	res.Body.Close()

	//fmt.Println(string(b))

	var ApiVK ApiVK
	err = json.Unmarshal(b, &ApiVK)
	if err != nil {
		log.Fatal()
	}

	if ApiVK.Error != nil {
		return nil, errors.New(ApiVK.Error.ErrorMsg)
	}

	if ApiVK.Responses != nil {
		return ApiVK.Responses, nil
	}

	return nil, errors.New("Not found")
}

func (item *Items) GetAllPhoto() (photos []string) {
	for _, v := range item.Attachments {
		if v.Type == PHOTO {
			photos = append(photos, v.Photo.GetMaxPhotoHref())
		}
	}

	return
}

func GetGroup(groupId int, groupIds string) (group Group, err error) {
	res, err := http.Get(API_METHOD_URL + "groups.getById?group_id=" + fmt.Sprint(groupId) + "&group_ids=" + groupIds + "&v=" + VERSION)
	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	type R struct {
		Response []Group `json:"response"`
		Error    *Error  `json:"error"`
	}

	var r R

	res.Body.Close()

	err = json.Unmarshal(b, &r)
	if err != nil {
		return
	}

	if r.Error != nil {
		return group, errors.New(r.Error.ErrorMsg)
	}

	return r.Response[0], err
}

func (item *Items) GetOwnerInfo() (group *Group, err error) {
	ownerId := item.OwnerID
	if ownerId < 0 {
		ownerId = -ownerId
	}

	g, err := GetGroup(ownerId, "")
	if err != nil {
		return nil, err
	}

	return &g, err
}

func (photo *Photo) GetMaxPhotoHref() string {
	if photo.Photo2560 != "" {
		return photo.Photo2560
	}

	if photo.Photo1280 != "" {
		return photo.Photo1280
	}

	if photo.Photo807 != "" {
		return photo.Photo807
	}

	if photo.Photo604 != "" {
		return photo.Photo604
	}

	if photo.Photo130 != "" {
		return photo.Photo130
	}

	if photo.Photo75 != "" {
		return photo.Photo75
	}

	return ""
}
