package weather

import (
	"bufio"
	"log"
	"net/http"
	"time"
)

func SearchWeather(weather *string, logger *log.Logger) {
	for {
		res, err := http.Get("http://weather.nsu.ru/loadata.php")
		if err != nil {
			logger.Print("weather:", err)

			time.Sleep(time.Minute)
			continue
		}

		if res.Status != "200 OK" {
			logger.Print("Ошибка статуса страницы:", err)

			time.Sleep(time.Minute)
			continue
		}

		reader := bufio.NewReader(res.Body)

		text, err := reader.ReadBytes(' ')
		if err != nil {
			logger.Print("weather:", err)

			time.Sleep(time.Minute)
			continue
		}

		for string(text) != "'Температура " {
			text, err = reader.ReadBytes(' ')
			if err != nil {
				logger.Print("weather:", err)
			}
		}

		t, err := reader.ReadBytes('\'')
		if err != nil {
			logger.Println("weather:", err)

			time.Sleep(time.Minute)
			continue
		}

		res.Body.Close()

		mess := string(text[1:])
		mess += string(t[:len(t)-1])
		mess += "\nВремя последнего обновления: " + time.Now().Format("02.01.06 15:04")

		*weather = mess

		time.Sleep(time.Minute)
	}
}
