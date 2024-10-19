package handlers

import (
	"golang-mongodb-rest-api-starter/internal/captcha"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/otp"
	s "golang-mongodb-rest-api-starter/internal/server"
	"golang-mongodb-rest-api-starter/internal/services"
	"golang-mongodb-rest-api-starter/internal/utils/apiutils"
	"time"

	"fmt"
	"golang-mongodb-rest-api-starter/internal/logger"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type CustomAuthHandlerV1 struct {
	server       *s.Server
	emailService *services.EmailService
	tokenService *services.TokenService
	userService  *services.UserService
}

func NewCustomAuthHandlerV1(server *s.Server) *CustomAuthHandlerV1 {
	return &CustomAuthHandlerV1{
		server:       server,
		emailService: services.NewEmailService(server.Mailer, server.Config.Email.ReplyToEmail, server.Config.Email.FromName, server.Config.Email.FromEmail),
		tokenService: services.NewTokenService(server.Config),
		userService:  services.NewUserService(server.DB, server.Config.DB.Name),
	}
}

// Login godoc
// @Summary Authenticate a user
// @Description Perform user login
// @ID user-login
// @Tags Auth
// @Accept json
// @Produce json
// @Param params body models.LoginRequest true "User's credentials"
// @Success 200 {object} models.Data
// @Failure 401 {object} models.Error
// @Router /auth/login [post]
func (h *CustomAuthHandlerV1) Login(c echo.Context) error {
	req := new(models.LoginRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)

	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	if err := captcha.VerifyCaptcha(req.CaptchaToken, h.server.Config.Captcha); err != nil {
		log.Infof("captcha validation failed for %s error:%s", req.CaptchaToken, err)
		return apiutils.MessageResponse(c, http.StatusForbidden, "Captcha validation failed")
	}

	user := models.User{}
	if err := h.userService.GetByEmail(&user, req.Email, constants.EmptyString); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}
	var userPassword string
	if user.LoginOtp != constants.EmptyString {
		userPassword = user.LoginOtp
	} else {
		userPassword = user.Password
	}

	if user.Id == constants.EmptyString || (bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(req.Password)) != nil) {
		return apiutils.MessageResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	accessToken, _, err := h.tokenService.CreateAccessToken(&user)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	refreshToken, err := h.tokenService.CreateRefreshToken(&user)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	apiutils.WriteAccessTokenCookie(c, accessToken, h.server.Config.HTTP.Host)
	apiutils.WriteRefreshTokenCookie(c, refreshToken, h.server.Config.HTTP.Host)

	return apiutils.MessageResponse(c, http.StatusOK, "Login success")
}

// Logout godoc
// @Summary Authenticate a user
// @Description Perform user Logout
// @ID user-login
// @Tags Auth
// @Produce json
// @Success 200 {object} models.Data
// @Failure 401 {object} models.Error
// @Router /auth/logout [get]
func (h *CustomAuthHandlerV1) Logout(c echo.Context) error {
	apiutils.WriteLogoutCookie(c, h.server.Config.HTTP.Host)
	return apiutils.MessageResponse(c, http.StatusOK, "Logged out")
}

// Refresh godoc
// @Summary Refresh access token
// @Description Perform refresh access token
// @ID user-refresh
// @Tags Auth
// @Accept json
// @Produce json
// @Param params body models.RefreshRequest true "Refresh token"
// @Success 200 {object} models.Data
// @Failure 401 {object} models.Error
// @Router /auth/v1/refresh [get]
func (h *CustomAuthHandlerV1) RefreshToken(c echo.Context) error {
	req := new(models.RefreshRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}
	refreshTokenCookie, err := c.Cookie(constants.RefreshTokenCookieName)
	if err != nil {
		return apiutils.BadRequestResponse(c, "Missing Cookie")
	}
	token, err := jwt.ParseWithClaims(refreshTokenCookie.Value, &models.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// secret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(h.server.Config.Auth.RefreshSecret), nil
	})

	if err != nil {
		return apiutils.MessageResponse(c, http.StatusUnauthorized, "Invalid token")
	}

	claims, ok := token.Claims.(*models.JwtCustomClaims)
	if !ok || !token.Valid {
		return apiutils.MessageResponse(c, http.StatusUnauthorized, "Invalid token")
	}

	user := new(models.User)
	if err := h.userService.GetByIdAnyOrg(user, claims.UserId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if user.Id == constants.EmptyString {
		return apiutils.MessageResponse(c, http.StatusNotFound, "User not found")
	}

	accessToken, _, err := h.tokenService.CreateAccessToken(user)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	refreshToken, err := h.tokenService.CreateRefreshToken(user)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	apiutils.WriteAccessTokenCookie(c, accessToken, h.server.Config.HTTP.Host)
	apiutils.WriteRefreshTokenCookie(c, refreshToken, h.server.Config.HTTP.Host) // TODO replace with subdomain for prod

	return apiutils.MessageResponse(c, http.StatusNoContent, "Token refreshed")
}

