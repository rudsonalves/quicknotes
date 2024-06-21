package models

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type Note struct {
	Id        pgtype.Numeric
	Title     pgtype.Text
	Content   pgtype.Text
	Color     pgtype.Text
	CreatedAt pgtype.Date
	UpdatedAt pgtype.Date
}

func (n Note) String() string {
	return fmt.Sprintf(
		"Note{id: %d, title: %s, Content: %s, Color: %s, Created At: %v, Updated At: %v}",
		n.Id.Int, n.Title.String, n.Content.String, n.Color.String, n.CreatedAt.Time, n.UpdatedAt.Time)
}
