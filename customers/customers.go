package customers

// AddGroupNumber Привязывает к пользователю номер группы.
func AddGroupNumber(scheduleMap map[string][7]string, userGroup map[int]string, id int, group string) string {
	if group == "" {
		return "Вы не ввели номер группы."
	}

	if len(group) > 16 {
		return "Слишком много символов."
	}

	_, ok := scheduleMap[group]
	if !ok {
		group += ".1"
		_, ok = scheduleMap[group]
		if !ok {
			return "Введён некорректный номер группы, попробуйте повторить попытку или воспользоваться /help и /faq для помощи."
		}
	}

	userGroup[id] = group

	return "Группа '" + group + "' успешно назначена, нажмите на /today или /tomorrow для проверки правильности выбора."
}
