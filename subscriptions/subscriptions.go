package subscriptions

import (
	"TelegramBot/all_types"
	"TelegramBot/mymodule"
	"TelegramBot/vkapi"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var FitNsuNews = make(map[string]*NewsList)

type NewsList struct {
	MainTitle string      `json:"main_title"`
	Href      string      `json:"href"`
	IsActive  bool        `json:"is_active"`
	Pages     *[]NewsPage `json:"pages"`
	Users     map[int]int `json:"users"`
}

type NewsPage struct {
	Title string `json:"title"`
	Date  string `json:"modifer"`
	Href  string `json:"href"`
	Page  *Page  `json:"page"`
}

func (p *NewsPage) String() string {
	return "http://fit.nsu.ru" + p.Href + "\n" + "Дата: " + p.Date
}

type Page struct {
	Title string `json:"title"`
	Date  string `json:"modifer"`
	Text  string `json:"text"`
}

func ChangeFitNewsActivity(href string) string {
	l, ok := FitNsuNews[href]
	if !ok {
		return "Такого раздела не существует"
	}

	if l.IsActive {
		l.IsActive = false
		return "Деактивирован раздел [" + l.MainTitle + "]"
	} else {
		l.IsActive = true
		return "Активирован раздел [" + l.MainTitle + "]"
	}
}

func ShowAllFitNewsGroup() (groups []string) {
	for i, v := range FitNsuNews {
		groups = append(groups, "["+i+"] "+v.MainTitle+" ["+fmt.Sprint(v.IsActive)+"]")
	}
	if len(groups) == 0 {
		groups = append(groups, "Список групп пуст")
	}

	return
}

func ShowAllFitUsersGroup(href string) (message []string) {
	l, ok := FitNsuNews[href]
	if !ok {
		message = append(message, "Группа не найдена")
		return
	}

	for i, u := range l.Users {
		message = append(message, "ID: "+fmt.Sprint(i)+", состояние: "+fmt.Sprint(u))
	}

	if len(message) == 0 {
		message = append(message, "Подписки отсутсвуют")
		return
	}

	return
}

func ChangeUserFit(href string, id int) (answer string) {
	l, ok := FitNsuNews[href]
	if !ok {
		return "Ошибка обработки группы, сообщите об этом /feedback"
	}

	u, ok := l.Users[id]
	if !ok {
		l.Users[id] = all_types.Yes
		return "Вы были подписаны на рассылку новостей из раздела " + l.MainTitle
	}

	if u != 0 {
		l.Users[id] = all_types.No
		return "Вы были отписаны от рассылки новостей из раздела " + l.MainTitle
	} else {
		l.Users[id] = all_types.Yes
		return "Вы были подписаны на рассылку новостей из раздела " + l.MainTitle
	}
}

func DeleteFitNews(href string) string {
	_, ok := FitNsuNews[href]
	if !ok {
		return "Такой раздел не найден"
	} else {
		delete(FitNsuNews, href)
		return "Раздел " + href + " удалён"
	}
}

func CheckFitHref(href string) bool {
	res, err := http.Get("http://fit.nsu.ru/" + href)
	if err != nil {
		return false
	}

	if res.Status != "200 OK" {
		return false
	}

	return true
}

func (l *NewsList) GetAndRefreshLastNews() (message []string, err error) {
	newPage, err := GetNewPosts(l.Href)
	if err != nil {
		return
	}

	if l.Pages == nil {
		for i := 0; i < len(newPage); i++ {
			message = append(message, newPage[i].String())
		}

		l.Pages = &newPage
		return
	}

	for i := 0; i < len(newPage); i++ {
		flag := true

		for j := 0; j < len(*l.Pages); j++ {
			if (newPage[i].Href == (*l.Pages)[j].Href) || timeParse(newPage[i].Date) < timeParse((*l.Pages)[j].Date) {
				flag = false
				break
			}
		}

		if flag {
			message = append(message, newPage[i].String())
		}
	}

	l.Pages = &newPage

	return
}

func timeParse(date string) int {
	if len(date) != 8 {
		return 0
	}

	day, err := strconv.Atoi(date[0:2])
	if err != nil {
		return 0
	}

	month, err := strconv.Atoi(date[3:5])
	if err != nil {
		return 0
	}

	year, err := strconv.Atoi(date[6:8])
	if err != nil {
		return 0
	}

	return day + 100*month + 10000*year
}

func AddNewNewsList(href string, title string) (answer string) {
	if href == "" {
		return "Ошибочный ввод"
	}

	_, ok := FitNsuNews[href]
	if ok {
		return "Уже существует"
	}

	if !CheckFitHref(href) {
		return "Ошибочная ссылка"
	}

	var list NewsList
	list.Href = href
	list.MainTitle = title
	list.Users = make(map[int]int)

	FitNsuNews[href] = &list

	return "Раздел " + title + " успешно добавлен"
}

func RefreshFitNsuFile() (err error) {
	b, err := json.Marshal(FitNsuNews)
	if err != nil {
		return
	}

	file, err := os.OpenFile(all_types.FitNsuFilename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}

	_, err = file.Write(b)
	if err != nil {
		return
	}

	return file.Close()
}

func LoadFitNsuFile() (err error) {
	file, err := os.OpenFile(all_types.FitNsuFilename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}

	dec := json.NewDecoder(file)
	err = dec.Decode(&FitNsuNews)
	if err != nil {
		return
	}

	err = file.Close()
	return
}

func GetNewPosts(href string) (newPages []NewsPage, err error) {
	body, err := GetFitPage(href, all_types.NewsLimit)
	if err != nil {
		return
	}

	beginInd, endInd, err := mymodule.SearchBeginEnd(body, "<table class=\"category\">", "</table>", 1)
	if err != nil {
		return
	}

	tableText := body[beginInd[0][1]:endInd[0][0]]

	newPages, err = ParseTable(tableText)
	if err != nil {
		return
	}

	return
}

func ParseTable(text string) (np []NewsPage, err error) {
	beginInd, endInd, err := mymodule.SearchBeginEnd(text, "<tr class=\"cat-list.*>", "</tr>", -1)
	if err != nil {
		return
	}

	for ; (len(endInd) > 0) && (endInd[0][0] < beginInd[0][0]); endInd = endInd[1:] {
	}
	if len(endInd) != len(beginInd) {
		return np, errors.New("end > begin")
	}

	for i := range beginInd {
		if i == all_types.MaxCountPosts {
			break
		}

		tableBlock := text[beginInd[i][1]:endInd[i][0]]

		bgInd, eInd, err := mymodule.SearchBeginEnd(tableBlock, "<a href=\"/.*>", "</a>", 1)
		if err != nil {
			continue
		}

		var NewsPage NewsPage
		NewsPage.Href = tableBlock[bgInd[0][0]+9 : bgInd[0][1]-2]
		NewsPage.Title = tableBlock[bgInd[0][1]:eInd[0][0]]

		NewsPage.Title, err = mymodule.ChangeSymbol(NewsPage.Title, "", "\t")
		NewsPage.Title, err = mymodule.ChangeSymbol(NewsPage.Title, "", "\n")

		bgInd, eInd, err = mymodule.SearchBeginEnd(tableBlock, "<td class=\"list-date\">", "</td>", -1)
		if err != nil {
			continue
		}

		date := tableBlock[bgInd[0][1]:eInd[len(eInd)-1][0]]

		date, err = mymodule.ChangeSymbol(date, "", "\t")
		date, err = mymodule.ChangeSymbol(date, "", "\n")
		date, err = mymodule.ChangeSymbol(date, "", " ")

		NewsPage.Date = date
		np = append(np, NewsPage)
	}

	return
}

/*
	Объявления /news/announc
	События /news/news
	Конференции /news/konf
	Конкурсы /news/conc
	Вакансии /news/vac
	Объявления кафедры систем информатики /chairs/ksi/anksi
	Объявления кафедры компьютерных систем /chairs/kks/ankks
	Объявления кафедры общей информатики /chairs/koi/koinews
	Административные приказы /news/administrativnye-prikazy
	Объявления кафедры параллельных вычислений /chairs/kpv/kpvnews
	Объявления кафедры компьютерных технологий /chairs/k-kt/kktnews
*/

func GetFitPage(href string, postfix string) (string, error) {
	res, err := http.Get("http://fit.nsu.ru/" + href + postfix)
	if err != nil {
		return "", err
	}

	if res.Status != "200 OK" {
		return "", err
	}

	textBody, err := ioutil.ReadAll(bufio.NewReader(res.Body))
	if err != nil {
		return "", err
	}

	text := html.UnescapeString(string(textBody))

	return text, err
}

func ChangeBotSubscriptions(id int) string {

	u, ok := all_types.AllUsersInfo[id]
	if !ok {
		return "Ошибка обработки группы, сообщите об этом /feedback"
	}

	if u.PermissionToSend {
		u.PermissionToSend = false
		return "Вы были отписаны от рассылки обновлений"
	} else {
		u.PermissionToSend = true
		return "Вы были подписаны на рассылку обновлений"
	}
}

func ChangeGroupByDomain(domain string, id int) string {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		return "Ошибка обработки группы, сообщите об этом /feedback"
	}

	v, ok := s.UserSubscriptions[id]
	if !ok {
		s.UserSubscriptions[id] = all_types.Yes
		return "Вы были подписаны на рассылку " + s.Name + "."
	}

	if v != 0 {
		s.UserSubscriptions[id] = all_types.No
		return "Вы были отписаны от рассылки " + s.Name + "."
	} else {
		s.UserSubscriptions[id] = all_types.Yes
		return "Вы были подписаны на рассылку " + s.Name + "."
	}

}

