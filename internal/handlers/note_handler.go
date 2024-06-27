package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/rudsonalves/quicknotes/internal/models"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
)

type noteHandler struct {
	repo    repositories.NoteRepository
	session *scs.SessionManager
	render  *render.RenderTemplate
}

func NewNoteHandler(
	session *scs.SessionManager,
	noteRepo repositories.NoteRepository,
	render *render.RenderTemplate) *noteHandler {
	return &noteHandler{
		repo:    noteRepo,
		session: session,
		render:  render}
}

func (nh *noteHandler) getUserIdFromSession(r *http.Request) int64 {
	return nh.session.GetInt64(r.Context(), "userId")
}

func strconvInt64(sValue string) (int64, error) {
	value, err := strconv.ParseInt(sValue, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (nh *noteHandler) NoteList(w http.ResponseWriter, r *http.Request) error {
	notes, err := nh.repo.List(r.Context(), nh.getUserIdFromSession(r))
	if err != nil {
		return err
	}

	return nh.render.RenderPage(w, r, http.StatusOK, "note-home.html", newNoteResponseFromNoteList(notes))
}

func (nh *noteHandler) NoteView(w http.ResponseWriter, r *http.Request) error {
	idParm := r.PathValue("id")
	id, err := strconvInt64(idParm)
	if err != nil {
		return err
	}

	// ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	// defer cancel()
	// note, err := nh.repo.GetById(ctx, id)
	note, err := nh.repo.GetById(r.Context(), id)
	if err != nil {
		return err
	}

	return nh.render.RenderPage(w, r, http.StatusOK, "note-view.html", newNoteResponseFromNote(note))
}

func (nh *noteHandler) NoteNew(w http.ResponseWriter, r *http.Request) error {
	return nh.render.RenderPage(w, r, http.StatusOK, "note-new.html", newNoteRequest(nil))
}

func (nh *noteHandler) NoteSave(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	idParm := r.PostForm.Get("id")
	id, _ := strconvInt64(idParm)
	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")
	color := r.PostForm.Get("color")

	data := newNoteRequest(nil)
	data.Id = id
	data.Color = color
	data.Content = content
	data.Title = title

	// if strings.TrimSpace(title) == "" {
	// 	data.AddFieldError("title", "Título é obrigatório")
	// }
	if strings.TrimSpace(content) == "" {
		data.AddFieldError("content", "Conteúdo é obrigatório")
	}

	if !data.Valid() {
		if id > 0 {
			nh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "note-edit.html", data)
		} else {
			nh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "note-new.html", data)
		}
		return nil
	}

	var err error
	var note *models.Note
	if id > 0 {
		note, err = nh.repo.Update(r.Context(), id, title, content, color)
	} else {
		note, err = nh.repo.Create(r.Context(), nh.getUserIdFromSession(r), title, content, color)
	}
	if err != nil {
		return err
	}

	redirectUrl := fmt.Sprintf("/note/%d", note.Id.Int) // acho que aqui pode ser apenas "note/%d"
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
	return nil
}

func (nh *noteHandler) NoteDelete(w http.ResponseWriter, r *http.Request) error {
	idParm := r.PathValue("id")
	id, err := strconvInt64(idParm)
	if err != nil {
		return err
	}

	if err := nh.repo.Delete(r.Context(), id); err != nil {
		return err
	}

	return nil
}

func (nh *noteHandler) NoteEdit(w http.ResponseWriter, r *http.Request) error {
	idParm := r.PathValue("id")
	id, err := strconvInt64(idParm)
	if err != nil {
		return err
	}

	note, err := nh.repo.GetById(r.Context(), id)
	if err != nil {
		return err
	}
	return nh.render.RenderPage(w, r, http.StatusOK, "note-edit.html", newNoteRequest(note))
}
