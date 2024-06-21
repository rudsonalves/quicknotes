package handlers

import (
	"fmt"
	"net/http"
)

type userHandler struct{}

func NewUserHandlers() *userHandler {
	return &userHandler{}
}

func (uh *userHandler) SignupForm(w http.ResponseWriter, r *http.Request) error {
	return render(w, http.StatusOK, "user-signup.html", nil)
}

func (uh *userHandler) Signup(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	fmt.Println(email, password)
	return render(w, http.StatusOK, "user-signup.html", nil)
}
