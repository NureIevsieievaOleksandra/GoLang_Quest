package main

import (
	"errors"
	"fmt"
)

const (
	Food  = "food"
	Water = "water"
	Air   = "air"
)

var Resources = []string{Food, Water, Air}

var resourceNames = map[string]string{
	Food:  "Їжа",
	Water: "Вода",
	Air:   "Повітря",
}

func ResourceName(r string) string {
	if n, ok := resourceNames[r]; ok {
		return n
	}
	return r
}

func ValidResource(r string) bool {
	_, ok := resourceNames[r]
	return ok
}

const (
	MaxStartResource = 100000

	MinDurationMinutes = 1
	MaxDurationMinutes = 7 * 24 * 60
)

type Balance struct {
	Food  int
	Water int
	Air   int
}

func (b Balance) Get(r string) int {
	switch r {
	case Food:
		return b.Food
	case Water:
		return b.Water
	case Air:
		return b.Air
	}
	return 0
}

var (
	ErrBadResource       = errors.New("невідомий ресурс")
	ErrSameResource      = errors.New("ресурс, який ви пропонуєте, і ресурс ціни мають відрізнятися")
	ErrBadAmount         = errors.New("кількість має бути більшою за нуль")
	ErrBadPrice          = errors.New("стартова ціна має бути більшою за нуль")
	ErrBadBuyout         = errors.New("викупна ціна має бути не меншою за стартову")
	ErrBadDuration       = fmt.Errorf("тривалість аукціону — від %d хв до %d хв", MinDurationMinutes, MaxDurationMinutes)
	ErrNotEnough         = errors.New("недостатньо ресурсів")
	ErrOwnLot            = errors.New("не можна робити ставки чи викуповувати власний лот")
	ErrLowBid            = errors.New("ставка замала")
	ErrLotClosed         = errors.New("аукціон за цим лотом уже завершено")
	ErrNoBuyout          = errors.New("для цього лота викуп не передбачено")
	ErrBuyoutUnavailable = errors.New("викуп більше недоступний: поточна ставка перевищила половину викупної ціни")
	ErrAlreadyLeading    = errors.New("ваша ставка вже найбільша")
	ErrBadStartBalance   = fmt.Errorf("запаси мають бути цілими числами від 0 до %d", MaxStartResource)
)

type LotInput struct {
	OfferResource   string
	OfferAmount     int
	PriceResource   string
	StartPrice      int
	BuyoutPrice     int
	DurationMinutes int
}

func ValidateLot(in LotInput, seller Balance) error {
	if !ValidResource(in.OfferResource) || !ValidResource(in.PriceResource) {
		return ErrBadResource
	}
	if in.OfferResource == in.PriceResource {
		return ErrSameResource
	}
	if in.OfferAmount <= 0 {
		return ErrBadAmount
	}
	if in.StartPrice <= 0 {
		return ErrBadPrice
	}
	if in.BuyoutPrice != 0 && in.BuyoutPrice < in.StartPrice {
		return ErrBadBuyout
	}
	if in.DurationMinutes < MinDurationMinutes || in.DurationMinutes > MaxDurationMinutes {
		return ErrBadDuration
	}
	if seller.Get(in.OfferResource) < in.OfferAmount {
		return ErrNotEnough
	}
	return nil
}

func ValidateStartBalance(b Balance) error {
	for _, v := range []int{b.Food, b.Water, b.Air} {
		if v < 0 || v > MaxStartResource {
			return ErrBadStartBalance
		}
	}
	return nil
}

func MinNextBid(startPrice, maxBid int, hasBids bool) int {
	if !hasBids {
		return startPrice
	}
	return maxBid + 1
}

func ValidateBid(startPrice, maxBid int, hasBids bool, amount int, bidderBalance int) error {
	if amount < MinNextBid(startPrice, maxBid, hasBids) {
		return ErrLowBid
	}
	if bidderBalance < amount {
		return ErrNotEnough
	}
	return nil
}

func BuyoutAvailable(buyoutPrice, maxBid int, hasBids bool) bool {
	if buyoutPrice <= 0 {
		return false
	}
	if !hasBids {
		return true
	}
	return maxBid*2 <= buyoutPrice
}

func ValidateBuyout(buyoutPrice, maxBid int, hasBids bool, buyerBalance int) error {
	if buyoutPrice <= 0 {
		return ErrNoBuyout
	}
	if !BuyoutAvailable(buyoutPrice, maxBid, hasBids) {
		return ErrBuyoutUnavailable
	}
	if buyerBalance < buyoutPrice {
		return ErrNotEnough
	}
	return nil
}
