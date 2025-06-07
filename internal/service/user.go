package service

import (
	"context"
	"fmt"

	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/repository"
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword string, password string) bool
}

type UserCacher interface {
	GetEmailCode(ctx context.Context, email string) (string, error)
	SetEmailCode(ctx context.Context, email, emailCode string) error
}

type EmailSender interface {
	Send(email, emailCode string) error
}

type NumberRandomer interface {
	Generate() string
}

type JWTer interface {
	Generate(email string, userID int) (string, error)
}

type UserService struct {
	repo           repository.URLRepository
	passwordHasher PasswordHasher
	jwter          JWTer
	userCacher     UserCacher
	emailSender    EmailSender
	numberRandomer NumberRandomer
}

func NewUserService(repo repository.URLRepository, p PasswordHasher, j JWTer, u UserCacher, e EmailSender, n NumberRandomer) *UserService {
	return &UserService{
		repo:           repo,
		passwordHasher: p,
		jwter:          j,
		userCacher:     u,
		emailSender:    e,
		numberRandomer: n,
	}
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	if !s.passwordHasher.ComparePassword(user.PasswordHash, req.Password) {
		return nil, dto.ErrUserNameOrPasswordFailed
	}

	accessToken, err := s.jwter.Generate(user.Email, int(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to genrate access token: %v", err)
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		Email:       user.Email,
		UserID:      int(user.ID),
	}, nil
}
