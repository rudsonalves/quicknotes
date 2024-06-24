package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/rudsonalves/quicknotes/internal/repositories"
	"github.com/rudsonalves/quicknotes/utils"
)

type userHandler struct {
	repo          repositories.UserRepository
	passwordUtils utils.PasswordUtils
}

func NewUserHandlers(userRepo repositories.UserRepository) *userHandler {
	return &userHandler{
		repo:          userRepo,
		passwordUtils: utils.NewPasswordUtils(),
	}
}

func (uh *userHandler) SignupForm(w http.ResponseWriter, r *http.Request) error {
	return render(w, http.StatusOK, "user-signup.html", nil)
}

func (uh *userHandler) Signup(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	email := strings.TrimSpace(r.PostFormValue("email"))
	password := strings.TrimSpace(r.PostFormValue("password"))

	data := newUserRequest(email, password)
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9,\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		data.AddFieldError("email", "Use um email válido")
	}

	if data.Password == "" {
		data.AddFieldError("password", "Senha não pode ser vazia")
	}
	if len(data.Password) < 6 {
		data.AddFieldError("password", "Senha deve ter 6 ou mais caracteres")
	}

	if !data.Valid() {
		return render(w, http.StatusUnprocessableEntity, "user-signup.html", data)
	}

	hashPassword, err := uh.passwordUtils.HashPassword(data.Password)
	if err != nil {
		return err
	}

	hashToken := utils.GenerateTokenKey(email)
	user, token, err := uh.repo.Create(r.Context(), email, hashPassword, hashToken)
	if err != nil {
		if err == repositories.ErrDuplicateEmail {
			data.AddFieldError("email", "Email já está em uso")
			return render(w, http.StatusUnprocessableEntity, "user-signup.html", data)
		}
		return err
	}

	fmt.Printf("User Id: %d\n", user.Id.Int)
	fmt.Printf("Token: %s\n", token)

	return render(w, http.StatusOK, "user-signup-success.html", user)
}
