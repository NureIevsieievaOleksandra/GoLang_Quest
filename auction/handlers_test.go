package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) (*httptest.Server, *Store) {
	t.Helper()
	s := newTestStore(t)
	app, err := NewApp(s)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(app.Routes())
	t.Cleanup(srv.Close)
	return srv, s
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func post(t *testing.T, c *http.Client, u string, form url.Values) (int, string) {
	t.Helper()
	resp, err := c.PostForm(u, form)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func get(t *testing.T, c *http.Client, u string) (int, string) {
	t.Helper()
	resp, err := c.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func register(t *testing.T, srv *httptest.Server, name, email string) *http.Client {
	t.Helper()
	c := newClient(t)
	code, body := post(t, c, srv.URL+"/register", url.Values{
		"name": {name}, "email": {email}, "password": {"secret123"}, "confirm": {"secret123"}, "food": {"20"}, "water": {"20"}, "air": {"20"},
	})
	if code != http.StatusOK || !strings.Contains(body, "Реєстрація успішна") {
		t.Fatalf("реєстрація %s: код %d", email, code)
	}
	return c
}

func TestRegisterLoginLogout(t *testing.T) {
	srv, _ := newTestServer(t)
	register(t, srv, "Марія", "Maria@Example.com")

	c := newClient(t)
	code, body := post(t, c, srv.URL+"/register", url.Values{
		"name": {"Інша"}, "email": {"maria@example.com"}, "password": {"secret123"}, "confirm": {"secret123"}, "food": {"20"}, "water": {"20"}, "air": {"20"},
	})
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, ErrEmailTaken.Error()) {
		t.Fatalf("повторна реєстрація: код %d", code)
	}

	code, body = post(t, c, srv.URL+"/login", url.Values{"email": {"maria@example.com"}, "password": {"wrong-pass"}})
	if code != http.StatusUnauthorized || !strings.Contains(body, ErrBadLogin.Error()) {
		t.Fatalf("вхід із неправильним паролем: код %d", code)
	}

	code, body = post(t, c, srv.URL+"/login", url.Values{"email": {"maria@example.com"}, "password": {"secret123"}})
	if code != http.StatusOK || !strings.Contains(body, "Вийти") {
		t.Fatalf("вхід: код %d", code)
	}

	post(t, c, srv.URL+"/logout", nil)
	_, body = get(t, c, srv.URL+"/")
	if strings.Contains(body, "Вийти") {
		t.Fatal("після виходу користувач усе ще авторизований")
	}
}

func TestPagesRequireLogin(t *testing.T) {
	srv, _ := newTestServer(t)
	c := newClient(t)
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	for _, path := range []string{"/lots/new", "/me"} {
		resp, err := c.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login" {
			t.Errorf("%s без входу: код %d, переадресація %q", path, resp.StatusCode, resp.Header.Get("Location"))
		}
	}
}

func TestLotFlow(t *testing.T) {
	srv, s := newTestServer(t)
	seller := register(t, srv, "Продавець", "seller@example.com")
	buyer := register(t, srv, "Покупець", "buyer@example.com")

	code, body := post(t, seller, srv.URL+"/lots", url.Values{
		"offer_resource": {Food}, "offer_amount": {"5"}, "price_resource": {Water},
		"start_price": {"3"}, "buyout_price": {"10"}, "duration": {"10"},
	})
	if code != http.StatusOK || !strings.Contains(body, "Лот виставлено") {
		t.Fatalf("створення лота: код %d", code)
	}

	_, body = get(t, buyer, srv.URL+"/")
	if !strings.Contains(body, "5 × Їжа") {
		t.Fatal("лот не відображається на сторінці лотів")
	}

	code, body = post(t, buyer, srv.URL+"/lots/1/bid", url.Values{"amount": {"2"}})
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, ErrLowBid.Error()) {
		t.Fatalf("замала ставка: код %d", code)
	}

	code, body = post(t, buyer, srv.URL+"/lots/1/bid", url.Values{"amount": {"4"}})
	if code != http.StatusOK || !strings.Contains(body, "Ставку прийнято") {
		t.Fatalf("ставка: код %d", code)
	}

	code, body = post(t, seller, srv.URL+"/lots/1/bid", url.Values{"amount": {"5"}})
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, ErrOwnLot.Error()) {
		t.Fatalf("ставка на власний лот: код %d", code)
	}

	code, body = post(t, buyer, srv.URL+"/lots/1/buyout", nil)
	if code != http.StatusOK || !strings.Contains(body, "Лот викуплено") {
		t.Fatalf("викуп: код %d", code)
	}

	u, err := s.UserByEmail(t.Context(), "buyer@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.Balance.Food != StartFood+5 || u.Balance.Water != StartWater-10 {
		t.Fatalf("запаси покупця після викупу: %+v", u.Balance)
	}
}

func TestCreateLotValidationShownToUser(t *testing.T) {
	srv, _ := newTestServer(t)
	c := register(t, srv, "Марія", "maria@example.com")
	code, body := post(t, c, srv.URL+"/lots", url.Values{
		"offer_resource": {Food}, "offer_amount": {"500"}, "price_resource": {Water},
		"start_price": {"3"}, "duration": {"10"},
	})
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, ErrNotEnough.Error()) {
		t.Fatalf("лот більший за запас: код %d", code)
	}
}

func TestRegisterWithOwnResources(t *testing.T) {
	srv, s := newTestServer(t)
	c := newClient(t)
	code, body := post(t, c, srv.URL+"/register", url.Values{
		"name": {"Марія"}, "email": {"maria@example.com"}, "password": {"secret123"}, "confirm": {"secret123"},
		"food": {"7"}, "water": {"-1"}, "air": {"30"},
	})
	if code != http.StatusUnprocessableEntity || !strings.Contains(body, ErrBadStartBalance.Error()) {
		t.Fatalf("від'ємні запаси: код %d", code)
	}
	code, _ = post(t, c, srv.URL+"/register", url.Values{
		"name": {"Марія"}, "email": {"maria@example.com"}, "password": {"secret123"}, "confirm": {"secret123"},
		"food": {"7"}, "water": {"12"}, "air": {"30"},
	})
	if code != http.StatusOK {
		t.Fatalf("реєстрація: код %d", code)
	}
	u, err := s.UserByEmail(t.Context(), "maria@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.Balance != (Balance{Food: 7, Water: 12, Air: 30}) {
		t.Fatalf("запаси після реєстрації: %+v", u.Balance)
	}
}
