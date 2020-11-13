package jwt_tool

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenInfo struct {
	Token     string
	ExpiredAt int64
}

func CreateToken(claims jwt.MapClaims, expireSeconds int) (*TokenInfo, error) {
	// 创建token
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		return nil, errors.New("invalid token secret")
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expireDuration := time.Duration(expireSeconds) * time.Second
	claims["exp"] = time.Now().Add(expireDuration).Unix()
	token, err := tokenObj.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}
	return &TokenInfo{
		Token:     token,
		ExpiredAt: claims["exp"].(int64),
	}, nil
}

func verifyToken(token string) (*jwt.Token, error) {
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		return nil, errors.New("invalid token secret")
	}

	tokenObj, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return tokenObj, nil
}

func TokenClaims(token string) (jwt.MapClaims, error) {
	// 获取token claims
	tokenObj, err := verifyToken(token)
	if err != nil {
		return nil, err
	}
	var claims jwt.MapClaims
	var ok bool
	claims, ok = tokenObj.Claims.(jwt.MapClaims)
	if ok && tokenObj.Valid {
		return claims, nil
	}
	return nil, errors.New("get token claims failed")
}