func DeleteGroup(domain string) string {
	_, ok := all_types.AllSubscription[domain]
	if !ok {
		return "Такая группа не найдена"
	} else {
		delete(all_types.AllSubscription, domain)
		return "Группа " + domain + " удалена"
	}
}

func ChangeGroupActivity(domain string) string {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		return "Такой группы не существует"
	}

	if s.IsActive {
		s.IsActive = false
		return "Деактивирована группа [" + s.ScreenName + "]"
	} else {
		s.IsActive = true
		return "Активирована группа [" + s.ScreenName + "]"
	}
}

func ShowAllGroups() (groups []string) {
	for i, v := range all_types.AllSubscription {
		groups = append(groups, "["+i+"] "+v.Name+" ["+fmt.Sprint(v.IsActive)+"]"+", всем: "+fmt.Sprint(v.IsReady))
	}
	if len(groups) == 0 {
		groups = append(groups, "Список групп пуст")
	}

	return
}

func ShowAllUsersGroup(domain string) (message []string) {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		message = append(message, "Группа не найдена")
		return
	}

	for i, u := range s.UserSubscriptions {
		message = append(message, "ID: "+fmt.Sprint(i)+", состояние: "+fmt.Sprint(u))
	}

	if len(message) == 0 {
		message = append(message, "Подписки отсутсвуют")
		return
	}

	return
}

