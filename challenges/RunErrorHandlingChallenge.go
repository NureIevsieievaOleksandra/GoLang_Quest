package challenges

import (
	"Quest/models"
	"fmt"
	"os"
	"time"
)

type RiskyAction struct {
	FoodDelta  int
	WaterDelta float64
	AirDelta   float64
}

func (a RiskyAction) Execute(s *models.Survivor) {
	s.FoodMeals += a.FoodDelta
	s.WaterLiters += a.WaterDelta
	s.AirHours += a.AirDelta
}

func RunErrorHandlingChallenge(state *models.Survivor) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 7: ПАНІКИ ТА ПОМИЛКИ (ЗАВДАННЯ 11) ===")

	initialFood := state.FoodMeals
	initialWater := state.WaterLiters
	initialAir := state.AirHours

	actions := []Action{
		RiskyAction{FoodDelta: -state.FoodMeals + 1, WaterDelta: 0, AirDelta: 0},     // Залишить 1 їжу
		RiskyAction{FoodDelta: 0, WaterDelta: -state.WaterLiters + 1.0, AirDelta: 0}, // Залишить 1 воду
		RiskyAction{FoodDelta: -5, WaterDelta: 0, AirDelta: 0},                       // Опустить їжу нижче 0
		EatAction{},
	}

	errorCount := 0

	for {
		if errorCount > 5 {
			fmt.Println("[КРИТИЧНО] Кількість помилок перевищила 5! Завершення програми.")
			logErrorToFile("ФАТАЛЬНО: Перевищено ліміт помилок. Програму зупинено.")
			os.Exit(1)
		}

		success := executeWithRecovery(actions, state, &errorCount)
		if success {
			fmt.Println("Всі дії виконано успішно.")
			break
		} else {
			fmt.Println("[СИСТЕМА] Відновлення після паніки... Перезапуск списку дій.")
			state.FoodMeals = initialFood + 10
			state.WaterLiters = initialWater + 10.0
			state.AirHours = initialAir + 10.0
		}
	}

	if !state.Task11Done {
		state.Task11Done = true
		state.Points += 7
		fmt.Println("\nВипробування пройдено! Нараховано +7 балів.")
	}
}

func executeWithRecovery(actions []Action, s *models.Survivor, errCount *int) (success bool) {
	defer func() {
		if r := recover(); r != nil {
			*errCount += 3
			logMsg := fmt.Sprintf("[%s] PANIC INTERCEPTED: %v | Лічильник помилок: %d\n", time.Now().Format(time.RFC3339), r, *errCount)
			fmt.Print(logMsg)
			logErrorToFile(logMsg)
			success = false
		}
	}()

	for i, act := range actions {
		act.Execute(s)
		fmt.Printf("Дія %d виконана. Стан: Їжа=%d, Вода=%.1f, Повітря=%.1f\n", i+1, s.FoodMeals, s.WaterLiters, s.AirHours)

		if s.FoodMeals < 0 || s.WaterLiters < 0 || s.AirHours < 0 {
			panic(fmt.Sprintf("Ресурси впали нижче нуля! Їжа: %d, Вода: %.2f, Повітря: %.2f", s.FoodMeals, s.WaterLiters, s.AirHours))
		}

		if s.FoodMeals == 1 || s.WaterLiters == 1.0 || s.AirHours == 1.0 {
			err := fmt.Errorf("критично низький рівень ресурсів (1 одиниця)")
			*errCount += 1
			logMsg := fmt.Sprintf("[%s] ERROR: %v | Лічильник помилок: %d\n", time.Now().Format(time.RFC3339), err, *errCount)
			fmt.Print(logMsg)
			logErrorToFile(logMsg)

			if *errCount > 5 {
				return true
			}
		}
	}
	return true
}
