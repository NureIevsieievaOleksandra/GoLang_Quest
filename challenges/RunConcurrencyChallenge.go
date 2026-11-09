package challenges

import (
	"Quest/events"
	"Quest/models"
	"fmt"
	"math/rand"
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

func RunConcurrencyChallenge(state *models.Survivor) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 12: ПАРАЛЕЛІЗМ ===")
	fmt.Println("Запуск горутин... Для успішного завершення симуляція триватиме до вичерпання одного з ресурсів.")

	eventStream := make(chan events.Event, 10)
	actionStream := make(chan Action, 10)
	done := make(chan struct{})

	var wg sync.WaitGroup

	originalFood := state.FoodMeals
	originalWater := state.WaterLiters
	originalAir := state.AirHours

	stateMu.Lock()
	state.FoodMeals = 20
	state.WaterLiters = 20.0
	state.AirHours = 20.0
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
				return
			case <-ticker.C:
				stateMu.Lock()
				fmt.Printf("\nІтерація %d:\n", iteration)
				fmt.Printf("Запас їжі: %d;\n", state.FoodMeals)
				fmt.Printf("Запас води: %.2f;\n", state.WaterLiters)
				fmt.Printf("Запас повітря: %.2f.\n\n", state.AirHours)
				stateMu.Unlock()
				iteration++
			}
		}
	}()

	<-done

	wg.Wait()

	fmt.Println("\nСимуляцію реального часу завершено. Ресурс вичерпано.")

	state.FoodMeals = originalFood
	state.WaterLiters = originalWater
	state.AirHours = originalAir

	if !state.Task12Done {
		state.Task12Done = true
		state.Points += 8
		fmt.Println("Завдання 12 (Паралелізм) зараховано. +8 балів.")
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
