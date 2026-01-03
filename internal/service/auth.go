package service

import (
	"context"
	"spender-backend/v2/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo  domain.UserRepository
	JwtSecret []byte
}

func NewAuthService(userRepo domain.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{UserRepo: userRepo, JwtSecret: []byte(jwtSecret)}
}

func (auth *AuthService) Register(ctx context.Context, email, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := domain.User{
		Email:    email,
		Password: string(passwordHash),
	}

	return auth.UserRepo.CreateUser(ctx, &user)
}

// Login Verify and returns the jwt token
func (auth *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := auth.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	// Compare the hashed passwords.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	// Set expiration to 90 days for a "forever-ish" feel
	expirationTime := time.Now().Add(time.Hour * 24 * 90)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"iat": time.Now().Unix(),     // Issued At
		"exp": expirationTime.Unix(), // Expires At
	})

	return token.SignedString(auth.JwtSecret)
}
