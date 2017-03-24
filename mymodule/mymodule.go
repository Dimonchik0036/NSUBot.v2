package mymodule

import (
	"errors"
	"regexp"
)

func ChangeSymbol(text string, symbols string, reg string) (string, error) {
	symbolReg, err := regexp.Compile(reg)
	if err != nil {
		return "", err
	}

	for index := symbolReg.FindStringIndex(text); len(index) > 0; index = symbolReg.FindStringIndex(text) {
		text = text[:index[0]] + symbols + text[index[1]:]
	}

	return text, err
}

func SearchBeginEnd(text string, symbolBegin string, symbolEnd string, count int) (beginIndex [][]int, endIndex [][]int, err error) {
	beginReg, err := regexp.Compile(symbolBegin)
	if err != nil {
		return nil, nil, err
	}

	endReg, err := regexp.Compile(symbolEnd)
	if err != nil {
		return nil, nil, err
	}

	beginIndex = beginReg.FindAllStringIndex(text, count)
	endIndex = endReg.FindAllStringIndex(text, count)

	if len(beginIndex) == 0 || len(endIndex) == 0 || beginIndex[0][1] > endIndex[len(endIndex)-1][0] {
		return nil, nil, errors.New("Ошибка индексирования.")
	}

	return
}
