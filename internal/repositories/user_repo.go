package repositories

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/models"
)

var ErrDuplicateEmail = newRepositoriesError(errors.New("duplicate email"))
var ErrInvalidTokenOrUserAlreadyConfirmed = newRepositoriesError(errors.New("invalid token or user already confirmed"))

type UserRepository interface {
	Create(ctx context.Context, email, password, hashKey string) (*models.User, string, error)
	GetById(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, id int64, email, password string) (*models.User, error)
	Delete(ctx context.Context, id int64) error
	ConfirmUserByToken(ctx context.Context, token string) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbpoll *pgxpool.Pool) UserRepository {
	return &userRepository{
		db: dbpoll,
	}
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
		return newRepositoriesError(err)
	}

	queryUpdateUser := `UPDATE users SET active = true, updated_at = now() WHERE id = $1`
	_, err := u.db.Exec(ctx, queryUpdateUser, userId)
	if err != nil {
		slog.Error(err.Error())
		return newRepositoriesError(err)
	}

	queryUpdateToken := `UPDATE users_conf_tokens SET confirmed = true, updated_at = now() WHERE id = $1`
	_, err = u.db.Exec(ctx, queryUpdateToken, totokenId)
	if err != nil {
		return newRepositoriesError(err)
	}

	return nil
}

func (u *userRepository) Create(ctx context.Context, email string, password string, hashKey string) (*models.User, string, error) {
	var user models.User
	user.Email = pgtype.Text{String: email, Valid: true}
	user.Password = pgtype.Text{String: password, Valid: true}

	query := `
	INSERT INTO users (email, password)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := u.db.QueryRow(ctx, query, email, password)
	if err := row.Scan(&user.Id, &user.CreatedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return &user, "", ErrDuplicateEmail
		}
		return nil, "", newRepositoriesError(err)
	}

	userToken, err := u.createConfirmToken(ctx, &user, hashKey)
	if err != nil {
		return nil, "", newRepositoriesError(err)
	}

	return &user, userToken.Token.String, nil
}

func (u *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := u.db.Exec(ctx, query, id)
	if err != nil {
		return newRepositoriesError(err)
	}

	return nil
}

func (u *userRepository) GetById(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	query := `
	SELECT id, email, password, active, created_at, updated_at
		FROM users WHERE id = $1`

	row := u.db.QueryRow(ctx, query, id)
	if err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &user, nil
}

func (u *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password, active FROM users WHERE email = $1`

	row := u.db.QueryRow(ctx, query, email)
	if err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Active,
	); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &user, nil
}

func (u *userRepository) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	query := `
	SELECT id, email, password, active, created_at, updated_at FROM users`

	rows, err := u.db.Query(ctx, query)
	if err != nil {
		slog.Error(err.Error())
		return nil, newRepositoriesError(err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return nil, newRepositoriesError(err)
	}

	for rows.Next() {
		user := models.User{}
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Password,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt)
		if err != nil {
			slog.Error(err.Error())
			return nil, newRepositoriesError(err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (u *userRepository) Update(ctx context.Context, id int64, email string, password string) (*models.User, error) {
	var user models.User
	user.Id = pgtype.Numeric{Int: big.NewInt(int64(id)), Valid: true}

	newEmail := &email
	newPassword := &password

	if email == "" {
		newEmail = nil
	}
	if password == "" {
		newPassword = nil
	}
	user.UpdatedAt = pgtype.Date{Time: time.Now(), Valid: true}

	query := `
	UPDATE users
		SET email = COALESCE($1, email),
				password = COALESCE($2, password),
				updated_at = $3
		WHERE id = $4
		RETURNING id, email, password, action, created_at, updated_at`
	row := u.db.QueryRow(ctx, query,
		newEmail, newPassword, user.UpdatedAt, id)
	if err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, newRepositoriesError(err)
	}
	return &user, nil
}

func (u *userRepository) createConfirmToken(ctx context.Context, user *models.User, token string) (*models.ConfirmToken, error) {
	var userTotken models.ConfirmToken
	userTotken.UserId = user.Id
	userTotken.Token = pgtype.Text{String: token, Valid: true}
	query := `
	INSERT INTO users_conf_tokens (user_id, token)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := u.db.QueryRow(ctx, query, userTotken.UserId, userTotken.Token)
	if err := row.Scan(&userTotken.Id, &userTotken.CreatedAt); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &userTotken, nil
}
