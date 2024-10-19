package handlers

import (
	"golang-mongodb-rest-api-starter/internal/captcha"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/logger"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/otp"
	"golang-mongodb-rest-api-starter/internal/services"
	"golang-mongodb-rest-api-starter/internal/utils/apiutils"

	s "golang-mongodb-rest-api-starter/internal/server"

	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

type RegisterHandlerV1 struct {
	server              *s.Server
	userService         *services.UserService
	emailService        *services.EmailService
	tokenService        *services.TokenService
	organizationService *services.OrganizationService
}

func NewRegisterHandlerV1(server *s.Server) *RegisterHandlerV1 {
	return &RegisterHandlerV1{
		server:              server,
		userService:         services.NewUserService(server.DB, server.Config.DB.Name),
		emailService:        services.NewEmailService(server.Mailer, server.Config.Email.ReplyToEmail, server.Config.Email.FromName, server.Config.Email.FromEmail),
		tokenService:        services.NewTokenService(server.Config),
		organizationService: services.NewOrganizationService(server.DB, server.Config.DB.Name),
	}
}

// Register godoc
// @Summary Register
// @Description New user registration
// @ID user-register
// @Tags User Actions
// @Accept json
// @Produce json
// @Param params body models.RegisterRequest true "User's email, user's password"
// @Success 201 {object} models.LoginResponse
// @Failure 400 {object} models.Error
// @Router /users/register [post]
func (h *RegisterHandlerV1) RegisterOrganization(c echo.Context) error {
	req := new(models.RegisterRequest)
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
	var confirmOtp string
	if req.Email != constants.EmptyString {
		confirmOtp = otp.GenerateConfirmOTP()
	}

	user := &models.User{}

	if err := h.userService.GetByEmail(user, req.Email, constants.EmptyString); err != nil {
		apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}
	if user.Id != constants.EmptyString {
		if !user.Confirmed {
			return apiutils.MessageResponse(c, http.StatusBadRequest, "Account already exists, please login")
		}
		return apiutils.MessageResponse(c, http.StatusBadRequest, "User already exists")
	}
	user, err := h.userService.Register(req, confirmOtp)

	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.emailService.SendVerificationCode(req.Email, user.Name, confirmOtp); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	organization := models.Organization{
		Email:       req.Email,
		Name:        req.OrganizationName,
		CreatedById: user.Id,
	}
	if err := h.organizationService.Create(&organization); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	g := new(errgroup.Group)

	// Wait for all db ops to complete.
	if err := g.Wait(); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	updatedUser := &models.UpdateUserRequest{
		OrganizationId: organization.Id,
		Roles:          []string{constants.OrganizationAdminRole},
	}

	if err := h.userService.Update(updatedUser, user.Id, user.Id, constants.EmptyString); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}
	user.OrganizationId = organization.Id
	user.Roles = []string{constants.OrganizationAdminRole}
	accessToken, exp, err := h.tokenService.CreateAccessToken(user)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}
	refreshToken, err := h.tokenService.CreateRefreshToken(user)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)

	}

	apiutils.WriteAccessTokenCookie(c, accessToken, h.server.Config.HTTP.Host)
	apiutils.WriteRefreshTokenCookie(c, refreshToken, h.server.Config.HTTP.Host)

	res := models.NewLoginResponse(accessToken, refreshToken, exp)

	return apiutils.Response(c, http.StatusOK, res)
}
