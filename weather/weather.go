package weather

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

var CurrentWeather string = "Погода временно недоступна, попробуйте чуть позднее."

// SearchWeather Возвращает строку с температурой, в противном случае ошибку.
func SearchWeather() error {
	res, err := http.Get("http://weather.nsu.ru/loadata.php")
	if err != nil {
		return err
	}

	if res.Status != "200 OK" {
		return errors.New("Ошибка статуса страницы: " + res.Status)
	}

	textBody, err := ioutil.ReadAll(bufio.NewReader(res.Body))
	if err != nil {
		return err
	}

	reg, err := regexp.Compile("'Температура около .*?'")
	if err != nil {
		return err
	}

	err = res.Body.Close()
	if err != nil {
		return err
	}

	bytes := reg.Find(textBody)
	if len(bytes) == 0 {
		return errors.New("Не удалось вытащить температуру.")
	}

	mess := string(bytes[1 : len(bytes)-1])
	mess += "\nВремя последнего обновления: " + time.Now().Format("02.01.06 15:04")

	CurrentWeather = mess

	return nil
}
