package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService gerencia tokens JWT
type JWTService struct {
	secretKey        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// Claims representa as claims do JWT
type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	MerchantID *uuid.UUID `json:"merchant_id,omitempty"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair representa um par de tokens (access + refresh)
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewJWTService cria um novo serviço JWT
func NewJWTService(secretKey []byte, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       secretKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateTokenPair gera um par de tokens (access + refresh)
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, merchantID *uuid.UUID, email, role string) (*TokenPair, error) {
	// Access Token
	accessToken, expiresAt, err := s.GenerateAccessToken(userID, merchantID, email, role)
	if err != nil {
		return nil, err
	}
	
	// Refresh Token
	refreshToken, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}
	
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		ExpiresAt:    expiresAt,
	}, nil
}

// GenerateAccessToken gera um token de acesso
func (s *JWTService) GenerateAccessToken(userID uuid.UUID, merchantID *uuid.UUID, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTokenTTL)
	
	claims := &Claims{
		UserID:     userID,
		MerchantID: merchantID,
		Email:      email,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pixsaas",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}
	
	return tokenString, expiresAt, nil
}

// GenerateRefreshToken gera um token de refresh
func (s *JWTService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	expiresAt := time.Now().Add(s.refreshTokenTTL)
	
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "pixsaas",
		Subject:   userID.String(),
		ID:        uuid.New().String(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken valida um token JWT e retorna as claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, errors.New("invalid token")
}

// ValidateRefreshToken valida um refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})
	
	if err != nil {
		return uuid.Nil, err
	}
	
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	}
	
	return uuid.Nil, errors.New("invalid refresh token")
}

// ExtractTokenFromHeader extrai o token do header Authorization
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}
	
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}
	
	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with Bearer")
	}
	
	return authHeader[len(bearerPrefix):], nil
}
