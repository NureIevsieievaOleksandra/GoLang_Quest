package utils

import (
	"bufio"
	"strconv"
	"strings"
)

func ReadString(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return input
}

func ReadFloat(reader *bufio.Reader) float64 {
	input, _ := reader.ReadString('\n')
	val, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		return 0.0
	}
	return val
}

func StatusStr(done bool) string {
	if done {
		return "Виконано"
	}
	return "Не пройдено"
}
