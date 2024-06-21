package handlers

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	appError "github.com/rudsonalves/quicknotes/internal/app_error"
	"github.com/rudsonalves/quicknotes/internal/repositories"
)

var ErrNotFound = appError.WithStatus(errors.New("página não encontrada"), http.StatusNotFound)
var ErrInternal = appError.WithStatus(errors.New("ocorreu um erro ao executar essa página"), http.StatusInternalServerError)

type HandlerWithError func(w http.ResponseWriter, r *http.Request) error

func (f HandlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		var statusError appError.StatusError
		var repoError *repositories.RepositoriesError

		if errors.As(err, &statusError) {
			if statusError.StatusCode() == http.StatusNotFound {
				files := []string{
					"views/templates/base.html",
					"views/templates/pages/404.html",
				}
				t, err := template.ParseFiles(files...)
				if err != nil {
					http.Error(w, err.Error(), statusError.StatusCode())
				}
				t.ExecuteTemplate(w, "base", statusError.Error())
				return
			}
			http.Error(w, err.Error(), statusError.StatusCode())
			return
		}
		if errors.As(err, &repoError) {
			slog.Error(err.Error())
			http.Error(w, "ocorreu um erro ao executar esta operação", http.StatusInternalServerError)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
