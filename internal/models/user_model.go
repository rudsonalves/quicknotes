package models

import "github.com/jackc/pgx/v5/pgtype"

type User struct {
	Id        pgtype.Numeric
	Email     pgtype.Text
	Password  pgtype.Text
	Active    pgtype.Bool
	CreatedAt pgtype.Date
	UpdatedAt pgtype.Date
}
