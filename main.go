package main

import (
	"Quest/challenges"
	"Quest/models"
	"Quest/utils"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	state := models.GameState{
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

	state.Points += 3 + 5
	state.WaterLiters += 6
	state.Task3Done = true
	state.Task3StarDone = true

MenuLoop:
	if !state.IsAlive {
		fmt.Println("\n[КІНЕЦЬ ГРИ] Ресурси вичерпано. Ви не вижили.")
		return
	}

	fmt.Println("\n--- ГОЛОВНЕ МЕНЮ ---")
	fmt.Println("1. Випробування")
	fmt.Println("2. Запаси")
	fmt.Println("3. Бали")
	fmt.Println("4. Зробити крок / прожити час (зменшити запаси)")
	fmt.Println("5. Вихід")
	fmt.Print("Оберіть пункт (1-5): ")

	choice := strings.TrimSpace(utils.ReadString(reader))

	switch choice {
	case "1":
		challengeSelectMenu(reader, &state)
	case "2":
		showSupplies(&state)
	case "3":
		showScore(&state)
	case "4":
		fmt.Print("Скільки годин минуло? (наприклад, 8 або 24): ")
		hours := utils.ReadFloat(reader)
		if hours > 0 {
			consumeResources(&state, hours)
		}
	case "5":
		fmt.Println("\nТермінал вимкнено.")
		return
	default:
		fmt.Println("Невідома команда. Будь ласка, введіть число від 1 до 5.")
	}

	goto MenuLoop
}

func consumeResources(state *models.GameState, hours float64) {
	fmt.Printf("\n--- Минуло %.1f год. Симуляція витрати запасів ---\n", hours)

	state.AirHours -= hours
	if state.AirHours < 0 {
		state.AirHours = 0
		state.IsAlive = false
		fmt.Println("Закінчилося повітря! Ви задихнулися.")
		return
	}

	waterConsumed := (2.0 / 24.0) * hours
	state.WaterLiters -= waterConsumed
	if state.WaterLiters < 0 {
		state.WaterLiters = 0
		fmt.Println("Вода закінчилася! Організм зневоднений.")
	}

	mealsConsumed := int(hours / 8.0)
	if mealsConsumed > 0 {
		state.FoodMeals -= mealsConsumed
		if state.FoodMeals < 0 {
			state.FoodMeals = 0
			fmt.Println("Їжа закінчилася! Почалося голодування.")
		}
	}

	fmt.Printf("Залишилося: Повітря: %.1f год. | Вода: %.2f л | Їжа: %d прийомів\n",
		state.AirHours, state.WaterLiters, state.FoodMeals)
}

func challengeSelectMenu(reader *bufio.Reader, state *models.GameState) {
SubMenuLoop:
	fmt.Println("\n--- ВИБІР ВИПРОБУВАННЯ ---")
	fmt.Printf("1. Випробування 1. Наближені обчислення (Завдання 4) [%s]\n", utils.StatusStr(state.Task4Done))
	fmt.Printf("2. Випробування 2. Матриця пасток (Завдання 5) [%s]\n", utils.StatusStr(state.Task5Done))
	fmt.Printf("3. Випробування 3. Записна книжка рішень (Завдання 6) [%s]\n", utils.StatusStr(state.Task6Done))
	fmt.Printf("4. Випробування 4. Озоногенератор і пароль (Завдання 7) [%s]\n", utils.StatusStr(state.Task7Done))
	fmt.Println("5. Інші випробування (Ім'я, Розрахунок повітря)")
	fmt.Println("6. Назад")
	fmt.Print("Оберіть пункт: ")

	choice := strings.TrimSpace(utils.ReadString(reader))
	switch choice {
	case "1":
		challenges.RunRoundingChallenge(reader, state)
	case "2":
		//consumeResources(state, 24.0)
		challenges.RunMatrixChallenge(state)
	case "3":
		//consumeResources(state, 24.0)
		challenges.RunNotebookChallenge(reader, state)
	case "4":
		//consumeResources(state, 24.0)
		challenges.RunCipherChallenge()
	case "5":
		handleLegacyChallenges(reader, state)
	case "6":
		return
	default:
		fmt.Println("Неправильний вибір.")
	}

	goto SubMenuLoop
}

func handleLegacyChallenges(reader *bufio.Reader, state *models.GameState) {
LegacyLoop:
	fmt.Println("\n--- СПИСОК ВИПРОБУВАНЬ ---")
	fmt.Printf("1. Завдання 1: Ввести ім'я [%s]\n", utils.StatusStr(state.Task1Done))
	fmt.Printf("2. Завдання 2: Розрахунок повітря кімнати [%s]\n", utils.StatusStr(state.Task2Done))
	fmt.Println("3. Назад")
	fmt.Print("Ваш вибір: ")

	choice := strings.TrimSpace(utils.ReadString(reader))
	switch choice {
	case "1":
		if !state.Task1Done {
			fmt.Print("Введіть ваше ім'я: ")
			state.PlayerName = strings.TrimSpace(utils.ReadString(reader))
			state.Points += 1
			state.Task1Done = true
			fmt.Printf("Ім'я збережено: %s \n", state.PlayerName)
		} else {
			fmt.Printf("Ім'я вже встановлено: %s\n", state.PlayerName)
		}

	case "2":
		volume := 3.0 * 4.0 * 2.0
		hours := volume / 1.0
		fmt.Println("\n[Розрахунок запасів повітря кімнати]")
		fmt.Println("Розміри кімнати: 3м * 4м * 2м")
		fmt.Printf("Об'єм кімнати: %.1f куб. м.\n", volume)
		fmt.Printf("Витрата: 1 куб. м/год -> Повітря у приміщенні вистачить на: %.1f год.\n", hours)

		if !state.Task2Done {
			state.Points += 1 + 2
			state.AirHours += hours + 24.0
			state.Task2Done = true
			fmt.Println("Вірно! Нараховано +3 бали (Завдання 2 та 2*) і надано бонусний запас повітря на 24 години.")
		} else {
			fmt.Println("(Ви вже отримали бали та бонус за це завдання, але можете переглядати розрахунок повторно)")
		}

	case "3":
		return
	default:
		fmt.Println("Невірний пункт.")
	}

	goto LegacyLoop
}

func showSupplies(state *models.GameState) {
	fmt.Println("\n========== ПОТОЧНІ ЗАПАСИ ==========")
	fmt.Printf("• Повітря: %.1f год (~%.1f днів)\n", state.AirHours, state.AirHours/24.0)
	fmt.Printf("• Вода:    %.2f л (~%.1f днів)\n", state.WaterLiters, state.WaterLiters/2.0)
	fmt.Printf("• Їжа:     %d прийомів (~%.1f днів)\n", state.FoodMeals, float64(state.FoodMeals)/3.0)
	fmt.Println("=====================================")
}

func showScore(state *models.GameState) {
	fmt.Println("\n========== СТАТУС БАЛІВ ==========")
	if state.PlayerName != "" {
		fmt.Printf("Гравець: %s\n", state.PlayerName)
	} else {
		fmt.Println("Гравець: [Не ідентифіковано]")
	}
	fmt.Printf("Набрано балів: %d / 60 для виходу\n", state.Points)
	fmt.Printf("Виконані етапи: Завд.1: %t | Завд.2: %t | Завд.3(інтерфейс): %t | Завд.4: %t\n",
		state.Task1Done, state.Task2Done, state.Task3Done, state.Task4Done)
	fmt.Println("==================================")
}
