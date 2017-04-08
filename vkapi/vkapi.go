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

type Items struct {
	ID          int           `json:"id"`            // ID записи
	FromID      int           `json:"from_id"`       // ID автора
	OwnerID     int           `json:"owner_id"`      // ID владельца стены
	Date        int           `json:"date"`          // Дата размещения записи в unixtime
	MarkedAsAds int           `json:"marked_as_ads"` // Содержит ли рекламу
	PostType    string        `json:"post_type"`     // Тип записи (post, copy, reply, postpone, suggest)
	Text        string        `json:"text"`          // Текст поста
	Attachments []Attachments `json:"attachments"`
	IsPinned int `json:"is_pinned"`
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

type ApiVK struct {
	Responses *Response `json:"response"`
}

type Response struct {
	Count int     `json:"count"` // Количество постов
	Items []Items `json:"items"` // Записи со стены
}

func GetWallJson(domain string, offset int, count int, filter string) (response *Response, err error) {
	switch filter {
	case "all", "owner", "others":
		break
	default:
		return nil, errors.New("Неверное значение фильтра")
	}

	res, err := http.Get(API_METHOD_URL + "wall.get?domain=" + domain + "&offset=" + fmt.Sprintf("%d", offset) + "&count=" + fmt.Sprintf("%d", count) + "&filter=" + filter + "&v=" + VERSION)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	var ApiVK ApiVK
	err = json.Unmarshal(b, &ApiVK)
	if err != nil {
		log.Fatal()
	}

	response = ApiVK.Responses

	return
}

func (item *Items) GetAllPhoto() (photos []string) {
	for _, v := range item.Attachments {
		if v.Type == PHOTO {
			photos = append(photos, v.Photo.GetMaxPhotoHref())
		}
	}

	return
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
