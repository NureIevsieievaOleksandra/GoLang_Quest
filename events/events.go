package events

import (
	"Quest/models"
	"fmt"
	"math"
)

type Event func(*models.Survivor)

func ApplyEvent(s *models.Survivor, evs ...Event) {
	for _, e := range evs {
		if e == nil {
			continue
		}
		e(s)
	}
}

func Chain(evs ...Event) Event {
	return func(s *models.Survivor) {
		ApplyEvent(s, evs...)
	}
}

func Apply(a, b int, op func(int, int) int) int {
	return op(a, b)
}

func Add(a, b int) int { return a + b }
func Sub(a, b int) int { return a - b }
func Mul(a, b int) int { return a * b }

func ApplyFloat(a, b float64, op func(float64, float64) float64) float64 {
	return op(a, b)
}

func AddF(a, b float64) float64 { return a + b }
func SubF(a, b float64) float64 { return a - b }
func MulF(a, b float64) float64 { return a * b }

func ApplySupplies(cur, change models.DaySupplies, op func(float64, float64) float64) models.DaySupplies {
	return models.DaySupplies{
		Food:  int(math.Round(ApplyFloat(float64(cur.Food), float64(change.Food), op))),
		Water: ApplyFloat(cur.Water, change.Water, op),
		Air:   ApplyFloat(cur.Air, change.Air, op),
	}
}

func ModifySupplies(change models.DaySupplies, op func(float64, float64) float64) Event {
	return func(s *models.Survivor) {
		res := ApplySupplies(s.Supplies(), change, op)
		res.Food = Apply(res.Food, 0, maxInt)
		res.Water = ApplyFloat(res.Water, 0, math.Max)
		res.Air = ApplyFloat(res.Air, 0, math.Max)
		s.SetSupplies(res)
	}
}

func AddSupplies(d models.DaySupplies) Event {
	return ModifySupplies(d, AddF)
}

func ConsumeSupplies(d models.DaySupplies) Event {
	return ModifySupplies(d, SubF)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SetSupplies(d models.DaySupplies) Event {
	return func(s *models.Survivor) {
		s.SetSupplies(d)
	}
}

func MultiplySupplies(k int) Event {
	f := float64(k)
	return ModifySupplies(models.DaySupplies{Food: k, Water: f, Air: f}, MulF)
}

func KeepOnlyFood(meals int) Event {
	return func(s *models.Survivor) {
		s.FoodMeals = Apply(s.FoodMeals, meals, minInt)
	}
}

func AddAir(hours float64) Event {
	return func(s *models.Survivor) {
		s.AirHours = ApplyFloat(s.AirHours, hours, AddF)
	}
}

func AddWater(liters float64) Event {
	return func(s *models.Survivor) {
		s.WaterLiters = ApplyFloat(s.WaterLiters, liters, AddF)
	}
}

func AddFood(meals int) Event {
	return func(s *models.Survivor) {
		s.FoodMeals = Apply(s.FoodMeals, meals, Add)
	}
}

func AddPoints(points int) Event {
	return func(s *models.Survivor) {
		s.Points = Apply(s.Points, points, Add)
	}
}

func ConsumeAir(hours float64) Event {
	return func(s *models.Survivor) {
		s.AirHours = ApplyFloat(s.AirHours, hours, SubF)
		if s.AirHours < 0 {
			s.AirHours = 0
			ApplyEvent(s, LoseGame("Закінчилося повітря! Ви задихнулися."))
		}
	}
}

func ConsumeWater(liters float64) Event {
	return func(s *models.Survivor) {
		s.WaterLiters = ApplyFloat(s.WaterLiters, liters, SubF)
		if s.WaterLiters < 0 {
			s.WaterLiters = 0
			fmt.Println("Вода закінчилася! Організм зневоднений.")
		}
	}
}

func ConsumeFood(meals int) Event {
	return func(s *models.Survivor) {
		if meals <= 0 {
			return
		}
		s.FoodMeals = Apply(s.FoodMeals, meals, Sub)
		if s.FoodMeals < 0 {
			s.FoodMeals = 0
			fmt.Println("Їжа закінчилася! Почалося голодування.")
		}
	}
}

func LoseAllAir() Event {
	return func(s *models.Survivor) { s.AirHours = ApplyFloat(s.AirHours, 0, MulF) }
}

func LoseAllWater() Event {
	return func(s *models.Survivor) { s.WaterLiters = ApplyFloat(s.WaterLiters, 0, MulF) }
}

func LoseAllFood() Event {
	return func(s *models.Survivor) { s.FoodMeals = Apply(s.FoodMeals, 0, Mul) }
}

func LoseAllSupplies() Event {
	return Chain(LoseAllFood(), LoseAllWater(), LoseAllAir())
}

func LoseGame(reason string) Event {
	return func(s *models.Survivor) {
		if s.GameOver {
			return
		}
		if reason != "" {
			fmt.Println(reason)
		}
		s.IsAlive = false
		s.GameOver = true
		fmt.Println("\n[КІНЕЦЬ ГРИ] Ресурси вичерпано. Ви не вижили.")
	}
}

func WinGame() Event {
	return func(s *models.Survivor) {
		if s.GameOver {
			return
		}
		s.Won = true
		s.GameOver = true
		name := s.PlayerName
		if name == "" {
			name = "Невідомий"
		}
		fmt.Println("\n[КІНЕЦЬ ГРИ] Стіна відсувається, за нею — світло.")
		fmt.Printf("%s, ви вибралися! Набрано балів: %d.\n", name, s.Points)
	}
}

func UnlockExit() Event {
	return func(s *models.Survivor) {
		if s.ExitUnlocked {
			return
		}
		s.ExitUnlocked = true
		fmt.Println("\n*** Набрано 60 балів! Вихід відкрито. Оберіть «Вихід» у головному меню, щоб вибратися. ***")
	}
}

func CheckVictory() Event {
	return func(s *models.Survivor) {
		if !s.ExitUnlocked && s.Points >= 60 {
			ApplyEvent(s, UnlockExit())
		}
	}
}

func NewDay() Event {
	return func(s *models.Survivor) {
		s.PushToJournal()
		p := s.Journal.Player()
		cur := s.Supplies()
		_ = p.AddDay(cur.Food, cur.Water, cur.Air)
		fmt.Printf("Настав новий день (день %d у журналі).\n", len(p.Supplies))
	}
}

func PassTime(hours float64) Event {
	return func(s *models.Survivor) {
		if hours <= 0 {
			return
		}
		fmt.Printf("\n--- Минуло %.1f год. Симуляція витрати запасів ---\n", hours)

		remaining := hours
		for remaining > 0 {
			dayBefore := int(s.HoursLived / 24)
			chunk := math.Min(remaining, 24-math.Mod(s.HoursLived, 24))
			remaining = ApplyFloat(remaining, chunk, SubF)
			s.HoursLived = ApplyFloat(s.HoursLived, chunk, AddF)

			ApplyEvent(s, ConsumeAir(chunk))
			if !s.IsAlive {
				return
			}
			ApplyEvent(s,
				ConsumeWater((2.0/24.0)*chunk),
				ConsumeFood(int(chunk/8.0)),
			)
			if int(s.HoursLived/24) > dayBefore {
				ApplyEvent(s, NewDay())
			}
		}

		fmt.Printf("Залишилося: Повітря: %.1f год. | Вода: %.2f л | Їжа: %d прийомів\n",
			s.AirHours, s.WaterLiters, s.FoodMeals)
	}
}
