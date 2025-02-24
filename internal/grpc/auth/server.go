package auth

import (
	"context"
	"errors"

	"github.com/hard-gainer/auth-service/internal/db"
	"github.com/hard-gainer/auth-service/internal/domain"
	"github.com/hard-gainer/auth-service/internal/lib/jwt"
	pb "github.com/hard-gainer/auth-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
	RegisterNewUser(ctx context.Context, name, email, password, role string, isAdmin bool) (userId int64, err error)
	Login(ctx context.Context, email, password string, appID int) (token string, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	ValidateToken(ctx context.Context, token string) (domain.User, error)
	GetUser(ctx context.Context, userID int64) (domain.UserInfo, error)
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *pb.LoginRequest,
) (*pb.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, errors.New("invalid credintials")) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetName(), in.GetEmail(), in.GetPassword(), in.GetRole(), in.GetIsAdmin())
	if err != nil {
		if errors.Is(err, db.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &pb.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	in *pb.IsAdminRequest,
) (*pb.IsAdminResponse, error) {
	if in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, in.GetUserId())
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to check admin status")
	}

	return &pb.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverAPI) ValidateToken(
	ctx context.Context,
	in *pb.ValidateTokenRequest,
) (*pb.ValidateTokenResponse, error) {
	if in.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, err := s.auth.ValidateToken(ctx, in.GetToken())
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			return &pb.ValidateTokenResponse{
				IsValid: false,
			}, nil
		}

		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	return &pb.ValidateTokenResponse{
		UserId:  int32(user.ID),
		IsValid: true,
	}, nil
}

func (s *serverAPI) GetUser(
	ctx context.Context,
	in *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {
	user, err := s.auth.GetUser(ctx, int64(in.GetUserId()))
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to retrieve user")
	}

	return &pb.GetUserResponse{
		Id:    int32(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func Register(gRPC *grpc.Server, auth Auth) {
	pb.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}
