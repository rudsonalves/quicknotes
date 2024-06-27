package repositories

import (
	"context"
	"errors"
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
	// NewUserConfirmationToken(ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbpoll *pgxpool.Pool) UserRepository {
	return &userRepository{db: dbpoll}
}

func (ur *userRepository) getUserIdEmailTokenIdFromToken(ctx context.Context, token string) (userId pgtype.Numeric, email pgtype.Text, tokenId pgtype.Numeric, err error) {
	query := `
	SELECT u.id u_id, u.email, t.id t_id FROM users u INNER JOIN users_conf_tokens t
		ON u.id = t.user_id
		WHERE t.confirmed = false
		AND t.token = $1`

	row := ur.db.QueryRow(ctx, query, token)
	err = row.Scan(&userId, &email, &token)
	return
}

func (ur *userRepository) UpdatePasswordByToken(ctx context.Context, newPassword string, token string) (string, error) {
	userId, email, tokenId, err := ur.getUserIdEmailTokenIdFromToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrInvalidTokenOrUserAlreadyConfirmed
		}
		return "", newRepositoryError(err)
	}

	// transaction scope
	tx, err := ur.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		fail(err)
	}
	defer tx.Rollback(ctx)

	// update confirmation token
	query := `
	UPDATE users_conf_tokens
		SET confirmed = true, updated_at = now()
		WHERE id = $1`
	_, err = tx.Exec(ctx, query, tokenId)
	if err != nil {
		return "", fail(err)
	}

	// update user password
	query = `
	UPDATE users
		SET password = $1, updated_at = now()
		WHERE id = $2`
	_, err = tx.Exec(ctx, query, newPassword, userId)
	if err != nil {
		return "", fail(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fail(err)
	}

	return email.String, nil
}

func (ur *userRepository) CreateResetPasswordToken(ctx context.Context, email, hashToken string) (string, error) {
	user, err := ur.FindByEmail(ctx, email)
	if err != nil || !user.Active.Bool {
		return "", fail(ErrEmailNotFound)
	}

	tx, err := ur.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fail(err)
	}
	defer tx.Rollback(ctx)

	userToken, err := ur.createConfirmationToken(tx, ctx, user, hashToken)
	if err != nil {
		return "", fail(ErrEmailNotFound)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fail(err)
	}

	return userToken.Token.String, nil
}

func (ur *userRepository) fetchUserDetailsByToken(ctx context.Context, token string) (userId pgtype.Numeric, totokenId pgtype.Numeric, err error) {
	query := `
	SELECT u.id, t.id FROM users u INNER JOIN users_conf_tokens t
		ON u.id = t.user_id
		WHERE u.active = false
		AND t.confirmed = false
		AND t.token = $1`
	row := ur.db.QueryRow(ctx, query, token)
	err = row.Scan(&userId, &totokenId)
	return
}

func (ur *userRepository) ConfirmUserByToken(ctx context.Context, token string) error {
	userId, totokenId, err := ur.fetchUserDetailsByToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrInvalidTokenOrUserAlreadyConfirmed
		}
		return fail(err)
	}

	// Transaction scope
	tx, err := ur.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback(ctx)

	queryUpdateUser := `UPDATE users SET active = true, updated_at = now() WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdateUser, userId)
	if err != nil {
		return fail(err)
	}

	queryUpdateToken := `UPDATE users_conf_tokens SET confirmed = true, updated_at = now() WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdateToken, totokenId)
	if err != nil {
		return fail(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fail(err)
	}

	return nil
}

func (ur *userRepository) createConfirmationToken(tx pgx.Tx, ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error) {
	var userTotken models.UserConfirmationToken
	userTotken.UserId = user.Id
	userTotken.Token = pgtype.Text{String: token, Valid: true}
	query := `
	INSERT INTO users_conf_tokens (user_id, token)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := tx.QueryRow(ctx, query, userTotken.UserId, userTotken.Token)
	if err := row.Scan(&userTotken.Id, &userTotken.CreatedAt); err != nil {
		return nil, fail(err)
	}

	return &userTotken, nil
}

func (ur *userRepository) Create(ctx context.Context, email string, password string, hashToken string) (*models.User, string, error) {
	var user models.User
	user.Email = pgtype.Text{String: strings.TrimSpace(email), Valid: true}
	user.Password = pgtype.Text{String: strings.TrimSpace(password), Valid: true}

	tx, err := ur.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, "", fail(err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO users (email, password)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := tx.QueryRow(ctx, query, user.Email, user.Password)
	if err := row.Scan(&user.Id, &user.CreatedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return &user, "", fail(ErrDuplicateEmail)
		}
		return nil, "", fail(err)
	}

	// generate token confirmation
	userToken, err := ur.createConfirmationToken(tx, ctx, &user, hashToken)
	if err != nil {
		return nil, "", fail(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", fail(err)
	}

	return &user, userToken.Token.String, nil
}

func (ur *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password, active FROM users WHERE email = $1`

	row := ur.db.QueryRow(ctx, query, email)
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

func (ur *userRepository) GetUserConfirmationByToken(ctx context.Context, token string) (*models.UserConfirmationToken, error) {
	var userToken models.UserConfirmationToken
	query := `
	SELECT id, user_id, token, confirmed, created_at, updated_at
		FROM users_conf_tokens
		WHERE token = $1`

	row := ur.db.QueryRow(ctx, query, token)
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

// func (u *userRepository) NewUserConfirmationToken(ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error) {
// 	userToken, err := u.createConfirmationToken(ctx, user, token)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return nil, newRepositoryError(err)
// 	}

// 	return userToken, nil
// }
