package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"shortener/internal/storage"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(connectionString string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	query := "INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id"

	var id int64
	err := s.db.QueryRow(ctx, query, urlToSave, alias).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 - уникальное ограничение нарушено
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	query := "SELECT url FROM url WHERE alias = $1"

	var resURL string
	err := s.db.QueryRow(ctx, query, alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s execute statement: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	const op = "storage.postgres.DeleteURL"

	query := "DELETE FROM url WHERE alias = $1"

	res, err := s.db.Exec(ctx, query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
