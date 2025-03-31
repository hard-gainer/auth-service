package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hard-gainer/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

func New(dbURL string, log *slog.Logger) (*Storage, error) {
	const op = "db.New"

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("%s: failed to ping database: %w", op, err)
	}

	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func (s *Storage) Stop() error {
	s.db.Close()
	return nil
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
		`SELECT id, name, email, pass_hash, role
         FROM users 
         WHERE email = $1`,
		email).Scan(&user.ID, &user.Name, &user.Email, &user.PassHash, &user.Role)

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

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("retrieving user by id")

	var user domain.UserInfo
	err := s.db.QueryRow(ctx,
		`SELECT id, name, email, role
         FROM users 
         WHERE id = $1`,
		userID).Scan(&user.ID, &user.Name, &user.Email, &user.Role)

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
