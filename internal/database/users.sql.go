// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package database

import (
	"context"
	"time"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
	gen_random_uuid(),
	$1,
	$2,
	$3
) RETURNING id, created_at, updated_at, email
`

type CreateUserParams struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.CreatedAt, arg.UpdatedAt, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
	)
	return i, err
}
