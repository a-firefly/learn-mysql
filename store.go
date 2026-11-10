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

// 查询所有用户
func (s *Store) All[T any](
	ctx context.Context,
	query string,
	scan func(Scanner) (T, error),
	args ...any,
) ([]T, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]T, 0)

	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func scanUser(s Scanner) (User, error) {
	var u User
	err := s.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	return u, err
}