package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/rudsonalves/quicknotes/internal/mailer"
	"github.com/rudsonalves/quicknotes/internal/render"
	"github.com/rudsonalves/quicknotes/internal/repositories"
	"github.com/rudsonalves/quicknotes/utils"
)

type userHandler struct {
	session *scs.SessionManager
	repo    repositories.UserRepository
	render  *render.RenderTemplate
	mail    mailer.MailService
}

func NewUserHandler(
	session *scs.SessionManager,
	userRepo repositories.UserRepository,
	render *render.RenderTemplate,
	mail mailer.MailService) *userHandler {
	return &userHandler{
		session: session,
		repo:    userRepo,
		render:  render,
		mail:    mail}
}

func (uh *userHandler) Me(w http.ResponseWriter, r *http.Request) error {
	// cookie, err := r.Cookie("session")
	// if err != nil {
	// 	http.Redirect(w, r, "/user/signin", http.StatusTemporaryRedirect)
	// 	return nil
	// }
	// fmt.Fprintf(w, "Email: %s", cookie.Value)
	fmt.Fprint(w, "Dados do usuário")
	return nil
}

func (uh *userHandler) SigninForm(w http.ResponseWriter, r *http.Request) error {
	data := UserRequest{}
	data.Flash = uh.session.PopString(r.Context(), "flash")
	return uh.render.RenderPage(w, r, http.StatusOK, "user-signin.html", data)
}

func (uh *userHandler) Signin(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	email := strings.TrimSpace(r.PostFormValue("email"))
	password := strings.TrimSpace(r.PostFormValue("password"))

	data := newUserRequest(email, password)

	// Check if is a valid email address
	if !utils.IsEmailValid(email) {
		data.AddFieldError("email", "Email inválido")
	}

	// Check if password is valid
	if !utils.IsPasswordValid(data.Password) {
		data.AddFieldError("password", "Senha deve possuir 6 ou mais caracteres com letras e números")
	}

	if !data.Valid() {
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signin.html", data)
	}

	// Get user by email
	user, err := uh.repo.FindByEmail(r.Context(), data.Email)
	if err != nil {
		data.AddFieldError("validation", "Credenciais inválidas.")
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signin.html", data)
	}

	// check if user is active
	if !user.Active.Bool {
		data.AddFieldError("validation", "Conta do usuário ainda não foi confirmada.")
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signin.html", data)
	}

	// check user password
	if !utils.ValidatePassword(user.Password.String, data.Password) {
		data.AddFieldError("validation", "Credenciais inválidas.")
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signin.html", data)
	}

	// Renew token
	err = uh.session.RenewToken(r.Context())
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	// store userId and email in session
	uh.session.Put(r.Context(), "userId", user.Id.Int.Int64())
	uh.session.Put(r.Context(), "userEmail", user.Email.String)

	http.Redirect(w, r, "/note", http.StatusSeeOther)
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

	// Check if is a valid email address
	if !utils.IsEmailValid(email) {
		data.AddFieldError("email", "Email inválido")
	}

	// Check if password is valid
	if !utils.IsPasswordValid(data.Password) {
		data.AddFieldError("password", "Senha deve possuir 6 ou mais caracteres com letras e números")
	}

	if !data.Valid() {
		return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signup.html", data)
	}

	// create password hash
	hashPassword, err := utils.HashPassword(data.Password)
	if err != nil {
		return err
	}

	hashToken := utils.GenerateTokenKey()
	_, confirmationToken, err := uh.repo.Create(r.Context(), data.Email, hashPassword, hashToken)
	if err != nil {
		if err == repositories.ErrDuplicateEmail {
			data.AddFieldError("email", "Email já está em uso")
			return uh.render.RenderPage(w, r, http.StatusUnprocessableEntity, "user-signup.html", data)
		}
		return err
	}

	// Send email with account confirmation link
	rdata := map[string]string{"token": confirmationToken}
	body, err := uh.render.RenderMailBody(r, "confirmation.html", rdata)
	if err != nil {
		return err
	}
	if err := uh.mail.Send(mailer.MailMessage{
		To:      []string{data.Email},
		Subject: "Confirmação de Cadastro",
		IsHtml:  true,
		Body:    body,
	}); err != nil {
		return err
	}

	return uh.render.RenderPage(w, r, http.StatusOK, "user-signup-success.html", confirmationToken)
}

