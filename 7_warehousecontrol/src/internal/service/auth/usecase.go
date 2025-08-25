package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	CreateUser(ctx context.Context, role string) (id string, err error)
	GetRole(ctx context.Context, id string) (role string, err error)
}

type Usecase struct {
	userRepo UserRepository
	secret   string
}

func NewUsecase(userRepo UserRepository, secret string) *Usecase {
	return &Usecase{userRepo: userRepo, secret: secret}
}

func (u *Usecase) VerifyJWT(ctx context.Context, tokenString string) (role string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
		return "", fmt.Errorf("token expired")
	}

	role, err = u.userRepo.GetRole(ctx, claims["id"].(string))
	if err != nil {
		return "", fmt.Errorf("failed to get role: %w", err)
	}

	if role != claims["role"] {
		return "", fmt.Errorf("incorrect role")
	}

	return role, nil
}

func (u *Usecase) CreateJWT(ctx context.Context, role string) (string, error) {

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
