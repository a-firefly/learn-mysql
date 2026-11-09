package main

import (
	"context"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

type Scanner interface {
	Scan(dest ...any) error
}

// 查询单条记录
func (s *Store) One[T any](
	ctx context.Context,
	query string,
	scan func(Scanner) (T, error),
	args ...any,
) (T, error) {
	return scan(s.db.QueryRowContext(ctx, query, args...))
}

func scanUser(s Scanner) (User, error) {
	var u User
	err := s.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	return u, err
}