package customers

import (
	"encoding/json"
	"os"
	"regexp"
)

const MaxCountLabel = 6
const LabelsFile = "labels.txt"
const MyGroupLabel = "Моя"

type UserGroup struct {
	Group   map[string]string
	MyGroup string
}

type UserGroupLabels struct {
	Id     int    `json:"ID"`
	Labels string `json:"Labels"`
}

type UserLabels struct {
	Label string `json:"Label"`
	Group string `json:"Group"`
}

var AllLabels = make(map[int]UserGroup)

func UpdateUserLabels() error {
	userFile, err := os.OpenFile(LabelsFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	for i, v := range AllLabels {
		var user UserGroupLabels
		var text string

		user.Id = i

		var lab UserLabels
		lab.Label = MyGroupLabel
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

func PrintUserLabels(group UserGroup) (userLabels string) {
	if group.MyGroup != "" {
		userLabels = "Стандартная группа " + group.MyGroup + ".\n"
	}

	for l, g := range group.Group {
		userLabels += "Для группы " + g + " назначена метка \"" + l + "\".\n"
	}

	if userLabels == "" {
		userLabels = "Метки отсутствуют."
	}

	return
}

func DeleteUserLabels(userGroup UserGroup) string {
	if len(userGroup.Group) == 0 {
		return "Список меток пуст."
	}

	for i := range userGroup.Group {
		delete(userGroup.Group, i)
	}

	return "Были очищены все метки, кроме стандартной."
}

func GroupDecomposition(commang string) (group string, labelGroup string) {
	labelReg, err := regexp.Compile("[^ ]+")
	if err != nil {
		return "", ""
	}

	labelText := labelReg.FindAllString(commang, 2)

	if len(labelText) > 0 {
		group = labelText[0]
	}

	if len(labelText) > 1 {
		labelGroup = labelText[1]
	}

	return
}

// AddGroupNumber Привязывает к пользователю номер группы.
func AddGroupNumber(schedule map[string][7]string, id int, command string) string {
	group, labelGroup := GroupDecomposition(command)
	if group == "" {
		return "Вы не ввели номер группы."
	}

	if labelGroup == "" {
		labelGroup = MyGroupLabel
	}

	if (len(group) > 16) || (len(labelGroup) > 16) {
		return "Слишком много символов."
	}

	_, ok := schedule[group]
	if !ok {
		group += ".1"
		_, ok = schedule[group]
		if !ok {
			return "Введён некорректный номер группы, попробуйте повторить попытку или воспользоваться /help и /faq для помощи."
		}
	}

	v := AllLabels[id]
	var okay bool

	if labelGroup == MyGroupLabel {
		v.MyGroup = group
	} else {
		if v.Group == nil {
			v.Group = make(map[string]string)
		}

		if len(v.Group) > MaxCountLabel+1 {
			return "Превышен лимит меток."
		}

		_, okay = v.Group[labelGroup]
		if !okay && (len(v.Group) == MaxCountLabel) {
			return "Вы достигли предела меток. Вы можете изменять группы, привязанные к меткам, но не можете добавлять новые.\n" +
				"Вы можете очистить список меток, воспользовавшись командой /clearlabels."
		}

		if len(v.Group) == MaxCountLabel {

		}

		v.Group[labelGroup] = group
	}

	AllLabels[id] = v

	if labelGroup == MyGroupLabel {
		return "Группа '" + group + "' успешно назначена стандартной."
	} else {
		if okay {
			return "Изменена группа у метки \"" + labelGroup + "\" на " + group + "."
		} else {
			return "Добавлена новая метка '" + labelGroup + "' для группы " + group + "."
		}
	}
}
