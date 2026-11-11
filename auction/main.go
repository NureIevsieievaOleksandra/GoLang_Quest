package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Не задано DATABASE_URL. Приклад: postgres://postgres:ПАРОЛЬ@localhost:5432/auction")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("Не вдалося підключитися до бази: ", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("База даних не відповідає: ", err)
	}

	store := NewStore(db)
	if err := store.Migrate(ctx); err != nil {
		log.Fatal("Не вдалося створити таблиці: ", err)
	}

	app, err := NewApp(store)
	if err != nil {
		log.Fatal(err)
	}

	go closeExpiredLoop(ctx, store)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Аукціон запущено: http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func closeExpiredLoop(ctx context.Context, store *Store) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := store.CloseExpired(ctx)
			if err != nil && ctx.Err() == nil {
				log.Println("закриття лотів:", err)
			}
			if n > 0 {
				log.Printf("завершено аукціонів: %d", n)
			}
		}
	}
}
