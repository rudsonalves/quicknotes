package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	appError "github.com/rudsonalves/quicknotes/internal/app_error"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
)

var ErrNotFound = appError.WithStatus(errors.New("página não encontrada"), http.StatusNotFound)
var ErrInternal = appError.WithStatus(errors.New("ocorreu um erro ao executar essa página"), http.StatusInternalServerError)

type authMiddleware struct {
	session *scs.SessionManager
}

func NewAuthMiddleware(session *scs.SessionManager) *authMiddleware {
	return &authMiddleware{session: session}
}

func (au *authMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := au.session.GetInt64(r.Context(), "userId")
		if userId == 0 {
			slog.Warn("usuário não está logado")
			http.Redirect(w, r, "/user/signin", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type errorHandlerMiddleware struct {
	render *render.RenderTemplate
}

func NewErrorHandlerMiddleware(render *render.RenderTemplate) *errorHandlerMiddleware {
	return &errorHandlerMiddleware{render: render}
}

func (em *errorHandlerMiddleware) HandleError(next func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			var statusError appError.StatusError
			var repoError *repositories.RepositoriesError
			if errors.As(err, &statusError) {
				if statusError.StatusCode() == http.StatusNotFound {
					// render a default err page
					em.render.RenderPage(w, r, http.StatusNotFound, "404.html", nil)
					return
				}
			}

			// repositories errors
			if errors.As(err, &repoError) {
				slog.Error(err.Error())
				em.render.RenderPage(w, r, http.StatusInternalServerError, "generic-error.html", "aconteceu um erro ao executar essa operação.")
				return
			}

			// others generic errors
			slog.Error(err.Error())
			em.render.RenderPage(w, r, http.StatusInternalServerError, "generic-error.html", err.Error())
		}
	})
}