func (uh *userHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	msg := "Seu cadastro foi confirmado. Agora você já pode fazer o login no sistema."
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

func (uh *userHandler) ForgetPasswordForm(w http.ResponseWriter, r *http.Request) error {
	return uh.render.RenderPage(w, r, http.StatusOK, "user-forget-password.html", nil)
}

func (uh *userHandler) ForgetPassword(w http.ResponseWriter, r *http.Request) error {
	// get email from form
	email := r.PostFormValue("email")

	// generate token
	hashToken := utils.GenerateTokenKey()

	// insert a new record in tokens table (user_conf_tokens)
	token, err := uh.repo.CreateResetPasswordToken(r.Context(), email, hashToken)
	if err != nil {
		data := UserRequest{}
		data.Email = email
		data.AddFieldError("email", "Email não possui cadastro válido ou confirmado")
		return uh.render.RenderPage(w, r, http.StatusOK, "user-forget-password.html", data)
	}

	// send email with link to reset password
	rdata := map[string]string{"token": token}
	body, err := uh.render.RenderMailBody(r, "forgetpassword.html", rdata)
	if err != nil {
		return err
	}

	if err := uh.mail.Send(mailer.MailMessage{
		To:      []string{email},
		Subject: "Restaurar senha",
		IsHtml:  true,
		Body:    body,
	}); err != nil {
		return err
	}

	msg := "Foi enviado um email com um link para que você possa resetar a sua senha."

	return uh.render.RenderPage(w, r, http.StatusOK, "generic-success.html", msg)
}

func (uh *userHandler) ResetPasswordForm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")

	userToken, err := uh.repo.GetUserConfirmationByToken(r.Context(), token)
	elapsedTime := time.Since(userToken.CreatedAt.Time).Hours()
	if err != nil || userToken.Confirmed.Bool || elapsedTime > 4 {
		msg := "Token inválido ou expirado. Solicite uma nova alteração."
		return uh.render.RenderPage(w, r, http.StatusOK, "generic-error.html", msg)
	}

	data := struct {
		Token  string
		Errors []string
	}{
		Token: token,
	}
	return uh.render.RenderPage(w, r, http.StatusOK, "user-reset-password.html", data)
}

func (uh *userHandler) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	// get password data
	password := r.PostFormValue("password")
	token := r.PostFormValue("token")

	// password hash
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		data := struct {
			Token  string
			Errors []string
		}{
			Token:  token,
			Errors: []string{"Não foi possível alterar a senha. Solicite uma nova alteração."},
		}
		slog.Error(err.Error())
		return uh.render.RenderPage(w, r, http.StatusOK, "user-reset-password.html", data)
	}

	// update password in database
	email, err := uh.repo.UpdatePasswordByToken(r.Context(), hashedPassword, token)
	if err != nil {
		data := struct {
			Token  string
			Errors []string
		}{
			Token:  token,
			Errors: []string{"Não foi possível alterar a senha. Solicite uma nova alteração."},
		}
		slog.Error(err.Error())
		return uh.render.RenderPage(w, r, http.StatusOK, "user-reset-password.html", data)
	}

	// send email informing the password was updated
	uh.mail.Send(mailer.MailMessage{
		To:      []string{email},
		Subject: "Sua senha foi atualizada",
		Body:    []byte("Sua senha foi atualizada e agora você já pode fazer o login novamente."),
	})

	uh.session.Put(r.Context(), "flash", "Sua senha foi atualizada. Agora você pode fazer o login.")

	http.Redirect(w, r, "/user/signin", http.StatusSeeOther)
	return nil
}