func ChangeGroupById(domain string, id int) string {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		return "Такая группа не нейдена"
	}

	u, ok := s.UserSubscriptions[id]
	if !ok {
		s.UserSubscriptions[id] = all_types.Yes
		return "Подписка добавлена для " + fmt.Sprint(id)
	}

	switch u {
	case 1:
		s.UserSubscriptions[id] = all_types.No
		return "Подписка деактивирована для " + fmt.Sprint(id)
	default:
		s.UserSubscriptions[id] = all_types.Yes
		return "Подписка активирована для " + fmt.Sprint(id)
	}
}

func AddNewGroupToParse(domain string) (err error) {
	_, ok := all_types.AllSubscription[domain]
	if ok {
		return errors.New("Группа с таким названием уже существует.")
	}

	g, err := vkapi.GetGroup(0, domain)
	if err != nil {
		return err
	}

	var sub all_types.Subscription

	sub.Name = g.Name
	sub.ScreenName = domain
	sub.UserSubscriptions = make(map[int]int)
	sub.IsActive = false

	/*posts, err := all_types.GetPosts(domain, all_types.MaxCountPosts)
	if err != nil {
		return
	}

	sub.Posts = &posts*/

	all_types.AllSubscription[domain] = &sub

	return nil
}

func GroupReady(domain string) (answer string) {
	s, ok := all_types.AllSubscription[domain]
	if !ok {
		return "Такой группы не существует"
	}

	if s.IsReady {
		s.IsReady = false
		return "Не рассылаю людям [" + s.ScreenName + "]"
	} else {
		s.IsReady = true
		return "Рассылаю людям [" + s.ScreenName + "]"
	}
}
