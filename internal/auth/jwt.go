package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims は JWT に埋め込むカスタムクレーム。
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenGenerator は JWT の生成と検証を担う。
type TokenGenerator struct {
	secret      []byte
	expiryHours int
}

func NewTokenGenerator(secret string, expiryHours int) *TokenGenerator {
	return &TokenGenerator{secret: []byte(secret), expiryHours: expiryHours}
}

// GenerateToken は userID と role を含む署名済み JWT を生成する。
func (g *TokenGenerator) GenerateToken(userID, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(g.expiryHours) * time.Hour)
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(g.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return tokenString, expiresAt, nil
}

// ValidateToken はトークン文字列を検証し、Claims を返す。
// 期限切れ・署名不正・改ざん・none アルゴリズム攻撃を全てエラーとして返す。
func (g *TokenGenerator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
