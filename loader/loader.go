package loader

import (
	"errors"
	"log"
	"os"
	"time"
)

// InitLoggers Инициализирует логгеры.
func InitLoggers(logUser **log.Logger, logAll **log.Logger) (filenameLogUsers string, filenameLogAll string, err error) {
	filenameLogUsers = "logUsers.txt"
	filenameLogAll = time.Now().Format("020106_1504") + ".txt"

	fileLoggerAll, err := os.OpenFile(filenameLogAll, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return "", "", errors.New("Не удалось открыть файл: " + filenameLogAll)
	}

	*logAll = log.New(fileLoggerAll, "", log.LstdFlags)
	(*logAll).Println("Начинаю.")

	fileLogUsers, err := os.OpenFile(filenameLogUsers, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return "", "", errors.New("Не удалось открыть файл: " + filenameLogUsers)
	}

	_, err = fileLogUsers.Seek(0, os.SEEK_END)
	if err != nil {
		return "", "", errors.New("Не удалось перейти в конец файла.")
	}

	*logUser = log.New(fileLogUsers, "", log.LstdFlags)
	(*logUser).Println("\n<<<<<<Начало новой сессии>>>>>\n")

	return
}

// LoadUserGroup Загружает данные о запомненных группах.
func LoadUserGroup(userGroup map[int]string) error {
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

	return nil
}

// LoadUsers Загружает данные о пользователях.
func LoadUsers(users map[int]string) error {
	users[227605930] = "Создатель: @Dimonchik0036" // Создатель

	users[61219035] = "Ник: @banyrule\nИмя: Bany\nФамилия: Rule\nID: 61219035"
	users[57813058] = "Ник: @bsgun\nИмя: mitya\nФамилия: mihelson\nID: 57813058"
	users[244489778] = "Ник: @\nИмя: Elfrida\nФамилия: Bambutsa\nID: 244489778"
	users[129363483] = "Ник: @\nИмя: Андрей\nФамилия: Щербин\nID: 129363483"
	users[238697588] = "Ник: @\nИмя: George\nФамилия: K\nID: 238697588"
	users[248239658] = "Ник: @\nИмя: Tiko\nФамилия: Defect\nID: 248239658"
	users[243429867] = "Ник: @\nИмя: Александр\nФамилия: Афанасенков\nID: 243429867"
	users[221524772] = "Ник: @\nИмя: Павел\nФамилия: Тырышкин\nID: 221524772"
	users[172833377] = "Ник: @\nИмя: Piligrim_hola\nФамилия: \nID: 172833377"
	users[185802556] = "Ник: @Piter_Piter\nИмя: Яша\nФамилия: Филологический\nID: 185802556"
	users[107950408] = "Ник: @kvblinov\nИмя: Konstantin\nФамилия: Blinov\nID: 107950408"
	users[149906245] = "Ник: @mariapetlina\nИмя: Maria\nФамилия: Petlina\nID: 149906245"
	users[200867264] = "Ник: @dragn126\nИмя: Saint\nФамилия: Pilgrimage\nID: 200867264"
	users[94943173] = "Ник: @VLS_TLGRM\nИмя: VeLLeSSS/Сергей\nФамилия: Кулеша\nID: 94943173"
	users[258540109] = "Ник: @LeXT5\nИмя: Alexey\nФамилия: Taratenko\nID: 258540109"
	users[218567363] = "Ник: @\nИмя: Vitaly\nФамилия: Liber\nID: 218567363"
	users[1469626] = "Ник: @iwanko\nИмя: Iwan\nФамилия: 茴_茴\nID: 1469626"
	users[254438520] = "Ник: @\nИмя: Vladislav\nФамилия: Rublev\nID: 254438520"
	users[204767177] = "Ник: @\nИмя: Алексей\nФамилия: Р.\nID: 204767177"
	users[270519216] = "Ник: @therrer\nИмя: Человек-полторашка\nФамилия: \nID: 270519216"
	users[215065513] = "Ник: @\nИмя: Роман\nФамилия: Терехов\nID: 215065513"
	users[161872635] = "Ник: @kirpichik\nИмя: Кирилл\nФамилия: Полушин\nID: 161872635"
	users[200874470] = "Ник: @MrAkakuy\nИмя: Paul\nФамилия: Kholyavko\nID: 200874470"
	users[693712] = "Ник: @nberezowsky\nИмя: Николай\nФамилия: Березовский\nID: 693712"
	users[70167980] = "Ник: @mefbus\nИмя: Eba⚡️⚡️osina\nФамилия: \nID: 70167980"
	users[338030847] = "Ник: @\nИмя: Fedor\nФамилия: Pushkov\nID: 338030847"
	users[142080444] = "Ник: @dem1tris\nИмя: Dmitry\nФамилия: Ivanishkin\nID: 142080444"
	users[245647624] = "Ник: @\nИмя: Ksu\nФамилия: Pecherskikh\nID: 245647624"
	users[250493282] = "Ник: @\nИмя: Yulia\nФамилия: Krasnik\nID: 250493282"
	users[245090894] = "Ник: @\nИмя: Polina\nФамилия: L.\nID: 245090894"

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
