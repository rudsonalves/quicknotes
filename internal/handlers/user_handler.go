package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/rudsonalves/quicknotes/internal/mailer"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
	"github.com/rudsonalves/quicknotes/utils"
)

type userHandler struct {
	session       *scs.SessionManager
	repo          repositories.UserRepository
	render        *render.RenderTemplate
	mail          mailer.MailService
	passwordUtils utils.PasswordUtils
}

func NewUserHandlers(
	session *scs.SessionManager,
	userRepo repositories.UserRepository,
	render *render.RenderTemplate,
	mail mailer.MailService,
) *userHandler {
	return &userHandler{
		session:       session,
		repo:          userRepo,
		render:        render,
		mail:          mail,
		passwordUtils: utils.NewPasswordUtils(),
	}
}

func (uh *userHandler) SigninForm(w http.ResponseWriter, r *http.Request) error {
	return uh.render.RenderPage(w, r, http.StatusOK, "user-signin.html", nil)
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
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signin.html", data)
	}

	user, err := uh.repo.GetByEmail(r.Context(), data.Email)
	if err != nil {
		slog.Error("Usuário não encontrado")
		data.AddFieldError("validation", "Credenciais inválidas.")
		return uh.render.RenderPage(w, r, http.StatusUnauthorized, "user-signin.html", data)
	}

	if !user.Active.Bool {
		data.AddFieldError("validation", "Conta do usuário ainda não foi confirmada.")
		return uh.render.RenderPage(w, r, http.StatusUnauthorized, "user-signin.html", data)
	}

	if !uh.passwordUtils.ValidatePassword(user.Password.String, data.Password) {
		data.AddFieldError("validation", "Credenciais inválidas.")
		return uh.render.RenderPage(w, r, http.StatusUnauthorized, "user-signin.html", data)
	}

	// Renew token
	err = uh.session.RenewToken(r.Context())
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	uh.session.Put(r.Context(), "userId", user.Id.Int.Int64())
	uh.session.Put(r.Context(), "userEmail", user.Email.String)

	http.Redirect(w, r, "/note", http.StatusSeeOther)
	return nil
}

func (uh *userHandler) Me(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "Dados do usuário")
	return nil
}

func (uh *userHandler) SignupForm(w http.ResponseWriter, r *http.Request) error {
	return uh.render.RenderPage(w, r, http.StatusOK, "user-signup.html", nil)
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
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signup.html", data)
	}

	hashPassword, err := uh.passwordUtils.HashPassword(data.Password)
	if err != nil {
		return err
	}

	hashToken := utils.GenerateTokenKey(email)
	_, confirmationToken, err := uh.repo.Create(r.Context(), email, hashPassword, hashToken)
	if err != nil {
		if err == repositories.ErrDuplicateEmail {
			data.AddFieldError("email", "Email já está em uso")
			return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signup.html", data)
		}
		return err
	}

	// Enviar email de confirmação de cadastro
	body, err := uh.render.RenderMailBody("confirmation.html", confirmationToken)
	if err != nil {
		return err
	}
	msg := mailer.MailMessage{
		To:      []string{email},
		Subject: "Confirmação de Cadastro",
		IsHtml:  true,
		Body:    body,
	}
	uh.mail.Send(msg)

	return uh.render.RenderPage(w, r, http.StatusOK, "user-signup-success.html", confirmationToken)
}

func (uh *userHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	msg := "Seu cadastro foi confirmado."
	if err := uh.repo.ConfirmUserByToken(r.Context(), token); err != nil {
		msg = "Este cadastro já foi confirmado ou token inválido."
	}

	return uh.render.RenderPage(w, r, http.StatusOK, "user-confirm.html", msg)
}

func (uh *userHandler) Signout(w http.ResponseWriter, r *http.Request) error {
	// Renew token
	err := uh.session.RenewToken(r.Context())
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	uh.session.Remove(r.Context(), "userId")
	http.Redirect(w, r, "/user/signin", http.StatusSeeOther)
	return nil
}
