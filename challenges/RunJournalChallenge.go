package challenges

import (
	"Quest/events"
	"Quest/models"
	"Quest/utils"
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func RunJournal(reader *bufio.Reader, state *models.GameState) {
	fmt.Println("\n=== ЖУРНАЛ ЛЮДИНИ, ЩО ВИЖИЛА (ЗАВДАННЯ 8) ===")

JournalLoop:
	state.PushToJournal()

	fmt.Println("\n--- ЖУРНАЛ: ЗАПАСИ ---")
	fmt.Println("1. Мої поточні запаси")
	fmt.Println("2. Список учасників")
	fmt.Println("3. Додати учасника")
	fmt.Println("4. Редагувати учасника")
	fmt.Println("5. Видалити учасника")
	fmt.Println("6. Керувати запасами учасника")
	if state.Task8Done {
		fmt.Println("7. Назад")
	} else {
		fmt.Println("7. Назад (зарахувати бали)")
	}
	fmt.Print("Оберіть пункт (1-7): ")

	choice := strings.TrimSpace(utils.ReadString(reader))
	switch choice {
	case "1":
		journalShowMySupplies(state)

	case "2":
		journalShowParticipants(&state.Journal)

	case "3":
		name := journalReadLine(reader, "Ім'я учасника: ")
		age, ok := journalReadInt(reader, "Вік: ")
		if !ok {
			fmt.Println("Вік має бути цілим числом.")
			break
		}
		p, err := models.NewParticipant(name, age)
		if err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		if journalAskYes(reader, "Одразу додати перший день із запасами? (т/н): ") {
			if d, ok := journalReadSupplies(reader, models.DaySupplies{}); ok {
				if err := p.AddDay(d.Food, d.Water, d.Air); err != nil {
					fmt.Println("Помилка:", err)
				}
			}
		}
		state.Journal.AddParticipant(p)
		fmt.Printf("Учасника «%s» додано до журналу.\n", p.Name)

	case "4":
		idx, ok := journalPickParticipant(reader, &state.Journal)
		if !ok {
			break
		}
		p, _ := state.Journal.Get(idx)
		name := journalReadLine(reader, fmt.Sprintf("Нове ім'я [Enter = %s]: ", p.Name))
		if name == "" {
			name = p.Name
		}
		ageStr := journalReadLine(reader, fmt.Sprintf("Новий вік [Enter = %d]: ", p.Age))
		age := p.Age
		if ageStr != "" {
			v, err := strconv.Atoi(ageStr)
			if err != nil {
				fmt.Println("Вік має бути цілим числом.")
				break
			}
			age = v
		}
		if err := state.Journal.EditParticipant(idx, name, age); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Дані учасника оновлено.")

	case "5":
		idx, ok := journalPickParticipant(reader, &state.Journal)
		if !ok {
			break
		}
		p, _ := state.Journal.Get(idx)
		name := p.DisplayName()
		if err := state.Journal.RemoveParticipant(idx); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Printf("Учасника «%s» видалено з журналу.\n", name)

	case "6":
		idx, ok := journalPickParticipant(reader, &state.Journal)
		if !ok {
			break
		}
		journalSuppliesMenu(reader, state, idx)

	case "7":
		journalReward(state)
		return

	default:
		fmt.Println("Невідома команда. Введіть число від 1 до 7.")
	}

	goto JournalLoop
}

func journalSuppliesMenu(reader *bufio.Reader, state *models.GameState, idx int) {
SuppliesLoop:
	state.PushToJournal()
	p, err := state.Journal.Get(idx)
	if err != nil {
		return
	}

	fmt.Printf("\n--- ЗАПАСИ УЧАСНИКА: %s ---\n", p.DisplayName())
	if p.IsPlayer {
		fmt.Println("(Ваш поточний день — це ваші реальні запаси в грі)")
	}
	fmt.Println("1. Показати список днів")
	fmt.Println("2. Додати новий день виживання із запасами")
	fmt.Println("3. Додати запаси до поточного дня")
	fmt.Println("4. Редагувати запаси поточного дня")
	fmt.Println("5. Видалити конкретний день")
	fmt.Println("6. Обнулити окремий ресурс поточного дня")
	fmt.Println("7. Обнулити всі запаси поточного дня")
	fmt.Println("8. Назад")
	fmt.Print("Оберіть пункт (1-8): ")

	choice := strings.TrimSpace(utils.ReadString(reader))
	changed := false

	switch choice {
	case "1":
		journalShowDays(p)

	case "2":
		def := models.DaySupplies{}
		if cur, err := p.CurrentDay(); err == nil {
			def = *cur
		}
		fmt.Println("Запаси нового дня (Enter — перенести залишок поточного дня):")
		d, ok := journalReadSupplies(reader, def)
		if !ok {
			break
		}
		if err := p.AddDay(d.Food, d.Water, d.Air); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Printf("Додано день %d.\n", len(p.Supplies))
		changed = true

	case "3":
		if _, err := p.CurrentDay(); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Скільки додати (Enter — 0):")
		d, ok := journalReadSupplies(reader, models.DaySupplies{})
		if !ok {
			break
		}
		if err := p.AddToCurrentDay(d.Food, d.Water, d.Air); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Запаси поточного дня поповнено.")
		changed = true

	case "4":
		cur, err := p.CurrentDay()
		if err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Нові значення (Enter — залишити як є):")
		d, ok := journalReadSupplies(reader, *cur)
		if !ok {
			break
		}
		if err := p.EditCurrentDay(d.Food, d.Water, d.Air); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Запаси поточного дня змінено.")
		changed = true

	case "5":
		if len(p.Supplies) == 0 {
			fmt.Println("Помилка:", models.ErrNoDays)
			break
		}
		if p.IsPlayer && len(p.Supplies) == 1 {
			fmt.Println("Це ваш єдиний день — його видалити не можна.")
			break
		}
		journalShowDays(p)
		n, ok := journalReadInt(reader, "Номер дня для видалення: ")
		if !ok {
			fmt.Println("Потрібно ввести число.")
			break
		}
		if err := p.DeleteDay(n - 1); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Printf("День %d видалено.\n", n)
		changed = true

	case "6":
		fmt.Println("Який ресурс обнулити? 1 — їжа, 2 — вода, 3 — повітря")
		res := map[string]string{"1": models.ResourceFood, "2": models.ResourceWater, "3": models.ResourceAir}
		r, ok := res[journalReadLine(reader, "Ваш вибір: ")]
		if !ok {
			fmt.Println("Невідомий ресурс.")
			break
		}
		if err := p.ResetCurrentResource(r); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Ресурс поточного дня обнулено.")
		changed = true

	case "7":
		if err := p.ResetCurrentDay(); err != nil {
			fmt.Println("Помилка:", err)
			break
		}
		fmt.Println("Усі запаси поточного дня обнулено.")
		changed = true

	case "8":
		return

	default:
		fmt.Println("Невідома команда. Введіть число від 1 до 8.")
	}

	if changed && p.IsPlayer {
		if cur, err := p.CurrentDay(); err == nil {
			events.ApplyEvent(state, events.SetSupplies(*cur))
		}
	}

	goto SuppliesLoop
}

func journalShowMySupplies(state *models.GameState) {
	fmt.Println("\n========== ПОТОЧНІ ЗАПАСИ ==========")
	fmt.Printf("• Повітря: %.1f год (~%.1f днів)\n", state.AirHours, state.AirHours/24.0)
	fmt.Printf("• Вода:    %.2f л (~%.1f днів)\n", state.WaterLiters, state.WaterLiters/2.0)
	fmt.Printf("• Їжа:     %d прийомів (~%.1f днів)\n", state.FoodMeals, float64(state.FoodMeals)/3.0)
	if p := state.Journal.Player(); p != nil {
		fmt.Printf("• День виживання за журналом: %d\n", len(p.Supplies))
	}
	fmt.Println("=====================================")
}

func journalShowParticipants(j *models.Journal) {
	if len(j.Participants) == 0 {
		fmt.Println("(У журналі ще немає учасників)")
		return
	}
	fmt.Println("\n--- Учасники ---")
	for i, p := range j.Participants {
		age := strconv.Itoa(p.Age)
		if p.IsPlayer && p.Age == 0 {
			age = "не вказано"
		}
		fmt.Printf("%d. %s | Вік: %s | Днів у журналі: %d\n", i+1, p.DisplayName(), age, len(p.Supplies))
	}
}

func journalShowDays(p *models.Participant) {
	fmt.Printf("\n--- Дні виживання: %s ---\n", p.DisplayName())
	if len(p.Supplies) == 0 {
		fmt.Println("(Днів ще немає)")
		return
	}
	for i, d := range p.Supplies {
		mark := ""
		if i == len(p.Supplies)-1 {
			mark = "  <- поточний"
		}
		fmt.Printf("День %d: %s%s\n", i+1, d, mark)
	}
}

func journalReward(state *models.GameState) {
	if state.Task8Done {
		return
	}
	events.ApplyEvent(state,
		events.AddPoints(7),
		events.AddSupplies(models.DaySupplies{Food: 2, Water: 6}),
		func(s *models.Survivor) { s.Task8Done = true },
	)
	fmt.Println("\n[Журнал] Завдання 8 зараховано: +7 балів!")
	fmt.Println("Отримано запас на два прийоми їжі (+2) та воду на 3 дні (+6 л).")
}

func journalReadLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	return strings.TrimSpace(utils.ReadString(reader))
}

