package main

import (
	"Quest/challenges"
	"Quest/survival"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type menuItem struct {
	title string
	event survival.Event
	back  bool
}

type ActionWrapper struct {
	Event survival.Event
}

func (a ActionWrapper) Execute(s *survival.Survivor) {
	survival.ApplyEvent(s, a.Event)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	state := &survival.Survivor{
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

	survival.ApplyEvent(state,
		survival.AddPoints(3+5),
		survival.AddWater(6),
		func(s *survival.Survivor) {
			s.Task3Done = true
			s.Task3StarDone = true
		},
	)
	survival.ApplyEvent(state,
		survival.AddPoints(9),
		func(s *survival.Survivor) { s.Task9Done = true },
	)
	survival.ApplyEvent(state,
		survival.AddPoints(2),
		func(s *survival.Survivor) { s.Task14Done = true },
	)
	state.EnsurePlayerInJournal()

	runMenu(reader, state, "ГОЛОВНЕ МЕНЮ", func() []menuItem {
		return []menuItem{
			{title: "Випробування", event: func(s *survival.Survivor) { challengeSelectMenu(reader, s) }},
			{title: "Запаси (журнал виживання)", event: func(s *survival.Survivor) { challenges.RunJournal(reader, s) }},
			{title: "Бали", event: showScore},
			{title: "Зробити крок / прожити час (безпечне виконання - Завдання 11)", event: func(s *survival.Survivor) {
				fmt.Print("Скільки годин минуло? ")
				hours := survival.ReadFloat(reader)

				if hours <= 0 {
					fmt.Println("Час не може бути від'ємним або нульовим.")
					return
				}

				fmt.Printf("\n--- Минуло %.1f год. Симуляція витрати запасів (Завдання 11) ---\n", hours)

				waterToConsume := (2.0 / 24.0) * hours
				foodToConsume := int(hours / 8.0)

				actions := []survival.Action{
					ActionWrapper{Event: survival.ConsumeAir(hours)},
				}

				if waterToConsume > 0 {
					actions = append(actions, ActionWrapper{Event: survival.ConsumeWater(waterToConsume)})
				}
				if foodToConsume > 0 {
					actions = append(actions, ActionWrapper{Event: survival.ConsumeFood(foodToConsume)})
				}
				survival.ExecuteSafeActions(actions, s)

				fmt.Printf("Поточний стан: Їжа=%d, Вода=%.2f, Повітря=%.1f\n", s.FoodMeals, s.WaterLiters, s.AirHours)

				if !s.Task11Done {
					s.Task11Done = true
					survival.ApplyEvent(s, survival.AddPoints(7))
					fmt.Println("\n[Система] Завдання 11 (Обробка панік та помилок) зараховано! +7 балів.")
				}
			}},
			{title: "Вихід", event: exitEvent},
		}
	})
}

func runMenu(reader *bufio.Reader, state *survival.Survivor, title string, build func() []menuItem) {
MenuLoop:
	if state.GameOver {
		return
	}

	items := build()
	fmt.Printf("\n--- %s ---\n", title)
	printMenuItems(items, 0)
	fmt.Printf("Оберіть пункт (1-%d): ", len(items))

	n, err := strconv.Atoi(strings.TrimSpace(survival.ReadString(reader)))
	if err != nil || n < 1 || n > len(items) {
		fmt.Printf("Невідома команда. Будь ласка, введіть число від 1 до %d.\n", len(items))
		goto MenuLoop
	}

	item := items[n-1]
	if item.back {
		return
	}
	survival.ApplyEvent(state, item.event, survival.CheckVictory())

	goto MenuLoop
}

func printMenuItems(items []menuItem, i int) {
	if i >= len(items) {
		return
	}
	fmt.Printf("%d. %s\n", i+1, items[i].title)
	printMenuItems(items, i+1)
}

func exitEvent(s *survival.Survivor) {
	if s.ExitUnlocked {
		survival.ApplyEvent(s, survival.WinGame())
		return
	}
	fmt.Println("\nТермінал вимкнено.")
	s.GameOver = true
}

func challengeSelectMenu(reader *bufio.Reader, state *survival.Survivor) {
	runMenu(reader, state, "ВИБІР ВИПРОБУВАННЯ", func() []menuItem {
		return []menuItem{
			{title: fmt.Sprintf("Випробування 1. Наближені обчислення (Завдання 4) [%s]", survival.StatusStr(state.Task4Done)),
				event: func(s *survival.Survivor) { challenges.RunRoundingChallenge(reader, s) }},
			{title: fmt.Sprintf("Випробування 2. Матриця пасток (Завдання 5) [%s]", survival.StatusStr(state.Task5Done)),
				event: func(s *survival.Survivor) { challenges.RunMatrixChallenge(s) }},
			{title: fmt.Sprintf("Випробування 3. Записна книжка рішень (Завдання 6) [%s]", survival.StatusStr(state.Task6Done)),
				event: func(s *survival.Survivor) { challenges.RunNotebookChallenge(reader, s) }},
			{title: fmt.Sprintf("Випробування 4. Озоногенератор і пароль (Завдання 7) [%s]", survival.StatusStr(state.Task7Done)),
				event: func(s *survival.Survivor) { challenges.RunCipherChallenge(reader, s) }},
			{title: fmt.Sprintf("Випробування 5. Журнал виживання (Завдання 8) [%s]", survival.StatusStr(state.Task8Done)),
				event: func(s *survival.Survivor) { challenges.RunJournal(reader, s) }},
			{title: fmt.Sprintf("Випробування 6. Інтерфейси (Завдання 10) [%s]", survival.StatusStr(state.Task10Done)),
				event: func(s *survival.Survivor) { challenges.RunInterfaceChallenge(s) }},
			{title: fmt.Sprintf("Випробування 7. Паніки та помилки (Завдання 11) [%s]", survival.StatusStr(state.Task11Done)),
				event: func(s *survival.Survivor) { challenges.RunErrorHandlingChallenge(s) }},
			{title: fmt.Sprintf("Випробування 8. Паралелізм (Завдання 12) [%s]", survival.StatusStr(state.Task12Done)),
				event: func(s *survival.Survivor) { challenges.RunConcurrencyChallenge(s) }},
			{title: "Інші випробування (Ім'я, Розрахунок повітря)",
				event: func(s *survival.Survivor) { handleLegacyChallenges(reader, s) }},
			{title: "Назад", back: true},
		}
	})
}

func handleLegacyChallenges(reader *bufio.Reader, state *survival.Survivor) {
	runMenu(reader, state, "СПИСОК ВИПРОБУВАНЬ", func() []menuItem {
		return []menuItem{
			{title: fmt.Sprintf("Завдання 1: Ввести ім'я [%s]", survival.StatusStr(state.Task1Done)),
				event: func(s *survival.Survivor) { enterNameTask(reader, s) }},
			{title: fmt.Sprintf("Завдання 2: Розрахунок повітря кімнати [%s]", survival.StatusStr(state.Task2Done)),
				event: roomAirTask},
			{title: "Назад", back: true},
		}
	})
}

func enterNameTask(reader *bufio.Reader, s *survival.Survivor) {
	if s.Task1Done {
		fmt.Printf("Ім'я вже встановлено: %s\n", s.PlayerName)
		return
	}
	fmt.Print("Введіть ваше ім'я: ")
	name := strings.TrimSpace(survival.ReadString(reader))
	survival.ApplyEvent(s,
		func(s *survival.Survivor) {
			s.PlayerName = name
			s.Task1Done = true
		},
		survival.AddPoints(1),
	)
	s.EnsurePlayerInJournal()
	fmt.Printf("Ім'я збережено: %s \n", s.PlayerName)
}

func roomAirTask(s *survival.Survivor) {
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
	survival.ApplyEvent(s,
		survival.AddPoints(1+2),
		survival.AddAir(hours+24.0),
		func(s *survival.Survivor) { s.Task2Done = true },
	)
	fmt.Println("Вірно! Нараховано +3 бали (Завдання 2 та 2*) і надано бонусний запас повітря на 24 години.")
}

func showScore(state *survival.Survivor) {
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
