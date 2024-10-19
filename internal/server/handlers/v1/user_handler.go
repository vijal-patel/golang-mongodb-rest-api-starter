package handlers

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/logger"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/otp"
	s "golang-mongodb-rest-api-starter/internal/server"
	"golang-mongodb-rest-api-starter/internal/services"
	"golang-mongodb-rest-api-starter/internal/utils/apiutils"
	"math"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

type UserHandlerV1 struct {
	server              *s.Server
	userService         *services.UserService
	tokenService        *services.TokenService
	emailService        *services.EmailService
	organizationService *services.OrganizationService
}

func NewUserHandlerV1(server *s.Server) *UserHandlerV1 {
	return &UserHandlerV1{server: server,
		userService:         services.NewUserService(server.DB, server.Config.DB.Name),
		tokenService:        services.NewTokenService(server.Config),
		emailService:        services.NewEmailService(server.Mailer, server.Config.Email.ReplyToEmail, server.Config.Email.FromName, server.Config.Email.FromEmail),
		organizationService: services.NewOrganizationService(server.DB, server.Config.DB.Name),
	}
}

// CreateUser godoc
// @Summary Create user
// @Description Create user
// @ID user-create
// @Tags Users Actions
// @Accept json
// @Produce json
// @Param params body models.CreateUserRequest true "User content"
// @Success 200 {object} models.User
// @Failure 400 {object} models.Error
// @Security ApiKeyAuth
// @Router /users [post]
func (h *UserHandlerV1) CreateUser(c echo.Context) error {
	req := new(models.CreateUserRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	newUser := &models.User{}
	organization := models.Organization{}

	g := new(errgroup.Group)
	g.Go(func() error {
		return h.organizationService.GetById(&organization, claims.OrganizationId)
	})
	g.Go(func() error {
		return h.userService.GetByEmailAnyOrg(newUser, req.Email)
	})

	// Wait for all db ops to complete.
	if err := g.Wait(); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if newUser.Id != constants.EmptyString {
		return apiutils.BadRequestResponse(c, "A user with this email address already exists")
	}

	loginOtp := otp.GenerateLoginOTP()
	hashedLoginOtp, err := bcrypt.GenerateFromPassword(
		[]byte(loginOtp),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	newUser, err = h.userService.CreateFromInvite(req, claims.UserId, claims.OrganizationId, string(hashedLoginOtp))
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.emailService.SendInvite(newUser.Email, claims.Name, newUser.Name, loginOtp); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.Response(c, http.StatusOK, newUser)
}

// BulkCreateUsers godoc
// @Summary Bulk Create user
// @Description Bulk Create user
// @ID bulk-user-create
// @Tags Users Actions
// @Accept json
// @Produce json
// @Param params body models.BulkCreateUsersRequest true "Users"
// @Success 200 {object} models.User
// @Failure 400 {object} models.Error
// @Security ApiKeyAuth
// @Router /users/bulk [post]
func (h *UserHandlerV1) BulkCreateUsers(c echo.Context) error {
	req := new(models.BulkCreateUsersRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	if len(req.Users) > 50 {
		return apiutils.BadRequestResponse(c, "Cannot create more than 50 users at once")
	}

	organization := models.Organization{}
	users := []models.User{}
	emails := []string{}
	for _, user := range req.Users {
		emails = append(emails, user.Email)
	}
	var emailsLimit, emailsOffset, emailsTotal int64
	emailsLimit = math.MaxInt64

	g := new(errgroup.Group)
	g.Go(func() error {
		return h.organizationService.GetById(&organization, claims.OrganizationId)
	})

	g.Go(func() error {
		_, total, err := h.userService.GetByEmailsAnyOrg(&users, emails, emailsLimit, emailsOffset)
		emailsTotal = total
		return err
	})

	// Wait for all db ops to complete.
	if err := g.Wait(); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if emailsTotal != 0 {
		var errorMessage strings.Builder
		errorMessage.WriteString("Failed to create new users as the following emails already exist: ")
		for i := range users {
			errorMessage.WriteString(users[i].Email)
			errorMessage.WriteString(" ")
		}
		return apiutils.BadRequestResponse(c, errorMessage.String())
	}

	for i := range req.Users {
		emails = append(emails, req.Users[i].Email)
		loginOtp := otp.GenerateLoginOTP()
		hashedLoginOtp, err := bcrypt.GenerateFromPassword(
			[]byte(loginOtp),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
		req.Users[i].HashedLoginOtp = string(hashedLoginOtp)
		req.Users[i].LoginOtp = loginOtp
	}

	if err := h.userService.BulkCreateFromInvite(req, claims.UserId, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.emailService.BulkSendInvite(claims.Name, req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.MessageResponse(c, http.StatusOK, "Users Invited")
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user
// @ID user-update
// @Tags Users Actions
// @Accept json
// @Produce json
// @Param params body models.UpdateUserRequest true "User content"
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Security ApiKeyAuth
// @Router /users [patch]
func (h *UserHandlerV1) UpdateUser(c echo.Context) error {
	req := new(models.UpdateUserRequest)
	id := c.Param("id")
	log := logger.GetLoggerForContext(c, h.server.Log)
	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	user := &models.User{}
	createNewTokens := false
	if err := h.userService.GetById(user, id, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if user.Id == constants.EmptyString {
		return apiutils.NotFoundResponse(c, constants.UserNotFound)
	}

	var confirmOtp string
	if req.Email != constants.EmptyString {
		confirmOtp = otp.GenerateConfirmOTP()
		req.ConfirmOtp = confirmOtp
	}

	if err := h.userService.Update(req, id, claims.UserId, confirmOtp); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if req.Name != constants.EmptyString && user.Name != req.Name {
		g := new(errgroup.Group)
		// Wait for all db ops to complete.
		if err := g.Wait(); err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
		createNewTokens = true
	}

	if req.Email != constants.EmptyString {
		if err := h.emailService.SendVerificationCode(req.Email, user.Name, confirmOtp); err != nil {
			return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
		}
	}

	if req.Roles != nil {
		createNewTokens = true
	}

	if createNewTokens {
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

	}

	return apiutils.MessageResponse(c, http.StatusOK, constants.UserUpdated)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user
// @ID user-delete
// @Tags Users Actions
// @Accept json
// @Produce json
// @Success 204 {object} models.Data
// @Failure 400 {object} models.Error
// @Security ApiKeyAuth
// @Router /users/{id} [delete]
func (h *UserHandlerV1) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	userExists, err := h.userService.DoesExist(id, claims.OrganizationId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if !userExists {
		return apiutils.NotFoundResponse(c, constants.UserNotFound)
	}

	orgAdminCount, err := h.userService.GetTotalOrgAdminsInOrg(claims.OrganizationId, claims.UserId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if slices.Contains(claims.Roles, constants.OrganizationAdminRole) && orgAdminCount == 1 && claims.UserId == id {
		return apiutils.BadRequestResponse(c, "You are the only organization admin left, please delete the organization instead")
	}

	err = h.userService.Delete(id)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.Response(c, http.StatusNoContent, constants.UserDeleted)
}

// DeleteMyUser godoc
// @Summary Delete logged in user
// @Description Delete logged in user
// @ID user-delete
// @Tags Users Actions
// @Accept json
// @Produce json
// @Success 204 {object} models.Data
// @Failure 400 {object} models.Error
// @Security ApiKeyAuth
// @Router /users/me [delete]
func (h *UserHandlerV1) DeleteMyUser(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	userExists, err := h.userService.DoesExist(claims.UserId, claims.OrganizationId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if !userExists {
		return apiutils.NotFoundResponse(c, constants.UserNotFound)
	}

	orgAdminCount, err := h.userService.GetTotalOrgAdminsInOrg(claims.OrganizationId, claims.UserId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if slices.Contains(claims.Roles, constants.OrganizationAdminRole) && orgAdminCount == 1 {
		return apiutils.BadRequestResponse(c, "You are the only organization admin left, please delete the organization instead")
	}

	err = h.userService.Delete(claims.UserId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.Response(c, http.StatusNoContent, "User deleted successfully")
}

// GetUsers godoc
// @Summary Get user
// @Description Get the list of all users for organization
// @ID users_get
// @Tags User Actions
// @Produce json
// @Success 200 {array} models.User
// @Failure 400 {object} models.Error
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /users [get]
func (h *UserHandlerV1) GetUsers(c echo.Context) error {
	// TODO add more permissions here
	users := []models.User{}
	var hasNext bool
	var total int64
	var err error
	log := logger.GetLoggerForContext(c, h.server.Log)
	limit, offset := apiutils.GetLimitOffset(c)
	orderBy, sortType := apiutils.GetOrderBySortType(c)
	search := strings.ToLower(c.QueryParam("search"))
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	hasNext, total, err = h.userService.GetMany(&users, claims.OrganizationId, models.GetManyQuery{
		Limit:    limit,
		Offset:   offset,
		OrderBy:  orderBy,
		SortType: sortType,
	}, search)

	if err != nil {
		log.Error(err)
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.GetManyResponse(c, hasNext, total, limit, offset, users)
}

// GetUser godoc
// @Summary Get user
// @Description Get user by ID
// @ID user_get
// @Tags User Actions
// @Produce json
// @Success 200 {object} models.User
// @Failure 400 {object} models.Error
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /users/:id [get]
func (h *UserHandlerV1) GetUser(c echo.Context) error {
	id := c.Param("id")
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	user := &models.User{}
	currentUser := &models.User{}

	g := new(errgroup.Group)
	g.Go(func() error {
		return h.userService.GetById(currentUser, claims.UserId, claims.OrganizationId)
	})
	g.Go(func() error {
		return h.userService.GetById(user, id, claims.OrganizationId)
	})

	// Wait for all db ops to complete.
	if err := g.Wait(); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if user.Id == constants.EmptyString {
		return apiutils.NotFoundResponse(c, constants.UserNotFound)
	}

	return apiutils.Response(c, http.StatusOK, user)
}

// GetMyUser godoc
// @Summary Get user belonging to token
// @Description Get user by ID from token
// @ID user_get_token
// @Tags User Actions
// @Produce json
// @Success 200 {object} models.User
// @Failure 400 {object} models.Error
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /users/me [get]
func (h *UserHandlerV1) GetMyUser(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	currentUser := &models.User{}
	err := h.userService.GetByIdAnyOrg(currentUser, claims.UserId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if currentUser.Id == constants.EmptyString {
		return apiutils.NotFoundResponse(c, constants.UserNotFound)
	}
	return apiutils.Response(c, http.StatusOK, currentUser)
}

// Roles godoc
// @Summary User Roles
// @Description User Roles
// @ID user-roles
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Router /user/roles [get]
func (h *MetaHandlerV1) Roles(c echo.Context) error {
	return apiutils.Response(c, http.StatusOK, constants.GetAllowedRoles())
}
