package services

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"net/http"
)

type AuthzService struct {
}

func NewAuthzService() *AuthzService {
	return &AuthzService{}
}

func (a *AuthzService) MapHttpMethodToAction(httpMethod string) string {
	switch httpMethod {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPatch:
		return "write"
	case http.MethodPut:
		return "write"
	case http.MethodDelete:
		return "delete"
	default:
		return constants.EmptyString
	}
}
