package challenges

import (
	"Quest/events"
	"Quest/models"
	"fmt"
	"reflect"
)

type Action interface {
	Execute(s *models.Survivor)
}

type EatAction struct{}

func (a EatAction) Execute(s *models.Survivor) {
	events.ApplyEvent(s, events.ConsumeFood(3)) // Витрачаємо 3 їжі
}

type DrinkAction struct{}

func (a DrinkAction) Execute(s *models.Survivor) {
	events.ApplyEvent(s, events.ConsumeWater(5.0)) // Витрачаємо 5 води
}

type RestAction struct{}

func (a RestAction) Execute(s *models.Survivor) {
	fmt.Println("Я сплю. І спатиму ще 6 годин.")
}

type TriggerEvent interface {
	Trigger(s *models.Survivor, modifier int)
}

type AirLeakEvent struct{}

func (e AirLeakEvent) Trigger(s *models.Survivor, modifier int) {
	fmt.Printf("[ПОДІЯ] Витік повітря! Втрачено %d год.\n", modifier)
	events.ApplyEvent(s, events.ConsumeAir(float64(modifier)))
}

type WaterLossEvent struct{}

func (e WaterLossEvent) Trigger(s *models.Survivor, modifier int) {
	fmt.Printf("[ПОДІЯ] Втрата води! Вилилося %d л.\n", modifier)
	events.ApplyEvent(s, events.ConsumeWater(float64(modifier)))
}

type FoodSpoiledEvent struct{}

func (e FoodSpoiledEvent) Trigger(s *models.Survivor, modifier int) {
	fmt.Printf("[ПОДІЯ] Їжа зіпсувалася! Знищено %d порцій.\n", modifier)
	events.ApplyEvent(s, events.ConsumeFood(modifier))
}

func RunInterfaceChallenge(state *models.Survivor) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 6: ІНТЕРФЕЙСИ ТА TYPE SWITCH (ЗАВДАННЯ 10) ===")

	if state.Task10Done {
		fmt.Println("(Ви вже пройшли це випробування, але давайте запустимо симуляцію ще раз)")
	}

	mixedList := []interface{}{
		42,
		EatAction{},
		3,
		"Привіт",
		2,
		2,
		DrinkAction{},
		1,
		RestAction{},
		AirLeakEvent{},
		5,
		"Ой, щось пішло не так...",
		FoodSpoiledEvent{},
		2,
		99,
	}

	fmt.Println("\n--- Починаємо обробку списку ---")

	for i := 0; i < len(mixedList); i++ {
		item := mixedList[i]

		switch v := item.(type) {
		case string:
			fmt.Println(v)

		case Action:
			repeats := 1
			if i+1 < len(mixedList) {
				if num, ok := mixedList[i+1].(int); ok {
					repeats = num
					i++
				}
			}
			for k := 0; k < repeats; k++ {
				v.Execute(state)
				fmt.Printf("-> Стан після %s (виконання %d/%d): Їжа=%d, Вода=%.1f, Повітря=%.1f\n",
					reflect.TypeOf(v).Name(), k+1, repeats, state.FoodMeals, state.WaterLiters, state.AirHours)
			}

		case TriggerEvent:
			modifier := 1
			if i+1 < len(mixedList) {
				if num, ok := mixedList[i+1].(int); ok {
					modifier = num
					i++
				}
			}
			v.Trigger(state, modifier)
			fmt.Printf("-> Стан після події %s: Їжа=%d, Вода=%.1f, Повітря=%.1f\n",
				reflect.TypeOf(v).Name(), state.FoodMeals, state.WaterLiters, state.AirHours)

		case int:
			continue

		default:
			fmt.Printf("Невідомий тип у списку: %T\n", v)
		}
	}

	fmt.Println("--------------------------------")

	if !state.Task10Done {
		points := 8
		events.ApplyEvent(state,
			events.AddPoints(points),
			func(s *models.Survivor) { s.Task10Done = true },
		)
		fmt.Printf("\nНараховано +%d балів.\n", points)
	}
}
