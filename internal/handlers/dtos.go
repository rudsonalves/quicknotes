package handlers

import (
	"fmt"

	"github.com/rudsonalves/quicknotes/internal/models"
	"github.com/rudsonalves/quicknotes/internal/validations"
)

type NoteResponse struct {
	Id      int64
	Title   string
	Content string
	Color   string
}

func newNoteResponseFromNote(note *models.Note) (resp NoteResponse) {

	resp.Id = note.Id.Int.Int64()
	resp.Title = note.Title.String
	resp.Content = note.Content.String
	resp.Color = note.Color.String
	return
}

type NoteRequest struct {
	Id      int64
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
		req.Id = note.Id.Int.Int64()
		req.Title = note.Title.String
		req.Color = note.Color.String
		req.Content = note.Content.String
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
