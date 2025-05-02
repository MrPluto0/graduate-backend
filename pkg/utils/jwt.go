package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// InitJWTSecret 初始化 JWT 密钥
func InitJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // 新增：标识令牌类型
	jwt.RegisteredClaims
}

// GenerateToken 生成访问令牌
func GenerateToken(userID uint, username, role string) (string, error) {
	return generateTokenWithType(userID, username, role, "access", 24*time.Hour)
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint, username, role string) (string, error) {
	return generateTokenWithType(userID, username, role, "refresh", 7*24*time.Hour)
}

// generateTokenWithType 根据类型生成令牌
func generateTokenWithType(userID uint, username, role, tokenType string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
