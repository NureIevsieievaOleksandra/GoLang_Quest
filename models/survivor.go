package models

type Survivor = GameState

func (s *GameState) Supplies() DaySupplies {
	return DaySupplies{Food: s.FoodMeals, Water: s.WaterLiters, Air: s.AirHours}
}

func (s *GameState) SetSupplies(d DaySupplies) {
	s.FoodMeals = d.Food
	s.WaterLiters = d.Water
	s.AirHours = d.Air
}

func (s *GameState) EnsurePlayerInJournal() *Participant {
	p := s.Journal.Player()
	if p == nil {
		s.Journal.AddParticipant(Participant{Name: s.PlayerName, IsPlayer: true})
		p = s.Journal.Player()
	}
	if p.Name == "" && s.PlayerName != "" {
		p.Name = s.PlayerName
	}
	if len(p.Supplies) == 0 {
		p.Supplies = append(p.Supplies, s.Supplies())
	}
	return p
}

func (s *GameState) PushToJournal() {
	p := s.EnsurePlayerInJournal()
	p.Supplies[len(p.Supplies)-1] = s.Supplies()
}
