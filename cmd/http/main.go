package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/handlers"
	"github.com/rudsonalves/quicknotes/internal/repositories"
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

	noteRepo := repositories.NewNoteRepository(dbPool)

	mux := http.NewServeMux()

	staticHandler := http.FileServer(http.Dir("views/static/"))

	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))

	noteHandler := handlers.NewNoteHandlers(noteRepo)
	userHandler := handlers.NewUserHandlers()

	mux.Handle("/", handlers.HandlerWithError(noteHandler.NoteList))
	mux.Handle("GET /note/{id}", handlers.HandlerWithError(noteHandler.NoteView))
	mux.Handle("GET /note/new", handlers.HandlerWithError(noteHandler.NoteNew))
	mux.Handle("POST /note", handlers.HandlerWithError(noteHandler.NoteSave))
	mux.Handle("DELETE /note/{id}", handlers.HandlerWithError(noteHandler.NoteDelete))
	mux.Handle("GET /note/edit/{id}", handlers.HandlerWithError(noteHandler.NoteEdit))

	mux.Handle("GET /user/signup", handlers.HandlerWithError(userHandler.SignupForm))
	mux.Handle("POST /user/signup", handlers.HandlerWithError(userHandler.Signup))

	slog.Info(
		fmt.Sprintf("Servidor rodando em %s:%s", config.Hostname, config.ServerPort),
	)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.Hostname, config.ServerPort), mux); err != nil {
		panic(err)
	}
}
