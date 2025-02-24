package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hard-gainer/auth-service/internal/db"
	"github.com/hard-gainer/auth-service/internal/domain"
	"github.com/hard-gainer/auth-service/internal/lib/jwt"
	sl "github.com/hard-gainer/auth-service/internal/lib/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		name string,
		email string,
		passHash []byte,
		role string,
		isAdmin bool,
	) (uid int64, err error)
}

type UserProvider interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, userID int64) (domain.UserInfo, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (domain.App, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, name, email, password, role string, isAdmin bool) (int64, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, name, email, passHash, role, isAdmin)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) ValidateToken(ctx context.Context, token string) (domain.User, error) {
	const op = "Auth.ValidateToken"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("validating token")

	app, err := a.appProvider.App(ctx, 1)
	if err != nil {
		log.Info("no app with id 1")
	}

	claims, err := jwt.ValidateToken(token, app.Secret)
	if err != nil {
		log.Warn("invalid token", sl.Err(err))
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.usrProvider.GetUserByEmail(ctx, claims.Email)
	if err != nil {
		log.Error("failed to get user", sl.Err(err))
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (a *Auth) GetUser(ctx context.Context, userID int64) (domain.UserInfo, error) {
	const op = "Auth.GetUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("retrieving user")

	user, err := a.usrProvider.GetUserByID(ctx, userID)
	if err != nil {
		log.Error("failed to get user", sl.Err(err))
		return domain.UserInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
