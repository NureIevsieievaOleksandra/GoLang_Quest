package challenges

import (
	"Quest/models"
	"fmt"
	"os"
	"time"
)

var errorCounter int = 0 // Глобальний лічильник помилок

func ExecuteSafeActions(actions []Action, s *models.Survivor) {
	initialFood := s.FoodMeals
	initialWater := s.WaterLiters
	initialAir := s.AirHours

	success := runWithRecovery(actions, s)

	if !success {
		fmt.Println("[СИСТЕМА] Відновлення після критичного збою... Перезапуск дій з резервними запасами.")
		s.FoodMeals = initialFood + 2
		s.WaterLiters = initialWater + 2.0
		s.AirHours = initialAir + 2.0

		ExecuteSafeActions(actions, s)
	}
}

func runWithRecovery(actions []Action, s *models.Survivor) (success bool) {
	defer func() {
		if r := recover(); r != nil {
			errorCounter += 3
			logMsg := fmt.Sprintf("[%s] PANIC INTERCEPTED: %v | Лічильник: %d\n", time.Now().Format(time.RFC3339), r, errorCounter)
			fmt.Print(logMsg)
			logErrorToFile(logMsg)

			checkFatalErrors()
			success = false
		}
	}()

	for i, act := range actions {
		act.Execute(s)

		if s.FoodMeals == 1 || s.WaterLiters == 1.0 || s.AirHours == 1.0 {
			err := fmt.Errorf("критично низький рівень ресурсів (1 одиниця) на кроці %d", i+1)
			errorCounter += 1

			logMsg := fmt.Sprintf("[%s] ERROR: %v | Лічильник: %d\n", time.Now().Format(time.RFC3339), err, errorCounter)
			fmt.Print(logMsg)
			logErrorToFile(logMsg)

			checkFatalErrors()
		}
	}
	return true
}

func checkFatalErrors() {
	if errorCounter > 5 {
		msg := "ФАТАЛЬНО: Кількість помилок перевищила 5! Завершення програми."
		fmt.Println("\n[КРИТИЧНО]", msg)
		logErrorToFile(msg + "\n")
		os.Exit(1)
	}
}

func logErrorToFile(message string) {
	file, err := os.OpenFile("error_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Не вдалося відкрити файл логів:", err)
		return
	}
	defer file.Close()
	file.WriteString(message)
}
