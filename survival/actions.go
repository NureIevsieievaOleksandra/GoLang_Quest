package survival

import "fmt"

type Action interface {
	Execute(s *Survivor)
}

type EatAction struct{}

func (a EatAction) Execute(s *Survivor) {
	ApplyEvent(s, ConsumeFood(3))
}

type DrinkAction struct{}

func (a DrinkAction) Execute(s *Survivor) {
	ApplyEvent(s, ConsumeWater(5.0))
}

type RestAction struct{}

func (a RestAction) Execute(s *Survivor) {
	fmt.Println("Я сплю. І спатиму ще 6 годин.")
}

type TriggerEvent interface {
	Trigger(s *Survivor, modifier int)
}

type AirLeakEvent struct{}

func (e AirLeakEvent) Trigger(s *Survivor, modifier int) {
	fmt.Printf("[ПОДІЯ] Витік повітря! Втрачено %d год.\n", modifier)
	ApplyEvent(s, ConsumeAir(float64(modifier)))
}

type WaterLossEvent struct{}

func (e WaterLossEvent) Trigger(s *Survivor, modifier int) {
	fmt.Printf("[ПОДІЯ] Втрата води! Вилилося %d л.\n", modifier)
	ApplyEvent(s, ConsumeWater(float64(modifier)))
}

type FoodSpoiledEvent struct{}

func (e FoodSpoiledEvent) Trigger(s *Survivor, modifier int) {
	fmt.Printf("[ПОДІЯ] Їжа зіпсувалася! Знищено %d порцій.\n", modifier)
	ApplyEvent(s, ConsumeFood(modifier))
}
