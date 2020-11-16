package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/milad-abbasi/oauth-server/pkg/common"
	"github.com/milad-abbasi/oauth-server/pkg/user"
	"go.uber.org/zap"
)

type Service struct {
	l  *zap.Logger
	us *user.Service
}

func NewService(logger *zap.Logger, us *user.Service) *Service {
	return &Service{
		l:  logger.Named("AuthService"),
		us: us,
	}
}

func (s *Service) Register(ctx context.Context, ri *RegisterInfo) (*Tokens, error) {
	u, err := s.us.NewUser(
		ctx,
		&user.TinyUser{
			Name:     ri.Name,
			Email:    ri.Email,
			Password: ri.Password,
		},
	)
	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(u.Id)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) Login(ctx context.Context, li *LoginInfo) (*Tokens, error) {
	u, err := s.us.UserRepo.FindOne(
		ctx,
		&user.User{Email: li.Email},
	)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	ok, err := user.CompareHash(li.Password, u.Password)
	if !ok || err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(u.Id)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) generateTokens(subject string) (*Tokens, error) {
	secret := common.MustGetEnv("TOKEN_SECRET")
	issuer := common.GetEnvWithDefault("TOKEN_ISSUER", "OAuth-server")
	ate, err := strconv.Atoi(
		common.GetEnvWithDefault(
			"ACCESS_TOKEN_EXPIRY",
			"10"),
	)
	rte, err := strconv.Atoi(
		common.GetEnvWithDefault(
			"REFRESH_TOKEN_EXPIRY",
			"100"),
	)
	if err != nil {
		panic(err)
	}

	at := Token{
		ID:            uuid.New().String(),
		Issuer:        issuer,
		Subject:       subject,
		Audience:      []string{},
		Expiry:        time.Hour * time.Duration(ate),
		NotBefore:     time.Now(),
		IssuedAt:      time.Now(),
		PrivateClaims: []interface{}{},
	}
	rt := Token{
		ID:            uuid.New().String(),
		Issuer:        issuer,
		Subject:       subject,
		Audience:      []string{},
		Expiry:        time.Hour * time.Duration(rte),
		NotBefore:     time.Now(),
		IssuedAt:      time.Now(),
		PrivateClaims: []interface{}{},
	}

	sat, err := at.Sign(secret)
	srt, err := rt.Sign(secret)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		TokenType:    "Bearer",
		ExpiresIn:    int((time.Hour * time.Duration(ate)).Seconds()),
		Scope:        "read write",
		AccessToken:  sat,
		RefreshToken: srt,
	}, nil
}
