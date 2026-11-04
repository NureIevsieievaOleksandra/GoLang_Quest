package main

import (
	"Quest/challenges"
	"Quest/events"
	"Quest/models"
	"Quest/utils"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type menuItem struct {
	title string
	event events.Event
	back  bool
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	state := &models.Survivor{
		PlayerName:  "",
		Points:      0,
		AirHours:    0,
		WaterLiters: 0,
		FoodMeals:   0,
		IsAlive:     true,
	}

	fmt.Println("=== СИСТЕМА КВЕСТУ АКТИВОВАНА ===")
	fmt.Println("«Про те, що ти тут, ніхто не знає. Де це, ти не знаєш.")
	fmt.Println("Єдиний спосіб вибратися — набрати щонайменше 60 балів у випробуваннях.»")
	fmt.Println("-------------------------------------------------")

	events.ApplyEvent(state,
		events.AddPoints(3+5),
		events.AddWater(6),
		func(s *models.Survivor) {
			s.Task3Done = true
			s.Task3StarDone = true
		},
	)
	events.ApplyEvent(state,
		events.AddPoints(9),
		func(s *models.Survivor) { s.Task9Done = true },
	)
	state.EnsurePlayerInJournal()

	runMenu(reader, state, "ГОЛОВНЕ МЕНЮ", func() []menuItem {
		return []menuItem{
			{title: "Випробування", event: func(s *models.Survivor) { challengeSelectMenu(reader, s) }},
			{title: "Запаси (журнал виживання)", event: func(s *models.Survivor) { challenges.RunJournal(reader, s) }},
			{title: "Бали", event: showScore},
			{title: "Зробити крок / прожити час (зменшити запаси)", event: func(s *models.Survivor) {
				fmt.Print("Скільки годин минуло? ")
				events.ApplyEvent(s, events.PassTime(utils.ReadFloat(reader)))
			}},
			{title: "Вихід", event: exitEvent},
		}
	})
}

func runMenu(reader *bufio.Reader, state *models.Survivor, title string, build func() []menuItem) {
MenuLoop:
	if state.GameOver {
		return
	}

	items := build()
	fmt.Printf("\n--- %s ---\n", title)
	printMenuItems(items, 0)
	fmt.Printf("Оберіть пункт (1-%d): ", len(items))

	n, err := strconv.Atoi(strings.TrimSpace(utils.ReadString(reader)))
	if err != nil || n < 1 || n > len(items) {
		fmt.Printf("Невідома команда. Будь ласка, введіть число від 1 до %d.\n", len(items))
		goto MenuLoop
	}

	item := items[n-1]
	if item.back {
		return
	}
	events.ApplyEvent(state, item.event, events.CheckVictory())

	goto MenuLoop
}

func printMenuItems(items []menuItem, i int) {
	if i >= len(items) {
		return
	}
	fmt.Printf("%d. %s\n", i+1, items[i].title)
	printMenuItems(items, i+1)
}

func exitEvent(s *models.Survivor) {
	if s.ExitUnlocked {
		events.ApplyEvent(s, events.WinGame())
		return
	}
	fmt.Println("\nТермінал вимкнено.")
	s.GameOver = true
}

func challengeSelectMenu(reader *bufio.Reader, state *models.Survivor) {
	runMenu(reader, state, "ВИБІР ВИПРОБУВАННЯ", func() []menuItem {
		return []menuItem{
			{title: fmt.Sprintf("Випробування 1. Наближені обчислення (Завдання 4) [%s]", utils.StatusStr(state.Task4Done)),
				event: func(s *models.Survivor) { challenges.RunRoundingChallenge(reader, s) }},
			{title: fmt.Sprintf("Випробування 2. Матриця пасток (Завдання 5) [%s]", utils.StatusStr(state.Task5Done)),
				event: func(s *models.Survivor) { challenges.RunMatrixChallenge(s) }},
			{title: fmt.Sprintf("Випробування 3. Записна книжка рішень (Завдання 6) [%s]", utils.StatusStr(state.Task6Done)),
				event: func(s *models.Survivor) { challenges.RunNotebookChallenge(reader, s) }},
			{title: fmt.Sprintf("Випробування 4. Озоногенератор і пароль (Завдання 7) [%s]", utils.StatusStr(state.Task7Done)),
				event: func(s *models.Survivor) { challenges.RunCipherChallenge(reader, s) }},
			{title: fmt.Sprintf("Випробування 5. Журнал виживання (Завдання 8) [%s]", utils.StatusStr(state.Task8Done)),
				event: func(s *models.Survivor) { challenges.RunJournal(reader, s) }},
			{title: fmt.Sprintf("Випробування 6. Інтерфейси (Завдання 10) [%s]", utils.StatusStr(state.Task10Done)),
				event: func(s *models.Survivor) { challenges.RunInterfaceChallenge(s) }},
			{title: "Інші випробування (Ім'я, Розрахунок повітря)",
				event: func(s *models.Survivor) { handleLegacyChallenges(reader, s) }},
			{title: "Назад", back: true},
		}
	})
}

func handleLegacyChallenges(reader *bufio.Reader, state *models.Survivor) {
	runMenu(reader, state, "СПИСОК ВИПРОБУВАНЬ", func() []menuItem {
		return []menuItem{
			{title: fmt.Sprintf("Завдання 1: Ввести ім'я [%s]", utils.StatusStr(state.Task1Done)),
				event: func(s *models.Survivor) { enterNameTask(reader, s) }},
			{title: fmt.Sprintf("Завдання 2: Розрахунок повітря кімнати [%s]", utils.StatusStr(state.Task2Done)),
				event: roomAirTask},
			{title: "Назад", back: true},
		}
	})
}

func enterNameTask(reader *bufio.Reader, s *models.Survivor) {
	if s.Task1Done {
		fmt.Printf("Ім'я вже встановлено: %s\n", s.PlayerName)
		return
	}
	fmt.Print("Введіть ваше ім'я: ")
	name := strings.TrimSpace(utils.ReadString(reader))
	events.ApplyEvent(s,
		func(s *models.Survivor) {
			s.PlayerName = name
			s.Task1Done = true
		},
		events.AddPoints(1),
	)
	s.EnsurePlayerInJournal()
	fmt.Printf("Ім'я збережено: %s \n", s.PlayerName)
}

func roomAirTask(s *models.Survivor) {
	volume := 3.0 * 4.0 * 2.0
	hours := volume / 1.0
	fmt.Println("\n[Розрахунок запасів повітря кімнати]")
	fmt.Println("Розміри кімнати: 3м * 4м * 2м")
	fmt.Printf("Об'єм кімнати: %.1f куб. м.\n", volume)
	fmt.Printf("Витрата: 1 куб. м/год -> Повітря у приміщенні вистачить на: %.1f год.\n", hours)

	if s.Task2Done {
		fmt.Println("(Ви вже отримали бали та бонус за це завдання, але можете переглядати розрахунок повторно)")
		return
	}
	events.ApplyEvent(s,
		events.AddPoints(1+2),
		events.AddAir(hours+24.0),
		func(s *models.Survivor) { s.Task2Done = true },
	)
	fmt.Println("Вірно! Нараховано +3 бали (Завдання 2 та 2*) і надано бонусний запас повітря на 24 години.")
}

func showScore(state *models.Survivor) {
	fmt.Println("\n========== СТАТУС БАЛІВ ==========")
	if state.PlayerName != "" {
		fmt.Printf("Гравець: %s\n", state.PlayerName)
	} else {
		fmt.Println("Гравець: [Не ідентифіковано]")
	}
	fmt.Printf("Набрано балів: %d / 60 для виходу\n", state.Points)
	fmt.Printf("Виконані етапи: Завд.1: %t | Завд.2: %t | Завд.3(інтерфейс): %t | Завд.4: %t\n",
		state.Task1Done, state.Task2Done, state.Task3Done, state.Task4Done)
	fmt.Printf("                Завд.5: %t | Завд.6: %t | Завд.7: %t | Завд.8: %t | Завд.9: %t\n",
		state.Task5Done, state.Task6Done, state.Task7Done, state.Task8Done, state.Task9Done)
	fmt.Println("==================================")
}
