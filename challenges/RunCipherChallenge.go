package challenges

import (
	"Quest/events"
	"Quest/models"
	"Quest/utils"
	"bufio"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const hintText = `у цьому рядку зашифрований пароль, перша буква пароля - це літера, що найчастіше
зустрічається, яка буде в слові, що найчастіше зустрічається, ця літера повинна
бути конкатенована зі словом, розташованим по середині списку слів що є в
підказці, слова повинні бути відсортовані за зростанням частоти зустрічання,
потім повинні бути конкатеновані повторені стільки разів, скільки зустрічається
друга за рідкістю літера в тексті Підказка. Слова у різному регістрі – це різні
слова, літери у різному регістрі – це однакові літери. Завдання вирішується з
використанням рядків та функцій рядків. Після того, як підсумковий рядок буде
складений, його потрібно буде розгорнути по рунах. Це і буде правильна
відповідь. Після того як Ви отримаєте відповідь, його потрібно буде руками
ввести в інтерфейс користувача. Не забудьте перевірити правильність введення.`

func RunCipherChallenge(reader *bufio.Reader, state *models.GameState) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 4: ОЗОНОГЕНЕРАТОР ТА ПАРОЛЬ (ЗАВДАННЯ 7) ===")
	fmt.Println("УВАГА: Озоногенератор загрожує спалити всі запаси їжі!")
	fmt.Println("Розшифруйте підказку та введіть деактиваційний пароль.")
	fmt.Println("\n[Текст підказки]:")
	fmt.Println(hintText)

	expectedPassword := generateDecryptedPassword(hintText)

	fmt.Print("\nВведіть згенерований пароль: ")
	userInput := strings.TrimSpace(utils.ReadString(reader))

	if userInput == expectedPassword {
		fmt.Println("\nПароль правильний! Озоногенератор безпечно деактивовано.")
		if !state.Task7Done {
			pointsReward := 9
			foodReward := 3
			waterReward := 2.0

			events.ApplyEvent(state,
				events.AddPoints(pointsReward),
				events.AddSupplies(models.DaySupplies{Food: foodReward, Water: waterReward}),
				func(s *models.Survivor) { s.Task7Done = true },
			)

			fmt.Printf("Нараховано: +%d балів!\n", pointsReward)
			fmt.Printf("Отримано запас їжі та води на 1 день (+%d їжі, +%.1f л води)!\n", foodReward, waterReward)
		} else {
			fmt.Println("(Ви вже отримували нагороду за це завдання)")
		}
	} else {
		fmt.Println("\nПароль неправильний!")
		fmt.Printf("(Підказка системи: правильний пароль був \"%s\")\n", expectedPassword)
		fmt.Println("Озоногенератор спалив запаси їжі! Залишився лише сьогоднішній раціон.")
		events.ApplyEvent(state, events.KeepOnlyFood(3))
	}
}

func generateDecryptedPassword(text string) string {
	rawWords := strings.Fields(text)
	wordFreq := make(map[string]int)

	for _, w := range rawWords {
		cleaned := strings.Trim(w, ".,:;\"«»-–")
		if cleaned != "" {
			wordFreq[cleaned]++
		}
	}

	type wordItem struct {
		Word  string
		Count int
	}

	var wordsList []wordItem
	for w, c := range wordFreq {
		wordsList = append(wordsList, wordItem{Word: w, Count: c})
	}

	sort.Slice(wordsList, func(i, j int) bool {
		if wordsList[i].Count == wordsList[j].Count {
			return wordsList[i].Word < wordsList[j].Word
		}
		return wordsList[i].Count < wordsList[j].Count
	})

	midWord := wordsList[len(wordsList)/2].Word
	maxWordFreq := wordsList[len(wordsList)-1].Count

	midWordLetterFreq := make(map[rune]int)
	for _, r := range strings.ToLower(midWord) {
		if unicode.IsLetter(r) {
			midWordLetterFreq[r]++
		}
	}

	type letterItem struct {
		Letter rune
		Count  int
	}

	var midLetters []letterItem
	for r, c := range midWordLetterFreq {
		midLetters = append(midLetters, letterItem{Letter: r, Count: c})
	}
	sort.Slice(midLetters, func(i, j int) bool {
		if midLetters[i].Count == midLetters[j].Count {
			return midLetters[i].Letter < midLetters[j].Letter
		}
		return midLetters[i].Count > midLetters[j].Count
	})
	mostFreqLetterInMidWord := midLetters[0].Letter

	allLettersFreq := make(map[rune]int)
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) {
			allLettersFreq[r]++
		}
	}

	var allLetters []letterItem
	for r, c := range allLettersFreq {
		allLetters = append(allLetters, letterItem{Letter: r, Count: c})
	}

	sort.Slice(allLetters, func(i, j int) bool {
		if allLetters[i].Count == allLetters[j].Count {
			return allLetters[i].Letter < allLetters[j].Letter
		}
		return allLetters[i].Count < allLetters[j].Count
	})

	maxLetterFreq := allLetters[len(allLetters)-1].Count

	secondRarestCount := 1
	if len(allLetters) > 1 {
		secondRarestCount = allLetters[1].Count
	}

	head := string(mostFreqLetterInMidWord) + midWord
	numberPair := strconv.Itoa(maxWordFreq) + strconv.Itoa(maxLetterFreq)
	tail := strings.Repeat(numberPair, secondRarestCount)

	combinedString := head + tail

	return reverseByRune(combinedString)
}

func reverseByRune(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
