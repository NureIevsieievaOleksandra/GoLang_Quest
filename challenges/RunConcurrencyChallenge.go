package challenges

import (
	"Quest/survival"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type actionWrapper struct {
	Event survival.Event
}

func (a actionWrapper) Execute(s *survival.Survivor) {
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

func RunConcurrencyChallenge(state *survival.Survivor) {
	logMessage("\n=== ВИПРОБУВАННЯ 12 ТА 13: ПАРАЛЕЛІЗМ І ФАЙЛИ ===\n")
	logMessage("Запуск горутин... Для успішного завершення симуляція триватиме до вичерпання одного з ресурсів.\n")

	eventStream := make(chan survival.Event, 10)
	actionStream := make(chan survival.Action, 10)
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

	logMu.Lock()
	runLogs := currentRunLogs
	currentRunLogs = ""
	logMu.Unlock()
	survival.UpdateLogFile("quest_simulation.log", int(state.HoursLived), runLogs)

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

func eventWorker(id int, stream <-chan survival.Event, s *survival.Survivor, done chan struct{}, wg *sync.WaitGroup) {
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

func actionWorker(id int, stream <-chan survival.Action, s *survival.Survivor, done chan struct{}, wg *sync.WaitGroup) {
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

func safeExecuteEvent(ev survival.Event, s *survival.Survivor, done chan struct{}) {
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

func safeExecuteAction(act survival.Action, s *survival.Survivor, done chan struct{}) {
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

func checkGameOver(s *survival.Survivor, done chan struct{}) {
	if s.FoodMeals <= 0 || s.WaterLiters <= 0 || s.AirHours <= 0 {
		select {
		case <-done:
		default:
			close(done)
		}
	}
}

func generateRandomEvent() survival.Event {
	r := rand.Intn(6)
	switch r {
	case 0:
		return survival.AddFood(1)
	case 1:
		return survival.AddWater(1.0)
	case 2:
		return survival.AddAir(1.0)
	case 3:
		return func(s *survival.Survivor) { survival.AirLeakEvent{}.Trigger(s, 1) }
	case 4:
		return func(s *survival.Survivor) { survival.WaterLossEvent{}.Trigger(s, 1) }
	default:
		return func(s *survival.Survivor) { survival.FoodSpoiledEvent{}.Trigger(s, 1) }
	}
}

func generateRandomAction() survival.Action {
	r := rand.Intn(3)
	switch r {
	case 0:
		return survival.EatAction{}
	case 1:
		return survival.DrinkAction{}
	default:
		return actionWrapper{Event: survival.ConsumeAir(1.0)}
	}
}
