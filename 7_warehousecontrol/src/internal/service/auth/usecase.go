package auth

import (
	"context"
	"fmt"
	"time"
	"warehousecontrol/src/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	CreateUser(ctx context.Context, role models.Role) (id string, err error)
	GetRole(ctx context.Context, id string) (role models.Role, err error)
}

type Usecase struct {
	userRepo UserRepository
	secret   string
}

func NewUsecase(userRepo UserRepository, secret string) *Usecase {
	return &Usecase{userRepo: userRepo, secret: secret}
}

func (u *Usecase) VerifyJWT(ctx context.Context, tokenString string) (role models.Role, id string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.secret), nil
	})
	if err != nil {
		return 0, "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
		return 0, "", fmt.Errorf("token expired")
	}

	role, err = u.userRepo.GetRole(ctx, claims["id"].(string))
	if err != nil {
		return 0, "", fmt.Errorf("failed to get role: %w", err)
	}

	if role != claims["role"] {
		return 0, "", fmt.Errorf("incorrect role")
	}

	if role != models.RoleUser && role != models.RoleManager && role != models.RoleAdmin {
		return 0, "", fmt.Errorf("incorrect role")
	}

	return role, id, nil
}

func (u *Usecase) CreateJWT(ctx context.Context, role models.Role) (string, error) {
	id, err := u.userRepo.CreateUser(ctx, role)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"role": role,
		"id":   id,
		"exp":  time.Now().Add(time.Hour * 1).Unix(), // expires in 1h
		"iat":  time.Now().Unix(),
		"iss":  "warehousecontrol",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(u.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}
