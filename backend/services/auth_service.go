package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"tinystock/backend/models"
	"tinystock/backend/repository"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid email or password")
)

// AuthService handles authentication
type AuthService struct {
	repo   repository.UserRepository
	secret []byte
	expiry time.Duration
}

// NewAuthService creates a new AuthService
func NewAuthService(repo repository.UserRepository, secret string, expiry time.Duration) *AuthService {
	return &AuthService{
		repo:   repo,
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}
	user := &models.User{Email: req.Email, PasswordHash: req.Password}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// Login validates credentials and returns JWT
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return &models.LoginResponse{Token: token, User: *user}, nil
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken parses and validates a JWT, returns user ID
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return "", errors.New("invalid token")
}
