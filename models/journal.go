package models

import (
	"errors"
	"fmt"
	"strings"
)

type DaySupplies struct {
	Food  int
	Water float64
	Air   float64
}

const (
	ResourceFood  = "food"
	ResourceWater = "water"
	ResourceAir   = "air"
)

func (d *DaySupplies) Add(food int, water, air float64) {
	d.Food += food
	d.Water += water
	d.Air += air
}

func (d *DaySupplies) Edit(food int, water, air float64) {
	d.Food = food
	d.Water = water
	d.Air = air
}

func (d *DaySupplies) Clear() {
	d.Food = 0
	d.Water = 0
	d.Air = 0
}

func (d *DaySupplies) ResetResource(resource string) error {
	switch resource {
	case ResourceFood:
		d.Food = 0
	case ResourceWater:
		d.Water = 0
	case ResourceAir:
		d.Air = 0
	default:
		return fmt.Errorf("невідомий ресурс: %s", resource)
	}
	return nil
}

func (d DaySupplies) String() string {
	return fmt.Sprintf("Їжа: %d прийомів | Вода: %.2f л | Повітря: %.1f год", d.Food, d.Water, d.Air)
}

type Participant struct {
	Name     string
	Age      int
	Supplies []DaySupplies
	IsPlayer bool
}

var (
	ErrNoDays       = errors.New("у учасника ще немає жодного дня виживання")
	ErrBadIndex     = errors.New("некоректний номер")
	ErrEmptyName    = errors.New("ім'я не може бути порожнім")
	ErrBadAge       = errors.New("вік має бути від 0 до 150")
	ErrNegative     = errors.New("запаси не можуть бути від'ємними")
	ErrPlayerDelete = errors.New("себе видалити з журналу не можна")
)

func validateInfo(name string, age int) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyName
	}
	if age < 0 || age > 150 {
		return ErrBadAge
	}
	return nil
}

func validateSupplies(food int, water, air float64) error {
	if food < 0 || water < 0 || air < 0 {
		return ErrNegative
	}
	return nil
}

func NewParticipant(name string, age int) (Participant, error) {
	if err := validateInfo(name, age); err != nil {
		return Participant{}, err
	}
	return Participant{Name: strings.TrimSpace(name), Age: age}, nil
}

func (p *Participant) EditInfo(name string, age int) error {
	if err := validateInfo(name, age); err != nil {
		return err
	}
	p.Name = strings.TrimSpace(name)
	p.Age = age
	return nil
}

func (p *Participant) AddDay(food int, water, air float64) error {
	if err := validateSupplies(food, water, air); err != nil {
		return err
	}
	p.Supplies = append(p.Supplies, DaySupplies{Food: food, Water: water, Air: air})
	return nil
}

func (p *Participant) CurrentDay() (*DaySupplies, error) {
	if len(p.Supplies) == 0 {
		return nil, ErrNoDays
	}
	return &p.Supplies[len(p.Supplies)-1], nil
}

func (p *Participant) AddToCurrentDay(food int, water, air float64) error {
	if err := validateSupplies(food, water, air); err != nil {
		return err
	}
	day, err := p.CurrentDay()
	if err != nil {
		return err
	}
	day.Add(food, water, air)
	return nil
}

func (p *Participant) EditCurrentDay(food int, water, air float64) error {
	if err := validateSupplies(food, water, air); err != nil {
		return err
	}
	day, err := p.CurrentDay()
	if err != nil {
		return err
	}
	day.Edit(food, water, air)
	return nil
}

func (p *Participant) DeleteDay(index int) error {
	if index < 0 || index >= len(p.Supplies) {
		return ErrBadIndex
	}
	p.Supplies = append(p.Supplies[:index], p.Supplies[index+1:]...)
	return nil
}

func (p *Participant) ResetCurrentResource(resource string) error {
	day, err := p.CurrentDay()
	if err != nil {
		return err
	}
	return day.ResetResource(resource)
}

func (p *Participant) ResetCurrentDay() error {
	day, err := p.CurrentDay()
	if err != nil {
		return err
	}
	day.Clear()
	return nil
}

func (p Participant) DisplayName() string {
	name := p.Name
	if name == "" {
		name = "Гравець"
	}
	if p.IsPlayer {
		return name + " (ви)"
	}
	return name
}

type Journal struct {
	Participants []Participant
}

func (j *Journal) AddParticipant(p Participant) {
	j.Participants = append(j.Participants, p)
}

func (j *Journal) Get(index int) (*Participant, error) {
	if index < 0 || index >= len(j.Participants) {
		return nil, ErrBadIndex
	}
	return &j.Participants[index], nil
}

func (j *Journal) EditParticipant(index int, name string, age int) error {
	p, err := j.Get(index)
	if err != nil {
		return err
	}
	return p.EditInfo(name, age)
}

func (j *Journal) RemoveParticipant(index int) error {
	p, err := j.Get(index)
	if err != nil {
		return err
	}
	if p.IsPlayer {
		return ErrPlayerDelete
	}
	j.Participants = append(j.Participants[:index], j.Participants[index+1:]...)
	return nil
}

func (j *Journal) Player() *Participant {
	for i := range j.Participants {
		if j.Participants[i].IsPlayer {
			return &j.Participants[i]
		}
	}
	return nil
}
