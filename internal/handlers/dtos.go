package handlers

import (
	"fmt"

	"github.com/rudsonalves/quicknotes/internal/models"
	"github.com/rudsonalves/quicknotes/internal/validations"
)

type NoteResponse struct {
	Id      int
	Title   string
	Content string
	Color   string
}

func newNoteResponseFromNote(note *models.Note) (resp NoteResponse) {
	resp = NoteResponse{
		Id:      int(note.Id.Int.Int64()),
		Title:   note.Title.String,
		Content: note.Content.String,
		Color:   note.Color.String,
	}
	return
}

type NoteRequest struct {
	Id      int
	Title   string
	Content string
	Color   string
	Colors  []string
	validations.FormValidator
}

type UserRequest struct {
	Email    string
	Password string
	validations.FormValidator
}

func newUserRequest(email, password string) (req UserRequest) {
	req.Email = email
	req.Password = password
	return
}

func newNoteRequest(note *models.Note) (req NoteRequest) {
	for index := 1; index <= 9; index++ {
		req.Colors = append(req.Colors, fmt.Sprintf("color%d", index))
	}
	if note != nil {
		req.Id = int(note.Id.Int.Int64())
		req.Title = note.Title.String
		req.Content = note.Content.String
		req.Color = note.Color.String
	} else {
		req.Color = req.Colors[2]
	}
	return
}

func newNoteResponseFromNoteList(notes []models.Note) (resp []NoteResponse) {
	for _, note := range notes {
		resp = append(resp, newNoteResponseFromNote(&note))
	}

	return
}
