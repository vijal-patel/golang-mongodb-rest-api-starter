package apiutils

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/utils"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GetLimitOffset(c echo.Context) (int64, int64) {
	querylimit := c.QueryParam("limit")
	queryOffset := c.QueryParam("offset")
	limit, err := strconv.ParseInt(querylimit, 10, 64)
	if err != nil {
		limit = 10
	}
	offset, _ := strconv.ParseInt(queryOffset, 10, 64)
	limit = utils.IfTernary(limit > 1000, 1000, limit)
	return limit, offset
}

func GetOrderBySortType(c echo.Context) (string, string) {
	orderBy := c.QueryParam("orderBy")
	sortType := utils.IfTernary[string](c.QueryParam("sort") == constants.SortAsc, constants.SortAsc, constants.SortDesc)
	return orderBy, sortType
}

func Response(c echo.Context, statusCode int, data interface{}) error {
	// nolint // context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	// nolint // context.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	// nolint // context.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization")
	return c.JSON(statusCode, data)
}

func MessageResponse(c echo.Context, statusCode int, message string) error {
	return Response(c, statusCode, models.MessageResponse{
		Message: message,
	})
}

func BadRequestResponse(c echo.Context, message string) error {
	return Response(c, http.StatusBadRequest, models.MessageResponse{
		Message: message,
	})
}

func NotFoundResponse(c echo.Context, message string) error {
	fallBackMessage := "Does not exist"
	return Response(c, http.StatusNotFound, models.MessageResponse{
		Message: utils.IfTernary(message == constants.EmptyString, fallBackMessage, message),
	})
}

func InternalServerErrorResponse(c echo.Context, message string, log *zap.SugaredLogger, err error) error {
	if message == constants.EmptyString {
		message = "Oh no, an error occured. Please try again or contact support"
	}
	errorResponse := models.ErrorResponse{
		Message: message,
	}
	if err != nil {
		log.Errorf("Error:%s Stack trace:%s", err.Error(), debug.Stack())
		errorResponse.Error = err.Error()
	}

	return Response(c, http.StatusInternalServerError, errorResponse)

}

func ValidationErrorResponse(c echo.Context, err error, message string) error {
	if message == constants.EmptyString {
		message = "Required fields are empty or not valid"
	}
	return Response(c, http.StatusBadRequest, models.ErrorResponse{
		Message: message,
		Error:   err.Error(),
	})
}

func PermissionErrorResponse(c echo.Context, log *zap.SugaredLogger) error {
	return Response(c, http.StatusForbidden, models.ErrorResponse{
		Message: "You do not have permissions to perform this action",
	})
}

func GetManyResponse(c echo.Context, hasNext bool, total int64, limit int64, offset int64, items interface{}) error {
	return Response(c, http.StatusOK, models.GetManyResponse{
		Limit:   limit,
		Offset:  offset,
		HasNext: hasNext,
		Total:   total,
		Items:   items,
	})
}

func WriteAccessTokenCookie(c echo.Context, accessToken string, domain string) {
	accessTokenCookie := new(http.Cookie)
	accessTokenCookie.Name = constants.AccessTokenCookieName
	accessTokenCookie.Value = accessToken
	accessTokenCookie.Expires = time.Now().Add(constants.AccessTokenExpiry * time.Hour)
	accessTokenCookie.HttpOnly = true
	accessTokenCookie.Secure = true
	accessTokenCookie.Domain = domain
	accessTokenCookie.Path = "/api/"
	c.SetCookie(accessTokenCookie)
}

func WriteRefreshTokenCookie(c echo.Context, refreshToken string, domain string) {
	refreshTokenCookie := new(http.Cookie)
	refreshTokenCookie.Name = constants.RefreshTokenCookieName
	refreshTokenCookie.Value = refreshToken
	refreshTokenCookie.Expires = time.Now().Add(constants.RefreshTokenExpiry * time.Hour)
	refreshTokenCookie.HttpOnly = true
	refreshTokenCookie.Secure = true
	refreshTokenCookie.Domain = domain
	refreshTokenCookie.Path = "/api/v1/auth/"
	c.SetCookie(refreshTokenCookie)
}

func WriteLogoutCookie(c echo.Context, domain string) {
	accessTokenCookie := new(http.Cookie)
	accessTokenCookie.Name = constants.AccessTokenCookieName
	accessTokenCookie.Value = constants.EmptyString
	accessTokenCookie.Expires = time.Now().Add(1 * time.Second)
	accessTokenCookie.HttpOnly = true
	accessTokenCookie.Secure = true
	accessTokenCookie.Domain = domain
	accessTokenCookie.Path = "/api/"
	c.SetCookie(accessTokenCookie)

	refreshTokenCookie := new(http.Cookie)
	refreshTokenCookie.Name = constants.RefreshTokenCookieName
	refreshTokenCookie.Value = constants.EmptyString
	refreshTokenCookie.Expires = time.Now().Add(1 * time.Second)
	refreshTokenCookie.HttpOnly = true
	refreshTokenCookie.Secure = true
	refreshTokenCookie.Domain = domain
	refreshTokenCookie.Path = "/api/v1/auth/"
	c.SetCookie(refreshTokenCookie)
}
