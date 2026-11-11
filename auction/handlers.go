package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

type App struct {
	store *Store
	pages map[string]*template.Template
}

var statusNames = map[string]string{
	"active":     "Триває",
	"sold":       "Продано",
	"bought_out": "Викуплено",
	"expired":    "Без ставок",
}

var funcs = template.FuncMap{
	"res":    ResourceName,
	"status": func(s string) string { return statusNames[s] },
	"time":   func(t time.Time) string { return t.Local().Format("02.01.2006 15:04") },
	"left": func(t time.Time) string {
		d := time.Until(t)
		if d <= 0 {
			return "завершується"
		}
		d = d.Round(time.Minute)
		if d < time.Minute {
			return "менше хвилини"
		}
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if h > 0 {
			return fmt.Sprintf("%d год %d хв", h, m)
		}
		return fmt.Sprintf("%d хв", m)
	},
	"minBid":    MinNextBid,
	"canBuyout": BuyoutAvailable,
	"resources": func() []string { return Resources },
}

func NewApp(store *Store) (*App, error) {
	pages := map[string]*template.Template{}
	for _, name := range []string{"index", "register", "login", "new_lot", "lot", "me"} {
		t, err := template.New("layout.html").Funcs(funcs).ParseFS(templateFS, "templates/layout.html", "templates/"+name+".html")
		if err != nil {
			return nil, err
		}
		pages[name] = t
	}
	return &App{store: store, pages: pages}, nil
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.index)
	mux.HandleFunc("GET /register", a.registerForm)
	mux.HandleFunc("POST /register", a.register)
	mux.HandleFunc("GET /login", a.loginForm)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("POST /logout", a.logout)
	mux.HandleFunc("GET /lots/new", a.requireUser(a.newLotForm))
	mux.HandleFunc("POST /lots", a.requireUser(a.createLot))
	mux.HandleFunc("GET /lots/{id}", a.lot)
	mux.HandleFunc("POST /lots/{id}/bid", a.requireUser(a.bid))
	mux.HandleFunc("POST /lots/{id}/buyout", a.requireUser(a.buyout))
	mux.HandleFunc("GET /me", a.requireUser(a.me))
	return a.withUser(mux)
}

type page struct {
	User    *User
	Error   string
	Message string
	Data    any
	Form    url.Values
}

func (a *App) render(w http.ResponseWriter, r *http.Request, name string, status int, p page) {
	p.User = currentUser(r)
	if p.Message == "" {
		p.Message = r.URL.Query().Get("ok")
	}
	if p.Form == nil {
		p.Form = url.Values{}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := a.pages[name].Execute(w, p); err != nil {
		log.Println("template:", err)
	}
}

func (a *App) serverError(w http.ResponseWriter, err error) {
	log.Println("error:", err)
	http.Error(w, "Внутрішня помилка сервера", http.StatusInternalServerError)
}

func redirectWithMessage(w http.ResponseWriter, r *http.Request, path, msg string) {
	http.Redirect(w, r, path+"?ok="+url.QueryEscape(msg), http.StatusSeeOther)
}

func userError(err error) bool {
	for _, e := range []error{
		ErrBadResource, ErrSameResource, ErrBadAmount, ErrBadPrice, ErrBadBuyout, ErrBadDuration,
		ErrNotEnough, ErrOwnLot, ErrLowBid, ErrLotClosed, ErrNoBuyout, ErrBuyoutUnavailable, ErrAlreadyLeading,
	} {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

type indexData struct {
	Active   []Lot
	Finished []Lot
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	if _, err := a.store.CloseExpired(r.Context()); err != nil {
		a.serverError(w, err)
		return
	}
	active, err := a.store.ActiveLots(r.Context())
	if err != nil {
		a.serverError(w, err)
		return
	}
	finished, err := a.store.FinishedLots(r.Context(), 10)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "index", http.StatusOK, page{Data: indexData{Active: active, Finished: finished}})
}

func (a *App) registerForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "register", http.StatusOK, page{})
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некоректний запит", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.PostFormValue("name"))
	email := NormalizeEmail(r.PostFormValue("email"))
	password := r.PostFormValue("password")
	form := url.Values{
		"name": {name}, "email": {email},
		"food": {r.PostFormValue("food")}, "water": {r.PostFormValue("water")}, "air": {r.PostFormValue("air")},
	}
	bal := Balance{Food: atoi(r.PostFormValue("food")), Water: atoi(r.PostFormValue("water")), Air: atoi(r.PostFormValue("air"))}

	if err := ValidateRegistration(name, email, password, r.PostFormValue("confirm")); err != nil {
		a.render(w, r, "register", http.StatusUnprocessableEntity, page{Error: err.Error(), Form: form})
		return
	}
	if err := ValidateStartBalance(bal); err != nil {
		a.render(w, r, "register", http.StatusUnprocessableEntity, page{Error: err.Error(), Form: form})
		return
	}
	hash, err := HashPassword(password)
	if err != nil {
		a.serverError(w, err)
		return
	}
	id, err := a.store.CreateUser(r.Context(), name, email, hash, bal)
	if errors.Is(err, ErrEmailTaken) {
		a.render(w, r, "register", http.StatusUnprocessableEntity, page{Error: err.Error(), Form: form})
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	if err := a.startSession(w, r, id); err != nil {
		a.serverError(w, err)
		return
	}
	redirectWithMessage(w, r, "/me", "Реєстрація успішна! Ваші запаси внесено до системи.")
}

func (a *App) loginForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "login", http.StatusOK, page{})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некоректний запит", http.StatusBadRequest)
		return
	}
	email := NormalizeEmail(r.PostFormValue("email"))
	form := url.Values{"email": {email}}
	u, err := a.store.UserByEmail(r.Context(), email)
	if errors.Is(err, ErrNotFound) || (err == nil && !CheckPassword(u.PasswordHash, r.PostFormValue("password"))) {
		a.render(w, r, "login", http.StatusUnauthorized, page{Error: ErrBadLogin.Error(), Form: form})
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	if err := a.startSession(w, r, u.ID); err != nil {
		a.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = a.store.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) newLotForm(w http.ResponseWriter, r *http.Request) {
	form := url.Values{"offer_resource": {Food}, "price_resource": {Water}, "duration": {"10"}}
	a.render(w, r, "new_lot", http.StatusOK, page{Form: form})
}

func atoi(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return -1
	}
	return n
}

