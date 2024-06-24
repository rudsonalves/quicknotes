package repositories

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/models"
)

var ErrDuplicateEmail = newRepositoriesError(errors.New("duplicate email"))

type UserRepository interface {
	Create(ctx context.Context, email, password, hashKey string) (*models.User, string, error)
	GetById(ctx context.Context, id int) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, id int, email, password string) (*models.User, error)
	Delete(ctx context.Context, id int) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbpoll *pgxpool.Pool) UserRepository {
	return &userRepository{
		db: dbpoll,
	}
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

	userToken, err := u.createConfirmationToken(ctx, &user, hashKey)
	if err != nil {
		return nil, "", newRepositoriesError(err)
	}

	return &user, userToken.Token.String, nil
}

func (u *userRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := u.db.Exec(ctx, query, id)
	if err != nil {
		return newRepositoriesError(err)
	}

	return nil
}

func (u *userRepository) GetById(ctx context.Context, id int) (*models.User, error) {
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

func (u *userRepository) Update(ctx context.Context, id int, email string, password string) (*models.User, error) {
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

func (u *userRepository) createConfirmationToken(ctx context.Context, user *models.User, token string) (*models.UserConfirmationToken, error) {
	var userTotken models.UserConfirmationToken
	userTotken.UserId = user.Id
	userTotken.Token = pgtype.Text{String: token, Valid: true}
	query := `
	INSERT INTO users_confirmation_tokens (user_id, token)
		VALUES($1, $2)
		RETURNING id, created_at`

	row := u.db.QueryRow(ctx, query, userTotken.UserId, userTotken.Token)
	if err := row.Scan(&userTotken.Id, &userTotken.CreatedAt); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &userTotken, nil
}
