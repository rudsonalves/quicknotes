package main

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/handlers"
	"github.com/rudsonalves/quicknotes/internal/mailer"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
	"github.com/rudsonalves/quicknotes/views"
)

func LoadRoutes(
	dbPool *pgxpool.Pool,
	sessionManager *scs.SessionManager,
	mailservice mailer.MailService) http.Handler {
	mux := http.NewServeMux()

	staticFS, err := fs.Sub(views.Files, "static")
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	staticHandler := http.FileServerFS(staticFS)

	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))

	noteRepo := repositories.NewNoteRepository(dbPool)
	userRepo := repositories.NewUserRepository(dbPool)

	render := render.NewRender(sessionManager)

	noteHandler := handlers.NewNoteHandler(sessionManager, noteRepo, render)
	userHandler := handlers.NewUserHandler(sessionManager, userRepo, render, mailservice)

	authMidd := handlers.NewAuthMiddleware(sessionManager)
	errorMidd := handlers.NewErrorHandlerMiddleware(render)

	mux.HandleFunc("GET /", handlers.NewHomeHandler(render).HomeHandler)

	mux.Handle("GET /note", authMidd.RequireAuth(errorMidd.HandleError(noteHandler.NoteList)))
	mux.Handle("GET /note/{id}", authMidd.RequireAuth(errorMidd.HandleError(noteHandler.NoteView)))
	mux.Handle("GET /note/new", authMidd.RequireAuth(errorMidd.HandleError(noteHandler.NoteNew)))
	mux.Handle("POST /note", authMidd.RequireAuth(errorMidd.HandleError(noteHandler.NoteSave)))
	mux.Handle("DELETE /note/{id}", authMidd.RequireAuth(errorMidd.HandleError(noteHandler.NoteDelete)))
	mux.Handle("GET /note/{id}/edit", authMidd.RequireAuth(errorMidd.HandleError(noteHandler.NoteEdit)))

	mux.Handle("GET /user/signup", errorMidd.HandleError(userHandler.SignupForm))
	mux.Handle("POST /user/signup", errorMidd.HandleError(userHandler.Signup))

	mux.Handle("GET /user/signin", errorMidd.HandleError(userHandler.SigninForm))
	mux.Handle("POST /user/signin", errorMidd.HandleError(userHandler.Signin))

	mux.Handle("GET /user/signout", errorMidd.HandleError(userHandler.Signout))

	mux.Handle("GET /user/forgetpassword", errorMidd.HandleError(userHandler.ForgetPasswordForm))
	mux.Handle("POST /user/forgetpassword", errorMidd.HandleError(userHandler.ForgetPassword))
	mux.Handle("POST /user/password", errorMidd.HandleError(userHandler.ResetPassword))
	mux.Handle("GET /user/password/{token}", errorMidd.HandleError(userHandler.ResetPasswordForm))

	mux.Handle("GET /me", authMidd.RequireAuth(errorMidd.HandleError(userHandler.Me)))

	mux.Handle("GET /confirmation/{token}", errorMidd.HandleError(userHandler.Confirm))

	// mux.Handle("GET /confirmation", handlers.HandlerWithError(userHandler.NewConfirmationForm))
	// mux.Handle("POST /confirmation", handlers.HandlerWithError(userHandler.NewConfirmation))

	return mux
}
