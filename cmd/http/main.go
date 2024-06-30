package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/mailer"
)

func main() {
	config := loadConfig()

	slog.SetDefault(newLogger(os.Stderr, config.GetLevelLog()))

	dbPool, err := pgxpool.New(context.Background(), config.DBConnURL())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()
	slog.Info("Database connection successful")

	// Mail service
	mailPort, _ := strconv.Atoi(config.MailPort)
	smtp := mailer.SMTPConfig{
		Host:     config.MailHost,
		Port:     mailPort,
		UserName: config.MailUserName,
		Password: config.MailUserPass,
		From:     config.MailFrom}
	mailservice := mailer.NewSMTPMailService(smtp)

	sessionManager := scs.New()
	sessionManager.Lifetime = time.Hour
	sessionManager.Store = pgxstore.New(dbPool)
	// Run cleanup every 30 minutes.
	pgxstore.NewWithCleanupInterval(dbPool, 30*time.Minute)

	csrfMiddleware := csrf.Protect([]byte(config.CSRFKey))

	mux := LoadRoutes(dbPool, sessionManager, mailservice)

	addr := fmt.Sprintf(":%s", config.ServerPort)
	slog.Info(fmt.Sprintf("Server running in %s", addr))
	if err := http.ListenAndServeTLS(
		addr,
		"cer.cer",
		"cer.key",
		sessionManager.LoadAndSave(csrfMiddleware(mux))); err != nil {
		panic(err)
	}
}
