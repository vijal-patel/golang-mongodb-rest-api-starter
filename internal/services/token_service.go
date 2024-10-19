package services

import (
	"golang-mongodb-rest-api-starter/internal/config"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenServiceWrapper interface {
	CreateAccessToken(user *models.User) (accessToken string, exp int64, err error)
	CreateRefreshToken(user *models.User) (t string, err error)
}

type TokenService struct {
	config *config.Config
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{
		config: cfg,
	}
}

func (tokenService *TokenService) CreateAccessToken(user *models.User) (accessToken string, exp int64, err error) {
	tokenExp := time.Now().Add(time.Hour * constants.AccessTokenExpiry)
	claims := &models.JwtCustomClaims{
		Name:           user.Name,
		UserId:         user.Id,
		Roles:          user.Roles,
		OrganizationId: user.OrganizationId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenExp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(tokenService.config.Auth.AccessSecret))
	if err != nil {
		return constants.EmptyString, 0, err
	}

	return t, tokenExp.Unix(), err
}

func (tokenService *TokenService) CreateRefreshToken(user *models.User) (t string, err error) {

	claims := &models.JwtCustomRefreshClaims{
		UserId: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * constants.RefreshTokenExpiry)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	rt, err := refreshToken.SignedString([]byte(tokenService.config.Auth.RefreshSecret))
	if err != nil {
		return constants.EmptyString, err
	}
	return rt, err
}

func (tokenService *TokenService) GetClaimsFromToken(token *jwt.Token) *models.JwtCustomClaims {
	return token.Claims.(*models.JwtCustomClaims)
}
