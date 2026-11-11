package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            BIGSERIAL PRIMARY KEY,
	name          TEXT NOT NULL,
	email         TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	food          INTEGER NOT NULL CHECK (food >= 0),
	water         INTEGER NOT NULL CHECK (water >= 0),
	air           INTEGER NOT NULL CHECK (air >= 0),
	created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
	token      TEXT PRIMARY KEY,
	user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS lots (
	id             BIGSERIAL PRIMARY KEY,
	seller_id      BIGINT NOT NULL REFERENCES users(id),
	offer_resource TEXT NOT NULL CHECK (offer_resource IN ('food', 'water', 'air')),
	offer_amount   INTEGER NOT NULL CHECK (offer_amount > 0),
	price_resource TEXT NOT NULL CHECK (price_resource IN ('food', 'water', 'air')),
	start_price    INTEGER NOT NULL CHECK (start_price > 0),
	buyout_price   INTEGER NOT NULL DEFAULT 0 CHECK (buyout_price >= 0),
	ends_at        TIMESTAMPTZ NOT NULL,
	status         TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'sold', 'bought_out', 'expired')),
	winner_id      BIGINT REFERENCES users(id),
	final_price    INTEGER,
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
	CHECK (offer_resource <> price_resource)
);

CREATE TABLE IF NOT EXISTS bids (
	id         BIGSERIAL PRIMARY KEY,
	lot_id     BIGINT NOT NULL REFERENCES lots(id) ON DELETE CASCADE,
	bidder_id  BIGINT NOT NULL REFERENCES users(id),
	amount     INTEGER NOT NULL CHECK (amount > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS bids_lot_idx ON bids (lot_id, amount DESC);
CREATE INDEX IF NOT EXISTS lots_status_idx ON lots (status, ends_at);
`

var (
	ErrEmailTaken = errors.New("користувач із такою поштою вже зареєстрований")
	ErrNotFound   = errors.New("не знайдено")
)

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Balance      Balance
}

type Lot struct {
	ID            int64
	SellerID      int64
	SellerName    string
	OfferResource string
	OfferAmount   int
	PriceResource string
	StartPrice    int
	BuyoutPrice   int
	EndsAt        time.Time
	Status        string
	WinnerID      int64
	WinnerName    string
	FinalPrice    int
	MaxBid        int
	HasBids       bool
	BidCount      int
	LeaderID      int64
}

type Bid struct {
	ID         int64
	BidderID   int64
	BidderName string
	Amount     int
	CreatedAt  time.Time
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, schema)
	return err
}

func column(r string) (string, error) {
	switch r {
	case Food:
		return "food", nil
	case Water:
		return "water", nil
	case Air:
		return "air", nil
	}
	return "", ErrBadResource
}

func addResource(ctx context.Context, tx pgx.Tx, userID int64, r string, delta int) error {
	col, err := column(r)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, fmt.Sprintf("UPDATE users SET %s = %s + $1 WHERE id = $2", col, col), delta, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func lockBalance(ctx context.Context, tx pgx.Tx, userID int64) (Balance, error) {
	var b Balance
	err := tx.QueryRow(ctx, "SELECT food, water, air FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&b.Food, &b.Water, &b.Air)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

func (s *Store) CreateUser(ctx context.Context, name, email, passwordHash string, b Balance) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, food, water, air)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		name, email, passwordHash, b.Food, b.Water, b.Air).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrEmailTaken
	}
	return id, err
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Balance.Food, &u.Balance.Water, &u.Balance.Air)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (*User, error) {
	return scanUser(s.db.QueryRow(ctx,
		"SELECT id, name, email, password_hash, food, water, air FROM users WHERE email = $1", email))
}

func (s *Store) UserByID(ctx context.Context, id int64) (*User, error) {
	return scanUser(s.db.QueryRow(ctx,
		"SELECT id, name, email, password_hash, food, water, air FROM users WHERE id = $1", id))
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Store) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx, "INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)",
		token, userID, time.Now().Add(ttl))
	return token, err
}

func (s *Store) UserBySession(ctx context.Context, token string) (*User, error) {
	return scanUser(s.db.QueryRow(ctx,
		`SELECT u.id, u.name, u.email, u.password_hash, u.food, u.water, u.air
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token = $1 AND s.expires_at > now()`, token))
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM sessions WHERE token = $1", token)
	return err
}

func (s *Store) CreateLot(ctx context.Context, sellerID int64, in LotInput) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	bal, err := lockBalance(ctx, tx, sellerID)
	if err != nil {
		return 0, err
	}
	if err := ValidateLot(in, bal); err != nil {
		return 0, err
	}
	if err := addResource(ctx, tx, sellerID, in.OfferResource, -in.OfferAmount); err != nil {
		return 0, err
	}

	var id int64
	err = tx.QueryRow(ctx,
		`INSERT INTO lots (seller_id, offer_resource, offer_amount, price_resource, start_price, buyout_price, ends_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		sellerID, in.OfferResource, in.OfferAmount, in.PriceResource, in.StartPrice, in.BuyoutPrice,
		time.Now().Add(time.Duration(in.DurationMinutes)*time.Minute)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

const lotSelect = `
SELECT l.id, l.seller_id, seller.name, l.offer_resource, l.offer_amount, l.price_resource,
       l.start_price, l.buyout_price, l.ends_at, l.status,
       COALESCE(l.winner_id, 0), COALESCE(winner.name, ''), COALESCE(l.final_price, 0),
       COALESCE(top.amount, 0), top.amount IS NOT NULL, COALESCE(top.bidder_id, 0),
       (SELECT count(*) FROM bids b WHERE b.lot_id = l.id)
FROM lots l
JOIN users seller ON seller.id = l.seller_id
LEFT JOIN users winner ON winner.id = l.winner_id
LEFT JOIN LATERAL (
	SELECT b.amount, b.bidder_id FROM bids b WHERE b.lot_id = l.id ORDER BY b.amount DESC, b.id DESC LIMIT 1
) top ON true
`

func scanLot(row pgx.Row) (*Lot, error) {
	var l Lot
	err := row.Scan(&l.ID, &l.SellerID, &l.SellerName, &l.OfferResource, &l.OfferAmount, &l.PriceResource,
		&l.StartPrice, &l.BuyoutPrice, &l.EndsAt, &l.Status,
		&l.WinnerID, &l.WinnerName, &l.FinalPrice,
		&l.MaxBid, &l.HasBids, &l.LeaderID, &l.BidCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *Store) queryLots(ctx context.Context, query string, args ...any) ([]Lot, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lots []Lot
	for rows.Next() {
		l, err := scanLot(rows)
		if err != nil {
			return nil, err
		}
		lots = append(lots, *l)
	}
	return lots, rows.Err()
}

func (s *Store) ActiveLots(ctx context.Context) ([]Lot, error) {
	return s.queryLots(ctx, lotSelect+" WHERE l.status = 'active' ORDER BY l.ends_at")
}

func (s *Store) FinishedLots(ctx context.Context, limit int) ([]Lot, error) {
	return s.queryLots(ctx, lotSelect+" WHERE l.status <> 'active' ORDER BY l.ends_at DESC LIMIT $1", limit)
}

func (s *Store) UserLots(ctx context.Context, userID int64) ([]Lot, error) {
	return s.queryLots(ctx, lotSelect+" WHERE l.seller_id = $1 ORDER BY l.created_at DESC", userID)
}

func (s *Store) LotsWithUserBids(ctx context.Context, userID int64) ([]Lot, error) {
	return s.queryLots(ctx, lotSelect+
		" WHERE EXISTS (SELECT 1 FROM bids b WHERE b.lot_id = l.id AND b.bidder_id = $1) ORDER BY l.ends_at DESC", userID)
}

func (s *Store) Lot(ctx context.Context, id int64) (*Lot, error) {
	return scanLot(s.db.QueryRow(ctx, lotSelect+" WHERE l.id = $1", id))
}

func (s *Store) Bids(ctx context.Context, lotID int64) ([]Bid, error) {
	rows, err := s.db.Query(ctx,
		`SELECT b.id, b.bidder_id, u.name, b.amount, b.created_at
		 FROM bids b JOIN users u ON u.id = b.bidder_id
		 WHERE b.lot_id = $1 ORDER BY b.amount DESC, b.id DESC`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bids []Bid
	for rows.Next() {
		var b Bid
		if err := rows.Scan(&b.ID, &b.BidderID, &b.BidderName, &b.Amount, &b.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, rows.Err()
}

type lockedLot struct {
	SellerID      int64
	OfferResource string
	OfferAmount   int
	PriceResource string
	StartPrice    int
	BuyoutPrice   int
	EndsAt        time.Time
	Status        string
	MaxBid        int
	HasBids       bool
	LeaderID      int64
}

func lockLot(ctx context.Context, tx pgx.Tx, lotID int64) (*lockedLot, error) {
	var l lockedLot
	err := tx.QueryRow(ctx,
		`SELECT seller_id, offer_resource, offer_amount, price_resource, start_price, buyout_price, ends_at, status
		 FROM lots WHERE id = $1 FOR UPDATE`, lotID).
		Scan(&l.SellerID, &l.OfferResource, &l.OfferAmount, &l.PriceResource, &l.StartPrice, &l.BuyoutPrice, &l.EndsAt, &l.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx,
		"SELECT amount, bidder_id FROM bids WHERE lot_id = $1 ORDER BY amount DESC, id DESC LIMIT 1", lotID).
		Scan(&l.MaxBid, &l.LeaderID)
	switch {
	case err == nil:
		l.HasBids = true
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return nil, err
	}
	return &l, nil
}

func (l *lockedLot) open(now time.Time) bool {
	return l.Status == "active" && now.Before(l.EndsAt)
}

func (s *Store) PlaceBid(ctx context.Context, lotID, bidderID int64, amount int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	lot, err := lockLot(ctx, tx, lotID)
	if err != nil {
		return err
	}
	if !lot.open(time.Now()) {
		return ErrLotClosed
	}
	if lot.SellerID == bidderID {
		return ErrOwnLot
	}
	if lot.HasBids && lot.LeaderID == bidderID {
		return ErrAlreadyLeading
	}
	bal, err := lockBalance(ctx, tx, bidderID)
	if err != nil {
		return err
	}
	if err := ValidateBid(lot.StartPrice, lot.MaxBid, lot.HasBids, amount, bal.Get(lot.PriceResource)); err != nil {
		return err
	}
	if lot.HasBids {
		if err := addResource(ctx, tx, lot.LeaderID, lot.PriceResource, lot.MaxBid); err != nil {
			return err
		}
	}
	if err := addResource(ctx, tx, bidderID, lot.PriceResource, -amount); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "INSERT INTO bids (lot_id, bidder_id, amount) VALUES ($1, $2, $3)", lotID, bidderID, amount); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Buyout(ctx context.Context, lotID, buyerID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	lot, err := lockLot(ctx, tx, lotID)
	if err != nil {
		return err
	}
	if !lot.open(time.Now()) {
		return ErrLotClosed
	}
	if lot.SellerID == buyerID {
		return ErrOwnLot
	}
	bal, err := lockBalance(ctx, tx, buyerID)
	if err != nil {
		return err
	}
	available := bal.Get(lot.PriceResource)
	if lot.HasBids && lot.LeaderID == buyerID {
		available += lot.MaxBid
	}
	if err := ValidateBuyout(lot.BuyoutPrice, lot.MaxBid, lot.HasBids, available); err != nil {
		return err
	}
	if lot.HasBids {
		if err := addResource(ctx, tx, lot.LeaderID, lot.PriceResource, lot.MaxBid); err != nil {
			return err
		}
	}
	if err := addResource(ctx, tx, buyerID, lot.PriceResource, -lot.BuyoutPrice); err != nil {
		return err
	}
	if err := addResource(ctx, tx, lot.SellerID, lot.PriceResource, lot.BuyoutPrice); err != nil {
		return err
	}
	if err := addResource(ctx, tx, buyerID, lot.OfferResource, lot.OfferAmount); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		"UPDATE lots SET status = 'bought_out', winner_id = $2, final_price = $3, ends_at = now() WHERE id = $1",
		lotID, buyerID, lot.BuyoutPrice); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CloseExpired(ctx context.Context) (int, error) {
	rows, err := s.db.Query(ctx, "SELECT id FROM lots WHERE status = 'active' AND ends_at <= now()")
	if err != nil {
		return 0, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	closed := 0
	for _, id := range ids {
		ok, err := s.closeLot(ctx, id)
		if err != nil {
			return closed, err
		}
		if ok {
			closed++
		}
	}
	return closed, nil
}

func (s *Store) closeLot(ctx context.Context, lotID int64) (bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	lot, err := lockLot(ctx, tx, lotID)
	if err != nil {
		return false, err
	}
	if lot.Status != "active" || lot.open(time.Now()) {
		return false, nil
	}
	if lot.HasBids {
		if err := addResource(ctx, tx, lot.SellerID, lot.PriceResource, lot.MaxBid); err != nil {
			return false, err
		}
		if err := addResource(ctx, tx, lot.LeaderID, lot.OfferResource, lot.OfferAmount); err != nil {
			return false, err
		}
		_, err = tx.Exec(ctx, "UPDATE lots SET status = 'sold', winner_id = $2, final_price = $3 WHERE id = $1",
			lotID, lot.LeaderID, lot.MaxBid)
	} else {
		if err := addResource(ctx, tx, lot.SellerID, lot.OfferResource, lot.OfferAmount); err != nil {
			return false, err
		}
		_, err = tx.Exec(ctx, "UPDATE lots SET status = 'expired' WHERE id = $1", lotID)
	}
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