func (a *App) createLot(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некоректний запит", http.StatusBadRequest)
		return
	}
	buyout := 0
	if s := strings.TrimSpace(r.PostFormValue("buyout_price")); s != "" {
		buyout = atoi(s)
		if buyout < 0 {
			a.render(w, r, "new_lot", http.StatusUnprocessableEntity, page{Error: ErrBadBuyout.Error(), Form: r.PostForm})
			return
		}
	}
	in := LotInput{
		OfferResource:   r.PostFormValue("offer_resource"),
		OfferAmount:     atoi(r.PostFormValue("offer_amount")),
		PriceResource:   r.PostFormValue("price_resource"),
		StartPrice:      atoi(r.PostFormValue("start_price")),
		BuyoutPrice:     buyout,
		DurationMinutes: atoi(r.PostFormValue("duration")),
	}
	id, err := a.store.CreateLot(r.Context(), currentUser(r).ID, in)
	if userError(err) {
		a.render(w, r, "new_lot", http.StatusUnprocessableEntity, page{Error: err.Error(), Form: r.PostForm})
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	redirectWithMessage(w, r, fmt.Sprintf("/lots/%d", id), "Лот виставлено на аукціон.")
}

type lotData struct {
	Lot  *Lot
	Bids []Bid
}

func (a *App) loadLot(w http.ResponseWriter, r *http.Request) (*Lot, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, false
	}
	if _, err := a.store.CloseExpired(r.Context()); err != nil {
		a.serverError(w, err)
		return nil, false
	}
	lot, err := a.store.Lot(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return nil, false
	}
	if err != nil {
		a.serverError(w, err)
		return nil, false
	}
	return lot, true
}

func (a *App) renderLot(w http.ResponseWriter, r *http.Request, lot *Lot, status int, errMsg string) {
	bids, err := a.store.Bids(r.Context(), lot.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "lot", status, page{Error: errMsg, Data: lotData{Lot: lot, Bids: bids}})
}

func (a *App) lot(w http.ResponseWriter, r *http.Request) {
	lot, ok := a.loadLot(w, r)
	if !ok {
		return
	}
	a.renderLot(w, r, lot, http.StatusOK, "")
}

func (a *App) bid(w http.ResponseWriter, r *http.Request) {
	lot, ok := a.loadLot(w, r)
	if !ok {
		return
	}
	amount := atoi(r.PostFormValue("amount"))
	err := a.store.PlaceBid(r.Context(), lot.ID, currentUser(r).ID, amount)
	if userError(err) {
		a.renderLot(w, r, lot, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	redirectWithMessage(w, r, fmt.Sprintf("/lots/%d", lot.ID), "Ставку прийнято.")
}

func (a *App) buyout(w http.ResponseWriter, r *http.Request) {
	lot, ok := a.loadLot(w, r)
	if !ok {
		return
	}
	err := a.store.Buyout(r.Context(), lot.ID, currentUser(r).ID)
	if userError(err) {
		a.renderLot(w, r, lot, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	redirectWithMessage(w, r, fmt.Sprintf("/lots/%d", lot.ID), "Лот викуплено!")
}

type meData struct {
	MyLots  []Lot
	MyBids  []Lot
	Balance Balance
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if _, err := a.store.CloseExpired(r.Context()); err != nil {
		a.serverError(w, err)
		return
	}
	fresh, err := a.store.UserByID(r.Context(), u.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	lots, err := a.store.UserLots(r.Context(), u.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	bids, err := a.store.LotsWithUserBids(r.Context(), u.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "me", http.StatusOK, page{Data: meData{MyLots: lots, MyBids: bids, Balance: fresh.Balance}})
}
