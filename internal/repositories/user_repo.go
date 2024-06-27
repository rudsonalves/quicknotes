package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/models"
)

var ErrDuplicateEmail = newRepositoryError(errors.New("duplicate email"))
var ErrEmailNotFound = newRepositoryError(errors.New("email not found"))
var ErrInvalidTokenOrUserAlreadyConfirmed = newRepositoryError(errors.New("invalid token or user already confirmed"))

type UserRepository interface {
	Create(ctx context.Context, email, password, hashToken string) (*models.User, string, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ConfirmUserByToken(ctx context.Context, token string) error
	CreateResetPasswordToken(ctx context.Context, email, hashToken string) (string, error)
	GetUserConfirmationByToken(ctx context.Context, token string) (*models.UserConfirmationToken, error)
	UpdatePasswordByToken(ctx context.Context, newPassword, token string) (string, error)
	NewUserConfirmationToken(ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbpoll *pgxpool.Pool) UserRepository {
	return &userRepository{db: dbpoll}
}

// UpdatePasswordByToken implements UserRepository.
func (u *userRepository) UpdatePasswordByToken(ctx context.Context, newPassword string, token string) (string, error) {
	query := `
	SELECT u.id u_id, u.email, t.id t_id FROM users u INNER JOIN users_conf_tokens t
		ON u.id = t.user_id
		WHERE t.confirmed = false
		AND t.token = $1`

	row := u.db.QueryRow(ctx, query, token)
	var userId, tokenId pgtype.Numeric
	var email pgtype.Text
	if err := row.Scan(&userId, &email, &token); err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrInvalidTokenOrUserAlreadyConfirmed
		}
		return "", newRepositoryError(err)
	}

	// update confirmation token
	query = `
	UPDATE users_conf_tokens
		SET confirmed = true, updated_at = now()
		WHERE id = $1`
	_, err := u.db.Exec(ctx, query, tokenId)
	if err != nil {
		return "", newRepositoryError(err)
	}

	// update user password
	query = `
	UPDATE users
		SET password = $1, updated_at = now()
		WHERE id = $2`
	_, err = u.db.Exec(ctx, query, newPassword, userId)
	if err != nil {
		return "", newRepositoryError(err)
	}

	return email.String, nil
}

func (u *userRepository) CreateResetPasswordToken(ctx context.Context, email, hashToken string) (string, error) {
	user, err := u.FindByEmail(ctx, email)
	if err != nil || !user.Active.Bool {
		return "", ErrEmailNotFound
	}

	userToken, err := u.createConfirmationToken(ctx, user, hashToken)
	if err != nil {
		return "", ErrEmailNotFound
	}
	return userToken.Token.String, nil
}

func (u *userRepository) createConfirmationToken(ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error) {
	var userTotken models.UserConfirmationToken
	userTotken.UserId = user.Id
	userTotken.Token = pgtype.Text{String: token, Valid: true}
	query := `
	INSERT INTO users_conf_tokens (user_id, token)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := u.db.QueryRow(ctx, query, userTotken.UserId, userTotken.Token)
	if err := row.Scan(&userTotken.Id, &userTotken.CreatedAt); err != nil {
		return nil, err
	}

	return &userTotken, nil
}

func (u *userRepository) ConfirmUserByToken(ctx context.Context, token string) error {
	query := `
	SELECT u.id, t.id FROM users u INNER JOIN users_conf_tokens t
		ON u.id = t.user_id
		WHERE u.active = false
		AND t.confirmed = false
		AND t.token = $1`
	var userId, totokenId pgtype.Numeric
	row := u.db.QueryRow(ctx, query, token)
	if err := row.Scan(&userId, &totokenId); err != nil {
		if err == pgx.ErrNoRows {
			return ErrInvalidTokenOrUserAlreadyConfirmed
		}
		return newRepositoryError(err)
	}

	queryUpdateUser := `UPDATE users SET active = true, updated_at = now() WHERE id = $1`
	_, err := u.db.Exec(ctx, queryUpdateUser, userId)
	if err != nil {
		slog.Error(err.Error())
		return newRepositoryError(err)
	}

	queryUpdateToken := `UPDATE users_conf_tokens SET confirmed = true, updated_at = now() WHERE id = $1`
	_, err = u.db.Exec(ctx, queryUpdateToken, totokenId)
	if err != nil {
		return newRepositoryError(err)
	}

	return nil
}

func (u *userRepository) Create(ctx context.Context, email string, password string, hashToken string) (*models.User, string, error) {
	var user models.User
	user.Email = pgtype.Text{String: strings.TrimSpace(email), Valid: true}
	user.Password = pgtype.Text{String: strings.TrimSpace(password), Valid: true}

	query := `
	INSERT INTO users (email, password)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := u.db.QueryRow(ctx, query, user.Email, user.Password)
	if err := row.Scan(&user.Id, &user.CreatedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return &user, "", ErrDuplicateEmail
		}
		return nil, "", newRepositoryError(err)
	}

	// generate token confirmation
	userToken, err := u.createConfirmationToken(ctx, &user, hashToken)
	if err != nil {
		return nil, "", newRepositoryError(err)
	}

	return &user, userToken.Token.String, nil
}

func (u *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password, active FROM users WHERE email = $1`

	row := u.db.QueryRow(ctx, query, email)
	if err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Active,
	); err != nil {
		return nil, newRepositoryError(err)
	}

	return &user, nil
}

func (u *userRepository) GetUserConfirmationByToken(ctx context.Context, token string) (*models.UserConfirmationToken, error) {
	var userToken models.UserConfirmationToken
	query := `
	SELECT id, user_id, token, confirmed, created_at, updated_at
		FROM users_conf_tokens
		WHERE token = $1`

	row := u.db.QueryRow(ctx, query, token)
	if err := row.Scan(
		&userToken.Id,
		&userToken.UserId,
		&userToken.Token,
		&userToken.Confirmed,
		&userToken.CreatedAt,
		&userToken.UpdatedAt); err != nil {
		return nil, newRepositoryError(err)
	}

	return &userToken, nil
}

func (u *userRepository) NewUserConfirmationToken(ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error) {
	userToken, err := u.createConfirmationToken(ctx, user, token)
	if err != nil {
		slog.Error(err.Error())
		return nil, newRepositoryError(err)
	}

	return userToken, nil
}
