package models

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	Id        pgtype.Numeric
	Email     pgtype.Text
	Password  pgtype.Text
	Active    pgtype.Bool
	CreatedAt pgtype.Date
	UpdatedAt pgtype.Date
}

func (u *User) String() string {
	return fmt.Sprintf("User {Id: %d  Email: %s  Active: %v}", u.Id.Int, u.Email.String, u.Active.Bool)
}
