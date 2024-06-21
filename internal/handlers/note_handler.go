package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	appError "github.com/rudsonalves/quicknotes/internal/app_error"
	"github.com/rudsonalves/quicknotes/internal/models"
	"github.com/rudsonalves/quicknotes/internal/repositories"
)

type noteHandler struct {
	repo repositories.NoteRepository
}

func NewNoteHandlers(noteRepo repositories.NoteRepository) *noteHandler {
	return &noteHandler{repo: noteRepo}
}

func (nh *noteHandler) NoteList(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != "/" {
		return errors.New("lista não encontrada")
	}

	notes, err := nh.repo.List(r.Context())
	if err != nil {
		return err
	}

	return render(w, http.StatusOK, "home.html", newNoteResponseFromNoteList(notes))
}

func (nh *noteHandler) NoteView(w http.ResponseWriter, r *http.Request) error {
	idParm := r.PathValue("id")
	id, err := strconv.Atoi(idParm)
	if err != nil {
		err := errors.New("id da nota não foi fornecida")
		return appError.WithStatus(err, http.StatusBadRequest)
	}

	// ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	// defer cancel()
	// note, err := nh.repo.GetById(ctx, id)
	note, err := nh.repo.GetById(r.Context(), id)
	if err != nil {
		return err
	}

	return render(w, http.StatusOK, "note-view.html", newNoteResponseFromNote(note))
}

func (nh *noteHandler) NoteSave(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return appError.WithStatus(errors.New("error no formulário"), http.StatusBadRequest)
	}
	idParm := r.PostForm.Get("id")
	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")
	color := r.PostForm.Get("color")

	var err error
	id, _ := strconv.Atoi(idParm)

	data := newNoteRequest(nil)
	data.Color = color
	data.Content = content
	data.Title = title
	if strings.TrimSpace(title) == "" {
		data.AddFieldError("title", "Título é obrigatório")
	}
	if strings.TrimSpace(content) == "" {
		data.AddFieldError("content", "Conteúdo é obrigatório")
	}

	if !data.Valid() {
		if id > 0 {
			data.Id = id
			render(w, http.StatusUnprocessableEntity, "note-edit.html", data)
		} else {
			render(w, http.StatusUnprocessableEntity, "note-new.html", data)
		}
		return nil
	}

	var note *models.Note
	if id > 0 {
		note, err = nh.repo.Update(r.Context(), id, title, content, color)
	} else {
		note, err = nh.repo.Create(r.Context(), title, content, color)
	}
	if err != nil {
		return err
	}

	redirectUrl := fmt.Sprintf("note/%d", note.Id.Int)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
	return nil
}

func (nh *noteHandler) NoteNew(w http.ResponseWriter, r *http.Request) error {
	return render(w, http.StatusOK, "note-new.html", newNoteRequest(nil))
}

func (nh *noteHandler) NoteDelete(w http.ResponseWriter, r *http.Request) error {
	idParm := r.PathValue("id")

	id, err := strconv.Atoi(idParm)
	if err != nil {
		err := errors.New("id da nota não foi fornecida")
		return appError.WithStatus(err, http.StatusBadRequest)
	}

	if err := nh.repo.Delete(r.Context(), id); err != nil {
		return err
	}

	return nil
}

func (nh *noteHandler) NoteEdit(w http.ResponseWriter, r *http.Request) error {
	idParm := r.PathValue("id")
	id, err := strconv.Atoi(idParm)
	if err != nil {
		err := errors.New("id da nota não foi fornecida")
		return appError.WithStatus(err, http.StatusBadRequest)
	}

	note, err := nh.repo.GetById(r.Context(), id)
	if err != nil {
		return err
	}
	return render(w, http.StatusOK, "note-edit.html", newNoteRequest(note))
}