func journalReadInt(reader *bufio.Reader, prompt string) (int, bool) {
	v, err := strconv.Atoi(journalReadLine(reader, prompt))
	return v, err == nil
}

func journalAskYes(reader *bufio.Reader, prompt string) bool {
	ans := strings.ToLower(journalReadLine(reader, prompt))
	return ans == "т" || ans == "так" || ans == "y" || ans == "yes"
}

func journalPickParticipant(reader *bufio.Reader, j *models.Journal) (int, bool) {
	journalShowParticipants(j)
	if len(j.Participants) == 0 {
		return 0, false
	}
	n, ok := journalReadInt(reader, "Номер учасника: ")
	if !ok || n < 1 || n > len(j.Participants) {
		fmt.Println("Некоректний номер учасника.")
		return 0, false
	}
	return n - 1, true
}

func journalReadSupplies(reader *bufio.Reader, def models.DaySupplies) (models.DaySupplies, bool) {
	res := def

	s := journalReadLine(reader, fmt.Sprintf("  Їжа, прийомів [%d]: ", def.Food))
	if s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			fmt.Println("Їжа — ціле невід'ємне число.")
			return def, false
		}
		res.Food = v
	}

	s = journalReadLine(reader, fmt.Sprintf("  Вода, л [%.2f]: ", def.Water))
	if s != "" {
		v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
		if err != nil || v < 0 {
			fmt.Println("Вода — невід'ємне число.")
			return def, false
		}
		res.Water = v
	}

	s = journalReadLine(reader, fmt.Sprintf("  Повітря, год [%.1f]: ", def.Air))
	if s != "" {
		v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
		if err != nil || v < 0 {
			fmt.Println("Повітря — невід'ємне число.")
			return def, false
		}
		res.Air = v
	}

	return res, true
}
