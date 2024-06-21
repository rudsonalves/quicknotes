package repositories

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rudsonalves/quicknotes/internal/models"
)

type NoteRepository interface {
	Create(ctx context.Context, title, content, color string) (*models.Note, error)
	GetById(ctx context.Context, id int) (*models.Note, error)
	List(ctx context.Context) ([]models.Note, error)
	Update(ctx context.Context, id int, title, content, color string) (*models.Note, error)
	Delete(ctx context.Context, id int) error
}

type noteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(dbpool *pgxpool.Pool) NoteRepository {
	return &noteRepository{
		db: dbpool,
	}
}

func (nr *noteRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM notes WHERE id = $1`

	_, err := nr.db.Exec(ctx, query, id)
	if err != nil {
		return newRepositoriesError(err)
	}

	return nil
}

func (nr *noteRepository) Update(ctx context.Context, id int, title, content, color string) (*models.Note, error) {
	var note models.Note
	note.Id = pgtype.Numeric{Int: big.NewInt(int64(id)), Valid: true}

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
		WHERE id = $5
		RETURNING id, title, content, color, created_at, updated_at`
	row := nr.db.QueryRow(ctx, query,
		newTitle, newContent, newColor, note.UpdatedAt.Time, id)
	if err := row.Scan(
		&note.Id,
		&note.Title,
		&note.Content,
		&note.Color,
		&note.CreatedAt,
		&note.UpdatedAt); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &note, nil
}

func (nr *noteRepository) Create(ctx context.Context, title, content, color string) (*models.Note, error) {
	var note models.Note
	note.Title = pgtype.Text{String: title, Valid: true}
	note.Content = pgtype.Text{String: content, Valid: true}
	note.Color = pgtype.Text{String: color, Valid: true}

	query := `
	INSERT INTO notes (title, content, color)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	row := nr.db.QueryRow(ctx, query, title, content, color)
	if err := row.Scan(&note.Id, &note.CreatedAt); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &note, nil
}

func (nr *noteRepository) List(ctx context.Context) ([]models.Note, error) {
	var notes []models.Note
	query := `
	SELECT id, title, content, color, created_at, updated_at FROM notes`

	rows, err := nr.db.Query(ctx, query)
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
		note := models.Note{}
		err := rows.Scan(
			&note.Id,
			&note.Title,
			&note.Content,
			&note.Color,
			&note.CreatedAt,
			&note.UpdatedAt)
		if err != nil {
			slog.Error(err.Error())
			return nil, newRepositoriesError(err)
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func (nr *noteRepository) GetById(ctx context.Context, id int) (*models.Note, error) {
	var note models.Note
	query := `
	SELECT id, title, content, color, created_at, updated_at
		FROM notes WHERE Id = $1`

	row := nr.db.QueryRow(ctx, query, id)
	if err := row.Scan(
		&note.Id,
		&note.Title,
		&note.Content,
		&note.Color,
		&note.CreatedAt,
		&note.UpdatedAt,
	); err != nil {
		return nil, newRepositoriesError(err)
	}

	return &note, nil
}
