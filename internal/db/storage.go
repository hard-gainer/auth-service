package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/hard-gainer/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Storage struct {
	db *pgx.Conn
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

func New(dbURL string) (*Storage, error) {
	const op = "db.New"

	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close(context.Background())
}

func (s *Storage) SaveUser(ctx context.Context, name, email string, passHash []byte, role string, isAdmin bool) (int64, error) {
	const op = "db.SaveUser"

	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO users(name, email, pass_hash, role, is_admin)
         VALUES($1, $2, $3, $4, $5)
         RETURNING id`,
		name, email, passHash, role, isAdmin).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	const op = "db.GetUserByEmail"

	var user domain.User
	err := s.db.QueryRow(ctx,
		`SELECT id, name, email, pass_hash 
         FROM users 
         WHERE email = $1`,
		email).Scan(&user.ID, &user.Name, &user.Email, &user.PassHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID int64) (domain.UserInfo, error) {
	const op = "db.GetUserById"

	var user domain.UserInfo
	err := s.db.QueryRow(ctx,
		`SELECT id, name, email
         FROM users 
         WHERE id = $1`,
		userID).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserInfo{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return domain.UserInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, id int) (domain.App, error) {
	const op = "db.App"

	var app domain.App
	err := s.db.QueryRow(ctx,
		`SELECT id, name, secret 
         FROM apps 
         WHERE id = $1`,
		id).Scan(&app.ID, &app.Name, &app.Secret)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.App{}, fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		return domain.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "db.IsAdmin"

	var isAdmin bool
	err := s.db.QueryRow(ctx,
		`SELECT is_admin 
         FROM users 
         WHERE id = $1`,
		userID).Scan(&isAdmin)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}
