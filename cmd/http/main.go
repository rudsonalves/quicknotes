package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config := loadConfig()

	slog.SetDefault(newLogger(os.Stderr, config.GetLevelLog()))

	dbPool, err := pgxpool.New(context.Background(), config.DBUrl())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()
	slog.Info("Database connection successful")

	sessionManager := scs.New()
	sessionManager.Lifetime = time.Hour
	sessionManager.Store = pgxstore.New(dbPool)
	// Run cleanup every 30 minutes.
	pgxstore.NewWithCleanupInterval(dbPool, 30*time.Minute)

	mux := LoadRoutes(dbPool, sessionManager, config)

	csrfMiddlerewar := csrf.Protect([]byte("32-byte-long-auth-key"))

	slog.Info(
		fmt.Sprintf("Servidor rodando em %s:%s", config.Hostname, config.ServerPort),
	)
	if err := http.ListenAndServe(
		fmt.Sprintf("%s:%s", config.Hostname, config.ServerPort),
		sessionManager.LoadAndSave(csrfMiddlerewar(mux)),
	); err != nil {
		panic(err)
	}
}
