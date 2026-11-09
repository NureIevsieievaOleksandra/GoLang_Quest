package challenges

import (
	"Quest/events"
	"Quest/models"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type actionWrapper struct {
	Event events.Event
}

func (a actionWrapper) Execute(s *models.Survivor) {
	if a.Event != nil {
		a.Event(s)
	}
}

var stateMu sync.Mutex
var logMu sync.Mutex
var currentRunLogs string

func logMessage(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Print(msg)

	logMu.Lock()
	currentRunLogs += msg
	logMu.Unlock()
}

func RunConcurrencyChallenge(state *models.Survivor) {
	logMessage("\n=== ВИПРОБУВАННЯ 12 ТА 13: ПАРАЛЕЛІЗМ І ФАЙЛИ ===\n")
	logMessage("Запуск горутин... Для успішного завершення симуляція триватиме до вичерпання одного з ресурсів.\n")

	eventStream := make(chan events.Event, 10)
	actionStream := make(chan Action, 10)
	done := make(chan struct{})

	var wg sync.WaitGroup

	originalFood := state.FoodMeals
	originalWater := state.WaterLiters
	originalAir := state.AirHours

	stateMu.Lock()
	state.FoodMeals = 50
	state.WaterLiters = 50.0
	state.AirHours = 50.0
	state.HoursLived = 0
	stateMu.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				eventStream <- generateRandomEvent()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				actionStream <- generateRandomAction()
			}
		}
	}()

	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go eventWorker(i, eventStream, state, done, &wg)
	}

	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go actionWorker(i, actionStream, state, done, &wg)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		iteration := 1
		for {
			select {
			case <-done:
				stateMu.Lock()
				state.HoursLived = float64(iteration)
				stateMu.Unlock()
				return
			case <-ticker.C:
				stateMu.Lock()
				logMessage("\nІтерація %d:\nЗапас їжі: %d;\nЗапас води: %.2f;\nЗапас повітря: %.2f.\n\n",
					iteration, state.FoodMeals, state.WaterLiters, state.AirHours)
				stateMu.Unlock()
				iteration++
			}
		}
	}()

	<-done

	wg.Wait()

	logMessage("\nСимуляцію реального часу завершено. Ресурс вичерпано.\n")

	updateLogFile("quest_simulation.log", int(state.HoursLived))

	state.FoodMeals = originalFood
	state.WaterLiters = originalWater
	state.AirHours = originalAir

	if !state.Task12Done {
		state.Task12Done = true
		state.Points += 8
		fmt.Println("Завдання 12 (Паралелізм) зараховано. +8 балів.")
	}
	if !state.Task13Done {
		state.Task13Done = true
		state.Points += 8
		fmt.Println("Завдання 13 (Робота з файлами) зараховано. +8 балів.")
	}
}

func eventWorker(id int, stream <-chan events.Event, s *models.Survivor, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-done:
			return
		case ev := <-stream:
			safeExecuteEvent(ev, s, done)
		}
	}
}

func actionWorker(id int, stream <-chan Action, s *models.Survivor, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-done:
			return
		case act := <-stream:
			safeExecuteAction(act, s, done)
		}
	}
}

func safeExecuteEvent(ev events.Event, s *models.Survivor, done chan struct{}) {
	stateMu.Lock()
	defer stateMu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			select {
			case <-done:
			default:
				close(done)
			}
		}
	}()

	ev(s)
	checkGameOver(s, done)
}

func safeExecuteAction(act Action, s *models.Survivor, done chan struct{}) {
	stateMu.Lock()
	defer stateMu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			select {
			case <-done:
			default:
				close(done)
			}
		}
	}()

	act.Execute(s)
	checkGameOver(s, done)
}

func checkGameOver(s *models.Survivor, done chan struct{}) {
	if s.FoodMeals <= 0 || s.WaterLiters <= 0 || s.AirHours <= 0 {
		select {
		case <-done:
		default:
			close(done)
		}
	}
}

func generateRandomEvent() events.Event {
	r := rand.Intn(6)
	switch r {
	case 0:
		return events.AddFood(1)
	case 1:
		return events.AddWater(1.0)
	case 2:
		return events.AddAir(1.0)
	case 3:
		return func(s *models.Survivor) { AirLeakEvent{}.Trigger(s, 1) }
	case 4:
		return func(s *models.Survivor) { WaterLossEvent{}.Trigger(s, 1) }
	default:
		return func(s *models.Survivor) { FoodSpoiledEvent{}.Trigger(s, 1) }
	}
}

func generateRandomAction() Action {
	r := rand.Intn(3)
	switch r {
	case 0:
		return EatAction{}
	case 1:
		return DrinkAction{}
	default:
		return actionWrapper{Event: events.ConsumeAir(1.0)}
	}
}

func updateLogFile(filename string, currentTicks int) {
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

	logMu.Lock()
	file.WriteString(currentRunLogs)
	currentRunLogs = ""
	logMu.Unlock()

	file.WriteString("\n" + oldLogs)
}
