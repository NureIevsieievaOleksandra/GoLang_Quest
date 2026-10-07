package challenges

import (
	"Quest/models"
	"Quest/utils"
	"bufio"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

func RunRoundingChallenge(reader *bufio.Reader, state *models.GameState) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 1: НАБЛИЖЕНІ ОБЧИСЛЕННЯ ===")
	fmt.Println("Округліть число до будь-якої зручної кількості знаків.")
	fmt.Println("Потрібно дати 3 правильні відповіді поспіль.")

	correctCount := 0

RoundLoop:
	if correctCount == 3 {
		fmt.Println("\nВІТАЄМО! Випробування успішно пройдено!")
		if !state.Task4Done {
			foodReward := 1 + 9
			waterReward := 6.0
			pointsReward := 7

			state.Points += pointsReward
			state.FoodMeals += foodReward
			state.WaterLiters += waterReward
			state.Task4Done = true

			fmt.Printf("Нараховано: +%d балів!\n", pointsReward)
			fmt.Printf("Отримано: %d прийомів їжі та %.1f л води!\n", foodReward, waterReward)
		} else {
			fmt.Println("Бали за це завдання вже було отримано раніше.")
		}
		return
	}

	generatedNumStr := generateRandomFloatString(3, 7)
	fmt.Printf("\n[Успішно: %d/3]\n", correctCount)
	fmt.Printf("К: %s\n", generatedNumStr)
	fmt.Print("П: ")

	userInput := strings.TrimSpace(utils.ReadString(reader))
	userInput = strings.ReplaceAll(userInput, ",", ".")
	generatedNumStr = strings.ReplaceAll(generatedNumStr, ",", ".")

	if isCorrectRoundingWithoutLoops(generatedNumStr, userInput) {
		fmt.Println("К: вірно.")
		correctCount++
	} else {
		fmt.Println("К: не так.")
		correctCount = 0
	}

	goto RoundLoop
}

func generateRandomFloatString(minDec, maxDec int) string {
	intPart := rand.Intn(9) + 1
	decDigitsCount := minDec + rand.Intn(maxDec-minDec+1)

	var buildDec func(count int) string
	buildDec = func(count int) string {
		if count <= 0 {
			return ""
		}
		digit := strconv.Itoa(rand.Intn(10))
		return digit + buildDec(count-1)
	}

	return fmt.Sprintf("%d.%s", intPart, buildDec(decDigitsCount))
}

func isCorrectRoundingWithoutLoops(origStr, userStr string) bool {
	findDot := func(s string) int {
		var helper func(idx int) int
		helper = func(idx int) int {
			if idx >= len(s) {
				return -1
			}
			if s[idx] == '.' {
				return idx
			}
			return helper(idx + 1)
		}
		return helper(0)
	}

	userDot := findDot(userStr)
	var userDecPlaces int
	if userDot == -1 {
		userDecPlaces = 0
	} else {
		userDecPlaces = len(userStr) - userDot - 1
	}

	origDot := findDot(origStr)
	if origDot == -1 {
		return origStr == userStr
	}

	origDecPlaces := len(origStr) - origDot - 1
	if userDecPlaces >= origDecPlaces {
		return false
	}

	cutIndex := origDot + 1 + userDecPlaces
	if cutIndex >= len(origStr) {
		return false
	}

	nextDigit := int(origStr[cutIndex] - '0')
	roundUpNeeded := nextDigit >= 5

	var truncatedOrig string
	if userDecPlaces == 0 {
		truncatedOrig = origStr[:origDot]
	} else {
		truncatedOrig = origStr[:cutIndex]
	}

	if !roundUpNeeded {
		return truncatedOrig == userStr
	}

	return isIncrementMatch(truncatedOrig, userStr)
}

func isIncrementMatch(truncated, user string) bool {
	cleanTrunc := strings.ReplaceAll(truncated, ".", "")
	cleanUser := strings.ReplaceAll(user, ".", "")

	var addOne func(s string, idx int) string
	addOne = func(s string, idx int) string {
		if idx < 0 {
			return "1" + s
		}
		digit := int(s[idx] - '0')
		if digit < 9 {
			return s[:idx] + string('0'+byte(digit+1)) + s[idx+1:]
		}
		newS := s[:idx] + "0" + s[idx+1:]
		return addOne(newS, idx-1)
	}

	incremented := addOne(cleanTrunc, len(cleanTrunc)-1)
	return incremented == cleanUser
}
