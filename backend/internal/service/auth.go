package service

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/MaximovIlya/vk_practice_project/internal/auth"
	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailTaken        = errors.New("email already taken")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrTokenExpired      = errors.New("token expired or invalid")
)

type AuthService struct {
	store   *repository.Store
	manager *auth.Manager
}

func NewAuthService(store *repository.Store, manager *auth.Manager) *AuthService {
	return &AuthService{store: store, manager: manager}
}

type RegisterParams struct {
	Email    string
	Password string
	Name     string
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (s *AuthService) Register(ctx context.Context, p RegisterParams) (TokenPair, error) {
	existing, err := s.store.GetUserByEmail(ctx, p.Email)
	if err == nil && existing.ID != "" {
		return TokenPair{}, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), 12)
	if err != nil {
		return TokenPair{}, err
	}
	hashStr := string(hash)

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		ID:       ulid.Make().String(),
		Email:    p.Email,
		Password: &hashStr,
		Name:     p.Name,
		Role:     db.RolePARTICIPANT,
	})
	if err != nil {
		return TokenPair{}, err
	}

	return s.issueTokens(user.ID, user.Role)
}

type LoginParams struct {
	Email    string
	Password string
}

func (s *AuthService) Login(ctx context.Context, p LoginParams) (TokenPair, error) {
	user, err := s.store.GetUserByEmail(ctx, p.Email)
	if err != nil {
		return TokenPair{}, ErrUserNotFound
	}

	if user.Password == nil {
		return TokenPair{}, ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(p.Password)); err != nil {
		return TokenPair{}, ErrInvalidPassword
	}

	return s.issueTokens(user.ID, user.Role)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.manager.Parse(refreshToken)
	if err != nil {
		return TokenPair{}, ErrTokenExpired
	}

	if claims.Type != auth.RefreshToken {
		return TokenPair{}, ErrTokenExpired
	}

	user, err := s.store.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, ErrUserNotFound
	}

	return s.issueTokens(user.ID, user.Role)
}

func (s *AuthService) SelectRole(ctx context.Context, userID string, role db.Role) error {
	_, err := s.store.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   userID,
		Role: role,
	})
	return err
}

type ForgotPasswordResult struct {
	Token string
	Email string
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (*ForgotPasswordResult, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		// не раскрываем что email не существует
		return nil, nil
	}

	_ = s.store.DeletePasswordResetTokensByEmail(ctx, email)

	token := ulid.Make().String()
	_, err = s.store.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		ID:        ulid.Make().String(),
		Email:     user.Email,
		Token:     token,
		ExpiresAt: mustTimestamptz(time.Now().Add(1 * time.Hour)),
	})
	if err != nil {
		return nil, err
	}

	return &ForgotPasswordResult{Token: token, Email: user.Email}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	row, err := s.store.GetPasswordResetToken(ctx, token)
	if err != nil {
		return ErrTokenExpired
	}

	if !row.ExpiresAt.Valid || row.ExpiresAt.Time.Before(time.Now()) {
		_ = s.store.DeletePasswordResetToken(ctx, token)
		return ErrTokenExpired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	hashStr := string(hash)

	user, err := s.store.GetUserByEmail(ctx, row.Email)
	if err != nil {
		return ErrUserNotFound
	}

	// обновляем пароль напрямую через sql
	_, err = s.store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: &hashStr,
	})
	if err != nil {
		return err
	}

	return s.store.DeletePasswordResetToken(ctx, token)
}

func (s *AuthService) issueTokens(userID string, role db.Role) (TokenPair, error) {
	access, err := s.manager.GenerateAccess(userID, role)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := s.manager.GenerateRefresh(userID, role)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}
