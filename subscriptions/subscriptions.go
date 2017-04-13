package subscriptions

import (
	"TelegramBot/all_types"
	"TelegramBot/vkapi"
	"errors"
	"fmt"
)

func ChangeSubscriptions(argument string, id int) string {
	switch argument {
	case all_types.News:
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
	default:
		return ChangeGroupByDomain(argument, id)
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
	} else {
		if v != 0 {
			s.UserSubscriptions[id] = all_types.No
			return "Вы были отписаны от рассылки " + s.Name + "."
		} else {
			s.UserSubscriptions[id] = all_types.Yes
			return "Вы были подписаны на рассылку " + s.Name + "."
		}
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
		groups = append(groups, "["+i+"] "+v.Name+" ["+fmt.Sprint(v.IsActive)+"]")
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
