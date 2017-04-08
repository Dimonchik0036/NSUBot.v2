package customers

import (
	"encoding/json"
	"os"
	"regexp"
	"TelegramBot/types"
)

func UpdateUserLabels() error {
	userFile, err := os.OpenFile(types.LabelsFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	for i, v := range types.AllLabels {
		var user types.UserGroupLabels
		var text string

		user.Id = i

		var lab types.UserLabels
		lab.Label = types.MyGroupLabel
		lab.Group = v.MyGroup

		res, err := json.Marshal(&lab)
		if err == nil {
			text += string(res)
		}

		for l, g := range v.Group {
			lab.Label = l
			lab.Group = g

			res, err = json.Marshal(&lab)
			if err != nil {
				continue
			}

			text += string(res)
		}

		user.Labels = text

		out, err := json.Marshal(&user)
		if err != nil {
			continue
		}

		userFile.WriteString(string(out) + "\n")
	}

	err = userFile.Close()

	return err
}

func PrintUserLabels(id int) (userLabels string) {
	g, ok := types.AllLabels[id]
	if !ok || (g.MyGroup == "" && len(g.Group) == 0) {
		return "Метки отсутствуют"
	}

	userLabels = "Список меток"

	if g.MyGroup != "" {
		userLabels = "Моя группа " + g.MyGroup + "\n"
	}

	for l, g := range g.Group {
		userLabels += "У группы " + g + "  метка \"" + l + "\"\n"
	}

	return
}

func DeleteUserLabels(id int) string {
	g, ok := types.AllLabels[id]
	if !ok || g.Group == nil || len(g.Group) == 0 {
		return "Список меток пуст"
	}

	for i := range g.Group {
		delete(g.Group, i)
	}

	return "Были очищены все метки"
}

func DecomposeQuery(words string) (command string, arguments string) {
	labelReg, err := regexp.Compile("[^ ]*")
	if err != nil {
		return "", ""
	}

	index := labelReg.FindStringIndex(words)

	if len(index) > 0 {
		command = words[index[0]:index[1]]

		if len(words) > index[1] {
			arguments = words[index[1]+1:]
		}
	}

	return
}

// AddGroupNumber Привязывает к пользователю номер группы.
func AddGroupNumber(id int, command string) (int, string) {
	group, labelGroup := DecomposeQuery(command)
	if group == "" {
		return 0, "Вы не ввели номер группы, попробуте ещё раз:"
	}

	if labelGroup == "" {
		labelGroup = types.MyGroupLabel
	}

	if (len(group) > 16) || (len(labelGroup) > types.MaxCountSymbol) {
		return 0, "Слишком много символов, повторите попытку:"
	}

	_, ok := types.AllSchedule[group]
	if !ok {
		group += ".1"
		_, ok = types.AllSchedule[group]
		if !ok {
			return 0, "Введён некорректный номер группы, попробуйте повторить попытку:"
		}
	}

	v := types.AllLabels[id]
	var okay bool

	if labelGroup == types.MyGroupLabel {
		v.MyGroup = group
	} else {
		if v.Group == nil {
			v.Group = make(map[string]string)
		}

		if len(v.Group) > types.MaxCountLabel+1 {
			return 2, "Предел"
		}

		_, okay = v.Group[labelGroup]
		if !okay && (len(v.Group) == types.MaxCountLabel) {
			return 2, "Предел"
		}

		if len(v.Group) == types.MaxCountLabel {

		}

		v.Group[labelGroup] = group
	}

	types.AllLabels[id] = v

	if labelGroup == types.MyGroupLabel {
		return 1, "Группа '" + group + "' успешно назначена стандартной"
	} else {
		if okay {
			return 1, "Изменена группа у метки \"" + labelGroup + "\" на " + group
		} else {
			return 1, "Добавлена новая метка '" + labelGroup + "' для группы " + group
		}
	}
}
