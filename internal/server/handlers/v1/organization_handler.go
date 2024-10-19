package handlers

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/models"
	s "golang-mongodb-rest-api-starter/internal/server"
	"golang-mongodb-rest-api-starter/internal/services"
	"golang-mongodb-rest-api-starter/internal/utils/apiutils"

	"golang-mongodb-rest-api-starter/internal/logger"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

type OrganizationHandlerV1 struct {
	server              *s.Server
	organizationService *services.OrganizationService
	userService         *services.UserService
}

func NewOrganizationHandlerV1(server *s.Server) *OrganizationHandlerV1 {
	return &OrganizationHandlerV1{
		server:              server,
		organizationService: services.NewOrganizationService(server.DB, server.Config.DB.Name),
		userService:         services.NewUserService(server.DB, server.Config.DB.Name),
	}
}

// DeleteOrganization godoc
// @Summary Delete organization
// @Description Delete organization
// @ID organization-delete
// @Tags Organizations Actions
// @Param id path int true "Organization ID"
// @Success 204 {object} models.Data
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /organizations/me [delete]
func (h *OrganizationHandlerV1) DeleteOrganization(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	id := claims.OrganizationId
	organization := models.Organization{}

	if err := h.organizationService.GetById(&organization, id); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if organization.Id == constants.EmptyString {
		return apiutils.MessageResponse(c, http.StatusNotFound, "Organization not found")
	}

	g := new(errgroup.Group)

	g.Go(func() error {
		return h.userService.DeleteAllForOrganization(id)
	})
	g.Go(func() error {
		return h.organizationService.Delete(id)
	})

	// Wait for all db ops to complete.
	if err := g.Wait(); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.organizationService.Delete(id); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, nil)
	}

	return apiutils.MessageResponse(c, http.StatusNoContent, "Organization deleted successfully")
}

// GetOrganizations godoc
// @Summary Get organization
// @Description Get the list of all organizations
// @ID organizations-get
// @Tags Organizations Actions
// @Produce json
// @Success 200 {array} models.Organization
// @Security ApiKeyAuth
// @Router /organizations [get]
func (h *OrganizationHandlerV1) GetOrganizations(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	hasSuperRole := slices.Contains(claims.Roles, constants.SuperUserRole)

	if !hasSuperRole {
		return apiutils.PermissionErrorResponse(c, log)
	}
	var organizations []models.Organization
	h.organizationService.GetMany(&organizations, nil, 10, 0)
	return apiutils.Response(c, http.StatusOK, organizations)
}

// GetOrganization godoc
// @Summary Get organization
// @Description Get organization
// @ID organization-get
// @Tags Organizations Actions
// @Produce json
// @Success 200 {object} models.Organization
// @Security ApiKeyAuth
// @Router /organizations/me [get]
func (h *OrganizationHandlerV1) GetOrganization(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	id := claims.OrganizationId

	var organization models.Organization
	if err := h.organizationService.GetById(&organization, id); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if organization.Id == constants.EmptyString {
		return apiutils.NotFoundResponse(c, constants.OrganizationNotFound)
	}
	return apiutils.Response(c, http.StatusOK, organization)
}

// UpdateOrganization godoc
// @Summary Update organization
// @Description Update organization
// @ID organization-update
// @Tags Organizations Actions
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Param params body models.UpdateOrganizationRequest true "Organization title and content"
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /organizations/me [patch]
func (h *OrganizationHandlerV1) UpdateOrganization(c echo.Context) error {
	req := new(models.UpdateOrganizationRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	id := claims.OrganizationId

	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	organization := models.Organization{}

	if err := h.organizationService.GetById(&organization, id); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if organization.Id == constants.EmptyString {
		return apiutils.MessageResponse(c, http.StatusNotFound, "Organization not found")
	}

	if err := h.organizationService.Update(organization.Id, req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.MessageResponse(c, http.StatusOK, "Organization successfully updated")
}
