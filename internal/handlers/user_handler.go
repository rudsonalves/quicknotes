package handlers

import (
	"log/slog"
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

func (uh *userHandler) SigninForm(w http.ResponseWriter, r *http.Request) error {
	return render(w, http.StatusOK, "user-signin.html", nil)
}

func (uh *userHandler) Signin(w http.ResponseWriter, r *http.Request) error {
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

	if !data.Valid() {
		return render(w, http.StatusUnprocessableEntity, "user-signin.html", data)
	}

	user, err := uh.repo.GetByEmail(r.Context(), data.Email)
	if err != nil {
		slog.Error("Usuário não encontrado")
		data.AddFieldError("validation", "Credenciais inválidas.")
		return render(w, http.StatusUnauthorized, "user-signin.html", data)
	}

	if !user.Active.Bool {
		data.AddFieldError("validation", "Conta do usuário ainda não foi confirmada.")
		return render(w, http.StatusUnauthorized, "user-signin.html", data)
	}

	if !uh.passwordUtils.ValidatePassword(user.Password.String, data.Password) {
		data.AddFieldError("validation", "Credenciais inválidas.")
		return render(w, http.StatusUnauthorized, "user-signin.html", data)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
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
	_, token, err := uh.repo.Create(r.Context(), email, hashPassword, hashToken)
	if err != nil {
		if err == repositories.ErrDuplicateEmail {
			data.AddFieldError("email", "Email já está em uso")
			return render(w, http.StatusUnprocessableEntity, "user-signup.html", data)
		}
		return err
	}

	return render(w, http.StatusOK, "user-signup-success.html", token)
}

func (uh *userHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	msg := "Seu cadastro foi confirmado."
	if err := uh.repo.ConfirmUserByToken(r.Context(), token); err != nil {
		msg = "Este cadastro já foi confirmado ou token inválido."
	}

	return render(w, http.StatusOK, "user-confirm.html", msg)
}
