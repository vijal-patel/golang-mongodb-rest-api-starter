package models

import "github.com/golang-jwt/jwt/v5"

type JwtCustomClaims struct {
	Name           string   `json:"name"`
	UserId         string   `json:"userId"`
	Roles          []string `json:"roles"`
	OrganizationId string   `json:"organizationId"`
	jwt.RegisteredClaims
}

type JwtCustomRefreshClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}
