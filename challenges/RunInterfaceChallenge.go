package challenges

import (
	"Quest/survival"
	"fmt"
	"reflect"
)

func RunInterfaceChallenge(state *survival.Survivor) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 6: ІНТЕРФЕЙСИ ТА TYPE SWITCH (ЗАВДАННЯ 10) ===")

	if state.Task10Done {
		fmt.Println("(Ви вже пройшли це випробування, але давайте запустимо симуляцію ще раз)")
	}

	mixedList := []interface{}{
		42,
		survival.EatAction{},
		3,
		"Привіт",
		2,
		2,
		survival.DrinkAction{},
		1,
		survival.RestAction{},
		survival.AirLeakEvent{},
		5,
		"Ой, щось пішло не так...",
		survival.FoodSpoiledEvent{},
		2,
		99,
	}

	fmt.Println("\n--- Починаємо обробку списку ---")

	for i := 0; i < len(mixedList); i++ {
		item := mixedList[i]

		switch v := item.(type) {
		case string:
			fmt.Println(v)

		case survival.Action:
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

		case survival.TriggerEvent:
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
		survival.ApplyEvent(state,
			survival.AddPoints(points),
			func(s *survival.Survivor) { s.Task10Done = true },
		)
		fmt.Printf("\nНараховано +%d балів.\n", points)
	}
}
