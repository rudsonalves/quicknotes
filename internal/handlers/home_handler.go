package handlers

import (
	"net/http"

	"github.com/rudsonalves/quicknotes/internal/render"
)

type homeHandler struct {
	render *render.RenderTemplate
}

func NewHomeHandler(render *render.RenderTemplate) *homeHandler {
	return &homeHandler{render: render}
}

func (hh *homeHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	hh.render.RenderPage(w, r, http.StatusOK, "home.html", nil)
}
