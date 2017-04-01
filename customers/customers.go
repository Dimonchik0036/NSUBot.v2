package customers

import (
	"encoding/json"
	"os"
	"regexp"
)

const MaxCountLabel = 20
const LabelsFile = "labels.txt"
const MyGroupLabel = "Моя"
const MaxCountSymbol = 64

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

func PrintUserLabels(id int) (userLabels string) {
	g, ok := AllLabels[id]
	if !ok || (g.MyGroup == "" && len(g.Group) == 0) {
		return "Метки отсутствуют."
	}

	userLabels = "Список меток:"

	if g.MyGroup != "" {
		userLabels = "Моя группа " + g.MyGroup + ".\n"
	}

	for l, g := range g.Group {
		userLabels += "У группы " + g + "  метка \"" + l + "\".\n"
	}

	return
}

func DeleteUserLabels(id int) string {
	g, ok := AllLabels[id]
	if !ok || g.Group == nil || len(g.Group) == 0 {
		return "Список меток пуст."
	}

	for i := range g.Group {
		delete(g.Group, i)
	}

	return "Были очищены все метки."
}

func GroupDecomposition(words string) (group string, labelGroup string) {
	labelReg, err := regexp.Compile("[^ ]+")
	if err != nil {
		return "", ""
	}

	labelText := labelReg.FindAllStringIndex(words, 2)

	if len(labelText) > 0 {
		group = words[:labelText[0][1]]
	}

	if len(labelText) > 1 {
		buf := []byte(words[labelText[1][0]:])
		for ; (len(buf) > 0) && (buf[len(buf)-1] == ' '); buf = buf[:len(buf)-1] {
		}

		labelGroup = string(buf)
	}

	return
}

// AddGroupNumber Привязывает к пользователю номер группы.
func AddGroupNumber(schedule map[string][7]string, id int, command string) (bool, string) {
	group, labelGroup := GroupDecomposition(command)
	if group == "" {
		return false, "Вы не ввели номер группы, попробуте ещё раз:"
	}

	if labelGroup == "" {
		labelGroup = MyGroupLabel
	}

	if (len(group) > 16) || (len(labelGroup) > MaxCountSymbol) {
		return false, "Слишком много символов, повторите попытку:"
	}

	_, ok := schedule[group]
	if !ok {
		group += ".1"
		_, ok = schedule[group]
		if !ok {
			return false, "Введён некорректный номер группы, попробуйте повторить попытку:"
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
			return false, "Предел"
		}

		_, okay = v.Group[labelGroup]
		if !okay && (len(v.Group) == MaxCountLabel) {
			return false, "Предел"
		}

		if len(v.Group) == MaxCountLabel {

		}

		v.Group[labelGroup] = group
	}

	AllLabels[id] = v

	if labelGroup == MyGroupLabel {
		return true, "Группа '" + group + "' успешно назначена стандартной."
	} else {
		if okay {
			return true, "Изменена группа у метки \"" + labelGroup + "\" на " + group + "."
		} else {
			return true, "Добавлена новая метка '" + labelGroup + "' для группы " + group + "."
		}
	}
}
