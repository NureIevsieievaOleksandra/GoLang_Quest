package challenges

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

func RunCipherChallenge() {
	text := `у цьому рядку зашифрований пароль, перша буква пароля - це літера, що найчастіше
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

	rawWords := strings.Fields(text)
	wordFreq := make(map[string]int)
	for _, w := range rawWords {
		cleaned := strings.Trim(w, ".,:;\"«»-–")
		if cleaned != "" {
			wordFreq[cleaned]++
		}
	}

	type wordCount struct {
		Word  string
		Count int
	}
	var wordsList []wordCount
	for w, c := range wordFreq {
		wordsList = append(wordsList, wordCount{Word: w, Count: c})
	}
	sort.Slice(wordsList, func(i, j int) bool {
		if wordsList[i].Count == wordsList[j].Count {
			return wordsList[i].Word < wordsList[j].Word
		}
		return wordsList[i].Count < wordsList[j].Count
	})

	midWord := wordsList[len(wordsList)/2].Word

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

	maxWordFreq := wordsList[len(wordsList)-1].Count

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

	maxLetterFreq := allLetters[len(allLetters)-1].Count // найчастіша літера у тексті

	secondRarestLetter := allLetters[1]

	fmt.Printf("1. Слово з середини списку: \"%s\"\n", midWord)
	fmt.Printf("   Найчастіша літера в ньому: '%c' (зустрічається %d раз)\n\n",
		mostFreqLetterInMidWord, midLetters[0].Count)

	fmt.Println("2. Частоти:")
	fmt.Printf("   • Частота найчастішого слова (\"%s\"): %d\n",
		wordsList[len(wordsList)-1].Word, maxWordFreq)
	fmt.Printf("   • Частота найчастішої літери в тексті ('%c'): %d\n\n",
		allLetters[len(allLetters)-1].Letter, maxLetterFreq)

	fmt.Println("3. Рідкісні літери у тексті (початок списку):")
	for i := 0; i < 5 && i < len(allLetters); i++ {
		fmt.Printf("   [%d] '%c' : %d раз(и)\n", i, allLetters[i].Letter, allLetters[i].Count)
	}
	fmt.Printf("   -> Друга за рідкістю літера: '%c' з частотою %d\n",
		secondRarestLetter.Letter, secondRarestLetter.Count)
}
