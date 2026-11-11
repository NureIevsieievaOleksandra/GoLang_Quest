package survival

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func UpdateLogFile(filename string, currentTicks int, runLogs string) {
	var totalRuns, minTicks, maxTicks, totalTicksSum int

	minTicks = currentTicks
	maxTicks = currentTicks

	content, err := os.ReadFile(filename)
	oldLogs := ""

	if err == nil {
		lines := strings.Split(string(content), "\n")
		if len(lines) >= 4 && strings.HasPrefix(lines[0], "--- СТАТИСТИКА") {
			fmt.Sscanf(lines[1], "Всього запусків: %d", &totalRuns)
			fmt.Sscanf(lines[2], "Найкоротша гра: %d", &minTicks)
			fmt.Sscanf(lines[3], "Найдовша гра: %d", &maxTicks)
			fmt.Sscanf(lines[4], "Середня кількість тактів: %d", &totalTicksSum)

			totalTicksSum = totalTicksSum * totalRuns

			if len(lines) > 6 {
				oldLogs = strings.Join(lines[6:], "\n")
			}
		}
	}

	totalRuns++
	totalTicksSum += currentTicks
	if currentTicks < minTicks || minTicks == 0 {
		minTicks = currentTicks
	}
	if currentTicks > maxTicks {
		maxTicks = currentTicks
	}
	avgTicks := totalTicksSum / totalRuns

	statsHeader := fmt.Sprintf("--- СТАТИСТИКА ---\nВсього запусків: %d\nНайкоротша гра: %d\nНайдовша гра: %d\nСередня кількість тактів: %d\n------------------\n\n",
		totalRuns, minTicks, maxTicks, avgTicks)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Помилка створення файлу логу:", err)
		return
	}
	defer file.Close()

	file.WriteString(statsHeader)
	file.WriteString(fmt.Sprintf("=== ЗАПУСК %s ===\n", time.Now().Format("2006-01-02 15:04:05")))

	file.WriteString(runLogs)

	file.WriteString("\n" + oldLogs)
}
