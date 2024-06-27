package repositories

import (
	"context"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/models"
)

type NoteRepository interface {
	Create(ctx context.Context, userId int64, title, content, color string) (*models.Note, error)
	GetById(ctx context.Context, id int64) (*models.Note, error)
	List(ctx context.Context, userId int64) ([]models.Note, error)
	Update(ctx context.Context, id int64, title, content, color string) (*models.Note, error)
	Delete(ctx context.Context, id int64) error
}

type noteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(dbpool *pgxpool.Pool) NoteRepository {
	return &noteRepository{db: dbpool}
}

func (nr *noteRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM notes WHERE id = $1`

	_, err := nr.db.Exec(ctx, query, id)
	if err != nil {
		return newRepositoryError(err)
	}

	return nil
}

func (nr *noteRepository) Update(ctx context.Context, id int64, title, content, color string) (*models.Note, error) {
	var note models.Note
	note.Id = pgtype.Numeric{Int: big.NewInt(id), Valid: true}

	newTitle := &title
	newContent := &content
	newColor := &color

	if title == "" {
		newTitle = nil
	}
	if content == "" {
		newContent = nil
	}
	if color == "" {
		newColor = nil
	}
	note.UpdatedAt = pgtype.Date{Time: time.Now(), Valid: true}

	query := `
	UPDATE notes
		SET title = COALESCE($1, title),
				content = COALESCE($2, content),
				color = COALESCE($3, color),
				updated_at = $4
		WHERE id = $5`
	_, err := nr.db.Exec(ctx, query,
		newTitle, newContent, newColor, note.UpdatedAt.Time, id)
	if err != nil {
		return nil, newRepositoryError(err)
	}

	return &note, nil
}

func (nr *noteRepository) Create(ctx context.Context, userId int64, title, content, color string) (*models.Note, error) {
	var note models.Note
	note.Title = pgtype.Text{String: title, Valid: true}
	note.Content = pgtype.Text{String: content, Valid: true}
	note.Color = pgtype.Text{String: color, Valid: true}
	query := `
	INSERT INTO notes (user_id, title, content, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	row := nr.db.QueryRow(ctx, query, userId, title, content, color)
	if err := row.Scan(&note.Id, &note.CreatedAt); err != nil {
		return nil, newRepositoryError(err)
	}

	return &note, nil
}

func (nr *noteRepository) List(ctx context.Context, userId int64) ([]models.Note, error) {
	var notes []models.Note
	query := `
	SELECT id, user_id, title, content, color, created_at, updated_at
		FROM notes
		WHERE user_id = $1`

	rows, err := nr.db.Query(ctx, query, userId)
	if err != nil {
		return nil, newRepositoryError(err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, newRepositoryError(err)
	}

	for rows.Next() {
		note := models.Note{}
		err := rows.Scan(
			&note.Id,
			&note.UserId,
			&note.Title,
			&note.Content,
			&note.Color,
			&note.CreatedAt,
			&note.UpdatedAt)
		if err != nil {
			return nil, newRepositoryError(err)
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func (nr *noteRepository) GetById(ctx context.Context, id int64) (*models.Note, error) {
	var note models.Note
	query := `
	SELECT id, user_id, title, content, color, created_at, updated_at
		FROM notes
		WHERE id = $1`

	row := nr.db.QueryRow(ctx, query, id)
	if err := row.Scan(
		&note.Id,
		&note.UserId,
		&note.Title,
		&note.Content,
		&note.Color,
		&note.CreatedAt,
		&note.UpdatedAt,
	); err != nil {
		return nil, newRepositoryError(err)
	}

	return &note, nil
}
