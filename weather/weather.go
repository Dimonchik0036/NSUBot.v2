package weather

import (
	"bufio"
	"errors"
	"net/http"
	"time"
)

// SearchWeather Возвращает строку с температурой, в противном случае ошибку.
func SearchWeather() (string, error) {
	res, err := http.Get("http://weather.nsu.ru/loadata.php")
	if err != nil {
		return "", err
	}

	if res.Status != "200 OK" {
		return "", errors.New("Ошибка статуса страницы: " + res.Status)
	}

	reader := bufio.NewReader(res.Body)

	text, err := reader.ReadBytes(' ')
	if err != nil {
		return "", err
	}

	for string(text) != "'Температура " {
		text, err = reader.ReadBytes(' ')
		if err != nil {
			return "", err
		}
	}

	t, err := reader.ReadBytes('\'')
	if err != nil {
		return "", errors.New("Ошибка при чтении погоды.")
	}

	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	mess := string(text[1:])
	mess += string(t[:len(t)-1])
	mess += "\nВремя последнего обновления: " + time.Now().Format("02.01.06 15:04")

	return mess, nil
}
