package middlware

import (
	"fmt"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/services"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const contextKey = "user"

func CustomAuthMiddleware(secret string, enforcer *casbin.Enforcer, authzService *services.AuthzService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			accessTokenCookie, err := c.Cookie(constants.AccessTokenCookieName)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing authentication information")
			}
			// Extract the token without the "Bearer " prefix.
			token, err := jwt.ParseWithClaims(accessTokenCookie.Value, &models.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				// secret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return []byte(secret), nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			// Check if token is valid and claims contain the required role.
			claims, ok := token.Claims.(*models.JwtCustomClaims)
			if ok && token.Valid {
				// ctx := context.WithValue(c.Request().Context(), "userId", claims.UserId)
				// c.SetRequest(c.Request().WithContext(ctx))
				// Get the "role" claim from the JWT token.
				if claims.Roles == nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "JWT token missing 'role' claim")

				}
				requestUrl := c.Request().URL.String()
				// Check if the user's role matches one of the required roles.
				for _, role := range claims.Roles {
					// enforcer.get
					// fmt.Println(role, c.Request().URL.String(), authzService.MapHttpMethodToAction(c.Request().Method))
					isAuthorized, err := enforcer.Enforce(role, requestUrl, authzService.MapHttpMethodToAction(c.Request().Method))
					if err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Error authorizing user")
					}

					if isAuthorized {
						c.Set(contextKey, token)
						return next(c)
					}

				}

				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
		}
	}
}
