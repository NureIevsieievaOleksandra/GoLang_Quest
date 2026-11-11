package challenges

import (
	"Quest/survival"
	"bufio"
	"fmt"
	"strings"
)

type TaskItem struct {
	condition string
	solution  string
	seenToday bool
}

var notebook []TaskItem
var backupNotebook []TaskItem

func RunNotebookChallenge(reader *bufio.Reader, state *survival.GameState) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 3: ПРОГРАМНИЙ ЗАПИСНИК ===")

NotebookLoop:
	fmt.Println("\n--- МЕНЮ ЗАПИСНИКА ЗАВДАНЬ ---")
	fmt.Println("1. Показати всі завдання (з індексами)")
	fmt.Println("2. Додати нове завдання")
	fmt.Println("3. Відзначити завдання як таке, що було сьогодні")
	fmt.Println("4. Показати лише завдання, що НЕ зустрічалися сьогодні")
	fmt.Println("5. Пошук завдання за умовою")
	fmt.Println("6. Видалити завдання за номером")
	fmt.Println("7. Створити резервну копію (Backup)")
	fmt.Println("8. Перевірити незалежність резервної копії")
	fmt.Println("9. Завершити роботу з записником (Зарахувати бали)")
	fmt.Print("Оберіть дію (1-9): ")

	choice := strings.TrimSpace(survival.ReadString(reader))

	switch choice {
	case "1":
		showTasks(notebook)

	case "2":
		fmt.Print("Введіть умову нового завдання: ")
		cond := strings.TrimSpace(survival.ReadString(reader))
		if cond == "" {
			fmt.Println("Умова не може бути порожньою.")
			break
		}

		duplicateIdx := findTaskBycondition(cond)
		if duplicateIdx != -1 {
			fmt.Printf("Це завдання вже вирішувалося раніше (№%d)!\n", duplicateIdx)
			fmt.Printf("Знайдене рішення: %s\n", notebook[duplicateIdx].solution)
			break
		}

		fmt.Print("Введіть рішення завдання: ")
		sol := strings.TrimSpace(survival.ReadString(reader))
		notebook = append(notebook, TaskItem{
			condition: cond,
			solution:  sol,
			seenToday: false,
		})
		fmt.Println("Завдання успішно додано до записника!")

	case "3":
		showTasks(notebook)
		if len(notebook) == 0 {
			break
		}
		fmt.Print("Введіть індекс завдання, що зустрілося сьогодні: ")
		idx := int(survival.ReadFloat(reader))
		if idx >= 0 && idx < len(notebook) {
			notebook[idx].seenToday = true
			fmt.Printf("Завдання [%d] відзначено як побачене сьогодні.\n", idx)
		} else {
			fmt.Println("Некоректний індекс.")
		}

	case "4":
		fmt.Println("\n--- Завдання, що НЕ зустрічалися сьогодні ---")
		count := 0
		for i, t := range notebook {
			if !t.seenToday {
				fmt.Printf("[%d] Умова: %s | Рішення: %s\n", i, t.condition, t.solution)
				count++
			}
		}
		if count == 0 {
			fmt.Println("Усі завдання вже зустрічалися сьогодні або список порожній.")
		}

	case "5":
		fmt.Print("Введіть фрагмент умови для пошуку: ")
		query := strings.TrimSpace(survival.ReadString(reader))
		found := false
		for i, t := range notebook {
			if strings.Contains(strings.ToLower(t.condition), strings.ToLower(query)) {
				status := "Сьогодні не було"
				if t.seenToday {
					status = "Сьогодні було"
				}
				fmt.Printf("[%d] %s | Рішення: %s | Статус: %s\n", i, t.condition, t.solution, status)
				found = true
			}
		}
		if !found {
			fmt.Println("Нічого не знайдено.")
		}

	case "6":
		showTasks(notebook)
		if len(notebook) == 0 {
			break
		}
		fmt.Print("Введіть номер завдання для видалення: ")
		idx := int(survival.ReadFloat(reader))
		if idx >= 0 && idx < len(notebook) {
			notebook = append(notebook[:idx], notebook[idx+1:]...)
			fmt.Println("Завдання успішно видалено.")
		} else {
			fmt.Println("Некоректний індекс.")
		}

	case "7":
		backupNotebook = make([]TaskItem, len(notebook))
		copy(backupNotebook, notebook)
		fmt.Printf("Резервну копію створено (%d елементів).\n", len(backupNotebook))

	case "8":
		if len(backupNotebook) == 0 {
			fmt.Println("Резервна копія порожня. Спочатку зробіть Backup (пункт 7).")
			break
		}
		fmt.Println("\nПеревірка ізоляції: тимчасово змінимо перше завдання в бекапі...")
		oldVal := backupNotebook[0].condition
		backupNotebook[0].condition = "[ТЕСТ] Змінена копія"

		fmt.Printf("Оригінал [0]: %s\n", notebook[0].condition)
		fmt.Printf("Бекап    [0]: %s\n", backupNotebook[0].condition)

		if notebook[0].condition != backupNotebook[0].condition {
			fmt.Println("Перевірка пройдена: копія незалежна, зміна не вплинула на оригінал!")
		} else {
			fmt.Println("Помилка: списки зв'язані посиланням.")
		}
		backupNotebook[0].condition = oldVal

	case "9":
		rewardPlayer(state)
		return

	default:
		fmt.Println("Неправильний вибір. Введіть число від 1 до 9.")
	}

	goto NotebookLoop
}

func showTasks(list []TaskItem) {
	if len(list) == 0 {
		fmt.Println("(Список завдань наразі порожній)")
		return
	}
	fmt.Println("\n--- Список завдань у книжці ---")
	for i, t := range list {
		status := "не було"
		if t.seenToday {
			status = "Сьогодні вже було"
		}
		fmt.Printf("[%d] Умова: \"%s\" | Рішення: \"%s\" | [%s]\n", i, t.condition, t.solution, status)
	}
}

func findTaskBycondition(condition string) int {
	for i, t := range notebook {
		if strings.EqualFold(strings.TrimSpace(t.condition), strings.TrimSpace(condition)) {
			return i
		}
	}
	return -1
}

func rewardPlayer(state *survival.GameState) {
	if !state.Task6Done {
		pointsReward := 10
		foodReward := 10
		waterReward := 10.0

		survival.ApplyEvent(state,
			survival.AddPoints(pointsReward),
			survival.AddSupplies(survival.DaySupplies{Food: foodReward, Water: waterReward}),
			func(s *survival.Survivor) { s.Task6Done = true },
		)

		fmt.Printf("\nОтримано: +%d балів, +%d їжі, +%.1f л води!\n",
			pointsReward, foodReward, waterReward)
	} else {
		fmt.Println("Роботу з записником завершено.")
	}
}