// Refresh godoc
// @Summary Validate OTP
// @Description Validate OTP
// @ID validate-otp
// @Tags Auth
// @Accept json
// @Produce json
// @Param params body models.OtpRequest true "OTP"
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Router /auth/otp/validate [post]
func (h *CustomAuthHandlerV1) ValidateOtp(c echo.Context) error {
	req := new(models.OtpRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	claims := h.tokenService.GetClaimsFromToken(c.Get("user").(*jwt.Token))

	isValid, err := h.userService.ValidateConfirmOtp(claims.UserId, req.Otp)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, "Please request a new OTP", log, err)
	}
	if !isValid {
		return apiutils.BadRequestResponse(c, "Invalid OTP")
	}

	return apiutils.MessageResponse(c, http.StatusOK, "ok")
}

// Refresh godoc
// @Summary Send OTP
// @Description Send OTP
// @ID send-otp
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Router /auth/otp/send [post]
func (h *CustomAuthHandlerV1) SendOtp(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)

	claims := h.tokenService.GetClaimsFromToken(c.Get("user").(*jwt.Token))

	user := new(models.User)

	if err := h.userService.GetById(user, claims.UserId, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if user.Id == constants.EmptyString {
		return apiutils.NotFoundResponse(c, constants.UserNotFound)
	}

	if user.LastConfirmOtpAt != nil {
		nowMilli := time.Now().UnixMilli()
		if nowMilli-user.LastConfirmOtpAt.UnixMilli() < constants.UserConfirmOtpIntervalMilli {
			return apiutils.BadRequestResponse(c, "Please wait 1 minute before requesting another OTP")

		}
	}

	otp := otp.GenerateConfirmOTP()

	if err := h.userService.UpdateConfirmOtp(claims.UserId, otp); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.emailService.SendVerificationCode(user.Email, user.Name, otp); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.MessageResponse(c, http.StatusOK, "ok")
}

// Password change godoc
// @Summary change password
// @Description Perform user password change
// @ID user-password-change
// @Tags Auth
// @Accept json
// @Produce json
// @Param params body models.PasswordChangeRequest true "User's credentials"
// @Success 200 {object} models.Data
// @Failure 401 {object} models.Error
// @Router /auth/password [post]
func (h *CustomAuthHandlerV1) PasswordChange(c echo.Context) error {
	req := new(models.PasswordChangeRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	user := models.User{}
	if err := h.userService.GetById(&user, claims.UserId, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}
	if user.Id == constants.EmptyString || (bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)) != nil) {
		return apiutils.MessageResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	if err := h.userService.UpdatePassword(claims.UserId, req.NewPassword); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.MessageResponse(c, http.StatusOK, constants.UserPasswordUpdated)
}

// Password recover godoc
// @Summary recover password
// @Description Perform user password recover
// @ID user-password-recover
// @Tags Auth
// @Accept json
// @Produce json
// @Param params body models.PasswordRecoverRequest true "User's credentials"
// @Success 200 {object} models.MessageResponse
// @Failure 401 {object} models.Error
// @Router /auth/password/reset [post]
func (h *CustomAuthHandlerV1) PasswordRecover(c echo.Context) error {
	req := new(models.PasswordRecoverRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)

	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	if err := captcha.VerifyCaptcha(req.CaptchaToken, h.server.Config.Captcha); err != nil {
		log.Infof("captcha validation failed for %s error:%s", req.CaptchaToken, err)
		return apiutils.MessageResponse(c, http.StatusForbidden, "Captcha validation failed")
	}

	user := models.User{}
	if err := h.userService.GetByEmailAnyOrg(&user, req.Email); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if user.Id != constants.EmptyString {
		otp := otp.GenerateConfirmOTP()
		if err := h.userService.UpdateConfirmOtp(user.Id, otp); err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
		if err := h.emailService.SendPasswordResetCode(user.Email, user.Name, otp); err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
	}

	return apiutils.MessageResponse(c, http.StatusOK, constants.UserPasswordRecoverResponse)
}

// Password reset godoc
// @Summary reset password
// @Description Perform user password reset
// @ID user-password-reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param params body models.PasswordResetRequest true "User's credentials"
// @Success 200 {object} models.MessageResponse
// @Failure 401 {object} models.Error
// @Router /auth/password/reset [post]
func (h *CustomAuthHandlerV1) PasswordReset(c echo.Context) error {
	req := new(models.PasswordResetRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)

	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	user := models.User{}
	if err := h.userService.GetByEmailAnyOrg(&user, req.Email); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if user.Id != constants.EmptyString {
		isValid, err := h.userService.ValidateConfirmOtp(user.Id, req.Otp)
		if err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
		if !isValid {
			return apiutils.BadRequestResponse(c, constants.InvalidOTPError)
		}
		if err := h.userService.UpdatePassword(user.Id, req.NewPassword); err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
		h.userService.RevokeLoginOTP(user.Id)
		return apiutils.MessageResponse(c, http.StatusOK, constants.UserPasswordUpdated)
	}

	return apiutils.BadRequestResponse(c, constants.InvalidOTPError)
}
