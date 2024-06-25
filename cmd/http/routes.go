package main

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/handlers"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
)

func LoadRoutes(dbPool *pgxpool.Pool, session *scs.SessionManager) http.Handler {

	staticHandler := http.FileServer(http.Dir("views/static/"))
	mux := http.NewServeMux()

	noteRepo := repositories.NewNoteRepository(dbPool)
	userRepo := repositories.NewUserRepository(dbPool)

	reder := render.NewRenderTemplate(session)

	homeHandler := handlers.NewHomeHandler(reder)
	noteHandler := handlers.NewNoteHandlers(session, noteRepo, reder)
	userHandler := handlers.NewUserHandlers(session, userRepo, reder)

	authMiddleware := handlers.NewAuthMiddleware(session)

	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))

	mux.HandleFunc("GET /", homeHandler.HomeView)

	mux.Handle("GET /note", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteList)))
	mux.Handle("GET /note/{id}", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteView)))
	mux.Handle("GET /note/new", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteNew)))
	mux.Handle("POST /note", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteSave)))
	mux.Handle("DELETE /note/{id}", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteDelete)))
	mux.Handle("GET /note/edit/{id}", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteEdit)))

	mux.Handle("GET /user/signup", handlers.HandlerWithError(userHandler.SignupForm))
	mux.Handle("POST /user/signup", handlers.HandlerWithError(userHandler.Signup))

	mux.Handle("GET /user/signin", handlers.HandlerWithError(userHandler.SigninForm))
	mux.Handle("POST /user/signin", handlers.HandlerWithError(userHandler.Signin))

	mux.Handle("GET /user/signout", handlers.HandlerWithError(userHandler.Signout))

	mux.Handle("GET /me", authMiddleware.RequireAuth(handlers.HandlerWithError(userHandler.Me)))

	mux.Handle("GET /confirmation/{token}", handlers.HandlerWithError(userHandler.Confirm))

	return mux
}
