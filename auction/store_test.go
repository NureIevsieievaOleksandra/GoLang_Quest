package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StartFood  = 20
	StartWater = 20
	StartAir   = 20
)

var startBalance = Balance{Food: StartFood, Water: StartWater, Air: StartAir}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL не задано — тести з базою даних пропущено")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	s := NewStore(db)
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, "TRUNCATE bids, lots, sessions, users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	return s
}

func mustUser(t *testing.T, s *Store, name string) int64 {
	t.Helper()
	id, err := s.CreateUser(context.Background(), name, name+"@example.com", "hash", startBalance)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func balance(t *testing.T, s *Store, id int64) Balance {
	t.Helper()
	u, err := s.UserByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return u.Balance
}

func expire(t *testing.T, s *Store, lotID int64) {
	t.Helper()
	if _, err := s.db.Exec(context.Background(), "UPDATE lots SET ends_at = now() - interval '1 second' WHERE id = $1", lotID); err != nil {
		t.Fatal(err)
	}
}

func mustLot(t *testing.T, s *Store, seller int64, buyout int) int64 {
	t.Helper()
	id, err := s.CreateLot(context.Background(), seller, LotInput{
		OfferResource: Food, OfferAmount: 5, PriceResource: Water, StartPrice: 3, BuyoutPrice: buyout, DurationMinutes: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestDuplicateEmail(t *testing.T) {
	s := newTestStore(t)
	mustUser(t, s, "anna")
	if _, err := s.CreateUser(context.Background(), "Інша Анна", "anna@example.com", "hash", startBalance); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("очікувалась ErrEmailTaken, отримано %v", err)
	}
}

func TestCreateLotReservesResources(t *testing.T) {
	s := newTestStore(t)
	seller := mustUser(t, s, "seller")
	mustLot(t, s, seller, 0)
	if got := balance(t, s, seller).Food; got != StartFood-5 {
		t.Fatalf("їжа продавця = %d, очікувалось %d", got, StartFood-5)
	}
	lots, err := s.ActiveLots(context.Background())
	if err != nil || len(lots) != 1 {
		t.Fatalf("очікувався 1 активний лот, отримано %d (%v)", len(lots), err)
	}
}

func TestBidRefundsPreviousLeader(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller, a, b := mustUser(t, s, "seller"), mustUser(t, s, "a"), mustUser(t, s, "b")
	lot := mustLot(t, s, seller, 0)

	if err := s.PlaceBid(ctx, lot, a, 3); err != nil {
		t.Fatal(err)
	}
	if got := balance(t, s, a).Water; got != StartWater-3 {
		t.Fatalf("ставку A не зарезервовано: вода = %d", got)
	}
	if err := s.PlaceBid(ctx, lot, b, 4); err != nil {
		t.Fatal(err)
	}
	if got := balance(t, s, a).Water; got != StartWater {
		t.Fatalf("ставку A не повернуто: вода = %d", got)
	}
	if got := balance(t, s, b).Water; got != StartWater-4 {
		t.Fatalf("ставку B не зарезервовано: вода = %d", got)
	}
}

func TestBidRules(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller, a := mustUser(t, s, "seller"), mustUser(t, s, "a")
	lot := mustLot(t, s, seller, 0)

	if err := s.PlaceBid(ctx, lot, seller, 5); !errors.Is(err, ErrOwnLot) {
		t.Errorf("ставка на власний лот: %v", err)
	}
	if err := s.PlaceBid(ctx, lot, a, 2); !errors.Is(err, ErrLowBid) {
		t.Errorf("ставка нижче старту: %v", err)
	}
	if err := s.PlaceBid(ctx, lot, a, StartWater+1); !errors.Is(err, ErrNotEnough) {
		t.Errorf("ставка більша за запас: %v", err)
	}
	if err := s.PlaceBid(ctx, lot, a, 3); err != nil {
		t.Fatal(err)
	}
	if err := s.PlaceBid(ctx, lot, a, 4); !errors.Is(err, ErrAlreadyLeading) {
		t.Errorf("повторна ставка лідера: %v", err)
	}
	expire(t, s, lot)
	if err := s.PlaceBid(ctx, lot, mustUser(t, s, "late"), 10); !errors.Is(err, ErrLotClosed) {
		t.Errorf("ставка після завершення: %v", err)
	}
}

func TestCloseExpiredWithWinner(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller, a, b := mustUser(t, s, "seller"), mustUser(t, s, "a"), mustUser(t, s, "b")
	lot := mustLot(t, s, seller, 0)
	_ = s.PlaceBid(ctx, lot, a, 3)
	_ = s.PlaceBid(ctx, lot, b, 7)

	expire(t, s, lot)
	n, err := s.CloseExpired(ctx)
	if err != nil || n != 1 {
		t.Fatalf("CloseExpired = %d, %v", n, err)
	}
	l, _ := s.Lot(ctx, lot)
	if l.Status != "sold" || l.WinnerID != b || l.FinalPrice != 7 {
		t.Fatalf("лот: статус %s, переможець %d, ціна %d", l.Status, l.WinnerID, l.FinalPrice)
	}
	if got := balance(t, s, seller); got.Water != StartWater+7 || got.Food != StartFood-5 {
		t.Fatalf("продавець: %+v", got)
	}
	if got := balance(t, s, b); got.Food != StartFood+5 || got.Water != StartWater-7 {
		t.Fatalf("переможець: %+v", got)
	}
	if got := balance(t, s, a); got.Water != StartWater || got.Food != StartFood {
		t.Fatalf("учасник, що програв: %+v", got)
	}
	if n, _ := s.CloseExpired(ctx); n != 0 {
		t.Fatal("лот не можна закрити двічі")
	}
}

func TestCloseExpiredWithoutBids(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller := mustUser(t, s, "seller")
	lot := mustLot(t, s, seller, 0)
	expire(t, s, lot)
	if _, err := s.CloseExpired(ctx); err != nil {
		t.Fatal(err)
	}
	l, _ := s.Lot(ctx, lot)
	if l.Status != "expired" {
		t.Fatalf("статус %s", l.Status)
	}
	if got := balance(t, s, seller).Food; got != StartFood {
		t.Fatalf("ресурси не повернуто продавцю: їжа = %d", got)
	}
}

func TestBuyout(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller, a, c := mustUser(t, s, "seller"), mustUser(t, s, "a"), mustUser(t, s, "c")
	lot := mustLot(t, s, seller, 10)

	if err := s.PlaceBid(ctx, lot, a, 5); err != nil {
		t.Fatal(err)
	}
	if err := s.Buyout(ctx, lot, c); err != nil {
		t.Fatalf("при ставці, що дорівнює половині, викуп має бути доступним: %v", err)
	}
	l, _ := s.Lot(ctx, lot)
	if l.Status != "bought_out" || l.WinnerID != c || l.FinalPrice != 10 {
		t.Fatalf("лот: статус %s, переможець %d, ціна %d", l.Status, l.WinnerID, l.FinalPrice)
	}
	if got := balance(t, s, a).Water; got != StartWater {
		t.Fatalf("ставку A не повернуто: %d", got)
	}
	if got := balance(t, s, c); got.Water != StartWater-10 || got.Food != StartFood+5 {
		t.Fatalf("покупець: %+v", got)
	}
	if got := balance(t, s, seller).Water; got != StartWater+10 {
		t.Fatalf("продавець не отримав оплату: %d", got)
	}
	if err := s.PlaceBid(ctx, lot, mustUser(t, s, "late"), 11); !errors.Is(err, ErrLotClosed) {
		t.Fatalf("після викупу ставки неможливі: %v", err)
	}
}

func TestBuyoutUnavailableAfterHalf(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller, a, c := mustUser(t, s, "seller"), mustUser(t, s, "a"), mustUser(t, s, "c")
	lot := mustLot(t, s, seller, 10)
	_ = s.PlaceBid(ctx, lot, a, 6)
	if err := s.Buyout(ctx, lot, c); !errors.Is(err, ErrBuyoutUnavailable) {
		t.Fatalf("очікувалась ErrBuyoutUnavailable, отримано %v", err)
	}
	if got := balance(t, s, c).Water; got != StartWater {
		t.Fatalf("при відмові ресурси не мають змінюватися: %d", got)
	}
}

func TestLeaderCanBuyout(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller, a := mustUser(t, s, "seller"), mustUser(t, s, "a")
	lot := mustLot(t, s, seller, 10)
	_ = s.PlaceBid(ctx, lot, a, 4)
	if err := s.Buyout(ctx, lot, a); err != nil {
		t.Fatal(err)
	}
	if got := balance(t, s, a); got.Water != StartWater-10 || got.Food != StartFood+5 {
		t.Fatalf("лідер після викупу: %+v", got)
	}
}

func TestConcurrentBidsKeepResourcesConsistent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seller := mustUser(t, s, "seller")
	lot := mustLot(t, s, seller, 0)

	var users []int64
	for i := 0; i < 10; i++ {
		users = append(users, mustUser(t, s, fmt.Sprintf("u%d", i)))
	}
	var wg sync.WaitGroup
	for i, u := range users {
		wg.Add(1)
		go func(u int64, amount int) {
			defer wg.Done()
			_ = s.PlaceBid(ctx, lot, u, amount)
		}(u, 3+i)
	}
	wg.Wait()

	l, _ := s.Lot(ctx, lot)
	reserved := 0
	for _, u := range users {
		reserved += StartWater - balance(t, s, u).Water
	}
	if reserved != l.MaxBid {
		t.Fatalf("зарезервовано %d води, а найбільша ставка %d", reserved, l.MaxBid)
	}
}
