package service

import (
	"errors"
	"fmt"

	"github.com/mino/backend/internal/config"
	jwtpkg "github.com/mino/backend/internal/pkg/jwt"
	"github.com/mino/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtMgr   *jwtpkg.Manager
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, jwtMgr *jwtpkg.Manager, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, jwtMgr: jwtMgr, cfg: cfg}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) SignIn(username, password string) (*TokenPair, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.jwtMgr.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtMgr.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthService) Refresh(refreshToken string) (string, error) {
	claims, err := s.jwtMgr.Validate(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	if claims.TokenType != "refresh" {
		return "", errors.New("invalid token type: expected refresh token")
	}
	return s.jwtMgr.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
}

func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(userID, string(hash))
}

func (s *AuthService) ChangeUsername(userID, newUsername, password string) error {
	if len(newUsername) < 2 {
		return errors.New("username must be at least 2 characters")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Verify password before allowing username change
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid password")
	}

	// Check if new username is already taken
	existing, err := s.userRepo.FindByUsername(newUsername)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if existing != nil && existing.ID != userID {
		return errors.New("username already taken")
	}

	return s.userRepo.UpdateUsername(userID, newUsername)
}

// EnsureAdminUser creates the default admin user if it doesn't exist.
func (s *AuthService) EnsureAdminUser() error {
	user, err := s.userRepo.FindByUsername(s.cfg.Admin.Username)
	if err != nil {
		return err
	}
	if user != nil {
		return nil // already exists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.Admin.Password), 12)
	if err != nil {
		return err
	}

	role := "admin"
	return s.userRepo.CreateFromInput(&repository.UserCreate{
		Username:     s.cfg.Admin.Username,
		PasswordHash: string(hash),
		Role:         role,
	})
}
