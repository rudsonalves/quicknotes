package main

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/handlers"
	"github.com/rudsonalves/quicknotes/internal/mailer"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
)

func LoadRoutes(
	dbPool *pgxpool.Pool,
	sessionManager *scs.SessionManager,
	mailservice mailer.MailService) http.Handler {
	mux := http.NewServeMux()

	staticHandler := http.FileServer(http.Dir("views/static/"))

	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))

	noteRepo := repositories.NewNoteRepository(dbPool)
	userRepo := repositories.NewUserRepository(dbPool)

	render := render.NewRenderTemplate(sessionManager)

	noteHandler := handlers.NewNoteHandlers(sessionManager, noteRepo, render)
	userHandler := handlers.NewUserHandlers(sessionManager, userRepo, render, mailservice)

	authMiddleware := handlers.NewAuthMiddleware(sessionManager)

	mux.HandleFunc("GET /", handlers.NewHomeHandler(render).HomeHandler)

	mux.Handle("GET /note", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteList)))
	mux.Handle("GET /note/{id}", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteView)))
	mux.Handle("GET /note/new", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteNew)))
	mux.Handle("POST /note", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteSave)))
	mux.Handle("DELETE /note/{id}", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteDelete)))
	mux.Handle("GET /note/{id}/edit", authMiddleware.RequireAuth(handlers.HandlerWithError(noteHandler.NoteEdit)))

	mux.Handle("GET /user/signup", handlers.HandlerWithError(userHandler.SignupForm))
	mux.Handle("POST /user/signup", handlers.HandlerWithError(userHandler.Signup))

	mux.Handle("GET /user/signin", handlers.HandlerWithError(userHandler.SigninForm))
	mux.Handle("POST /user/signin", handlers.HandlerWithError(userHandler.Signin))

	mux.Handle("GET /user/signout", handlers.HandlerWithError(userHandler.Signout))

	mux.Handle("GET /user/forgetpassword", handlers.HandlerWithError(userHandler.ForgetPassowrd))

	mux.Handle("GET /me", authMiddleware.RequireAuth(handlers.HandlerWithError(userHandler.Me)))

	mux.Handle("GET /confirmation/{token}", handlers.HandlerWithError(userHandler.Confirm))

	return mux
}
