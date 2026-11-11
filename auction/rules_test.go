package main

import (
	"errors"
	"testing"
)

func TestValidateLot(t *testing.T) {
	bal := Balance{Food: 10, Water: 5, Air: 0}
	ok := LotInput{OfferResource: Food, OfferAmount: 5, PriceResource: Water, StartPrice: 3, BuyoutPrice: 10, DurationMinutes: 10}

	tests := []struct {
		name string
		edit func(*LotInput)
		want error
	}{
		{"коректний лот", func(*LotInput) {}, nil},
		{"без викупу", func(in *LotInput) { in.BuyoutPrice = 0 }, nil},
		{"невідомий ресурс", func(in *LotInput) { in.OfferResource = "gold" }, ErrBadResource},
		{"однакові ресурси", func(in *LotInput) { in.PriceResource = Food }, ErrSameResource},
		{"нульова кількість", func(in *LotInput) { in.OfferAmount = 0 }, ErrBadAmount},
		{"нульова ціна", func(in *LotInput) { in.StartPrice = 0 }, ErrBadPrice},
		{"викуп менший за старт", func(in *LotInput) { in.BuyoutPrice = 2 }, ErrBadBuyout},
		{"занадто коротко", func(in *LotInput) { in.DurationMinutes = 0 }, ErrBadDuration},
		{"занадто довго", func(in *LotInput) { in.DurationMinutes = MaxDurationMinutes + 1 }, ErrBadDuration},
		{"бракує ресурсів", func(in *LotInput) { in.OfferAmount = 11 }, ErrNotEnough},
		{"рівно весь запас", func(in *LotInput) { in.OfferAmount = 10 }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ok
			tt.edit(&in)
			if got := ValidateLot(in, bal); !errors.Is(got, tt.want) {
				t.Errorf("ValidateLot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinNextBid(t *testing.T) {
	if got := MinNextBid(5, 0, false); got != 5 {
		t.Errorf("без ставок мінімум має дорівнювати стартовій ціні, отримано %d", got)
	}
	if got := MinNextBid(5, 8, true); got != 9 {
		t.Errorf("мінімум має бути на 1 більшим за поточну ставку, отримано %d", got)
	}
}

func TestValidateBid(t *testing.T) {
	tests := []struct {
		name    string
		start   int
		maxBid  int
		hasBids bool
		amount  int
		balance int
		want    error
	}{
		{"перша ставка = старт", 5, 0, false, 5, 10, nil},
		{"перша ставка нижче старту", 5, 0, false, 4, 10, ErrLowBid},
		{"перебиває поточну", 5, 7, true, 8, 10, nil},
		{"дорівнює поточній", 5, 7, true, 7, 10, ErrLowBid},
		{"не вистачає ресурсів", 5, 0, false, 6, 5, ErrNotEnough},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateBid(tt.start, tt.maxBid, tt.hasBids, tt.amount, tt.balance); !errors.Is(got, tt.want) {
				t.Errorf("ValidateBid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuyoutAvailable(t *testing.T) {
	tests := []struct {
		name    string
		buyout  int
		maxBid  int
		hasBids bool
		want    bool
	}{
		{"викуп не задано", 0, 0, false, false},
		{"ставок ще немає", 10, 0, false, true},
		{"ставка менша за половину", 10, 4, true, true},
		{"ставка рівно половина", 10, 5, true, true},
		{"ставка перевищила половину", 10, 6, true, false},
		{"непарна викупна ціна", 9, 5, true, false},
		{"непарна викупна ціна, ставка 4", 9, 4, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuyoutAvailable(tt.buyout, tt.maxBid, tt.hasBids); got != tt.want {
				t.Errorf("BuyoutAvailable(%d, %d, %v) = %v, want %v", tt.buyout, tt.maxBid, tt.hasBids, got, tt.want)
			}
		})
	}
}

func TestValidateBuyout(t *testing.T) {
	if err := ValidateBuyout(0, 0, false, 100); !errors.Is(err, ErrNoBuyout) {
		t.Errorf("очікувалась ErrNoBuyout, отримано %v", err)
	}
	if err := ValidateBuyout(10, 6, true, 100); !errors.Is(err, ErrBuyoutUnavailable) {
		t.Errorf("очікувалась ErrBuyoutUnavailable, отримано %v", err)
	}
	if err := ValidateBuyout(10, 0, false, 9); !errors.Is(err, ErrNotEnough) {
		t.Errorf("очікувалась ErrNotEnough, отримано %v", err)
	}
	if err := ValidateBuyout(10, 5, true, 10); err != nil {
		t.Errorf("викуп мав бути дозволений, отримано %v", err)
	}
}

func TestValidateRegistration(t *testing.T) {
	tests := []struct {
		name, user, email, pass, confirm string
		want                             error
	}{
		{"коректні дані", "Марія", "maria@example.com", "secret123", "secret123", nil},
		{"порожнє ім'я", " ", "maria@example.com", "secret123", "secret123", ErrBadName},
		{"погана пошта", "Марія", "maria", "secret123", "secret123", ErrBadEmail},
		{"пошта з ім'ям", "Марія", "Maria <maria@example.com>", "secret123", "secret123", ErrBadEmail},
		{"короткий пароль", "Марія", "maria@example.com", "short", "short", ErrShortPassword},
		{"паролі різні", "Марія", "maria@example.com", "secret123", "secret124", ErrPasswordMatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateRegistration(tt.user, tt.email, tt.pass, tt.confirm); !errors.Is(got, tt.want) {
				t.Errorf("ValidateRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordHash(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "secret123" {
		t.Fatal("пароль не можна зберігати відкритим текстом")
	}
	if !CheckPassword(hash, "secret123") {
		t.Error("правильний пароль не пройшов перевірку")
	}
	if CheckPassword(hash, "wrong-pass") {
		t.Error("неправильний пароль пройшов перевірку")
	}
}

func TestValidateStartBalance(t *testing.T) {
	if err := ValidateStartBalance(Balance{Food: 0, Water: 5, Air: MaxStartResource}); err != nil {
		t.Errorf("коректні запаси відхилено: %v", err)
	}
	for _, b := range []Balance{{Food: -1}, {Water: -5}, {Air: MaxStartResource + 1}} {
		if err := ValidateStartBalance(b); !errors.Is(err, ErrBadStartBalance) {
			t.Errorf("запаси %+v мали бути відхилені, отримано %v", b, err)
		}
	}
}
