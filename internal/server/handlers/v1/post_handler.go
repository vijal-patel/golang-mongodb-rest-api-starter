package handlers

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/logger"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/services"
	"golang-mongodb-rest-api-starter/internal/utils/apiutils"
	"strings"

	s "golang-mongodb-rest-api-starter/internal/server"

	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type PostHandlerV1 struct {
	server      *s.Server
	postService *services.PostService
}

func NewPostHandlerV1(server *s.Server) *PostHandlerV1 {
	return &PostHandlerV1{
		server:      server,
		postService: services.NewPostService(server.DB, server.Config.DB.Name),
	}
}

// CreatePost godoc
// @Summary Create post
// @Description Create post
// @ID post-create
// @Tags Posts Actions
// @Accept json
// @Produce json
// @Param params body models.CreatePostRequest true "Post title and content"
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Security ApiKeyAuth
// @Router /pposts [get]
func (h *PostHandlerV1) CreatePost(c echo.Context) error {
	req := new(models.CreatePostRequest)
	log := logger.GetLoggerForContext(c, h.server.Log)
	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	post, err := h.postService.Create(req, claims.OrganizationId, claims.UserId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.Response(c, http.StatusOK, post)
}

// DeletePost godoc
// @Summary Delete post
// @Description Delete post
// @ID post-delete
// @Tags Posts Actions
// @Param id path int true "Post ID"
// @Param params body models.DeletePostRequest true "New post ID if post has members"
// @Success 204 {object} models.Data
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /pposts/{id} [delete]
func (h *PostHandlerV1) DeletePost(c echo.Context) error {
	id := c.Param("id")
	log := logger.GetLoggerForContext(c, h.server.Log)

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)
	post := models.Post{}

	if err := h.postService.GetById(&post, id, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.PostNotFound, log, err)
	}

	if post.Id == constants.EmptyString {
		return apiutils.MessageResponse(c, http.StatusNotFound, constants.PostNotFound)
	}

	if err := h.postService.Delete(id); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.MessageResponse(c, http.StatusNoContent, constants.PostDeleted)
}

// GetPosts godoc
// @Summary Get post
// @Description Get the list of all pposts
// @ID pposts-get
// @Tags Posts Actions
// @Produce json
// @Success 200 {array} models.Post
// @Security ApiKeyAuth
// @Router /pposts [get]
func (h *PostHandlerV1) GetPosts(c echo.Context) error {
	log := logger.GetLoggerForContext(c, h.server.Log)
	limit, offset := apiutils.GetLimitOffset(c)
	orderBy, sortType := apiutils.GetOrderBySortType(c)
	pposts := []models.Post{}
	token := c.Get("user").(*jwt.Token)
	search := strings.ToLower(c.QueryParam("search"))
	claims := token.Claims.(*models.JwtCustomClaims)
	organizationId := claims.OrganizationId

	hasNext, total, err := h.postService.GetMany(&pposts, organizationId, models.GetManyQuery{
		Limit:    limit,
		Offset:   offset,
		SortType: sortType,
		OrderBy:  orderBy,
	}, search)
	if err != nil {
		log.Error(err)
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.GetManyResponse(c, hasNext, total, limit, offset, pposts)
}

// GetPost godoc
// @Summary Get post
// @Description Get post
// @ID post-get
// @Tags Posts Actions
// @Produce json
// @Success 200 {object} models.Post
// @Security ApiKeyAuth
// @Router /pposts/{id} [get]
func (h *PostHandlerV1) GetPost(c echo.Context) error {
	id := c.Param("id")
	log := logger.GetLoggerForContext(c, h.server.Log)
	var post models.Post
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	if err := h.postService.GetById(&post, id, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if post.Id == constants.EmptyString {
		return apiutils.NotFoundResponse(c, constants.PostNotFound)
	}
	return apiutils.Response(c, http.StatusOK, post)
}

// UpdatePost godoc
// @Summary Update post
// @Description Update post
// @ID post-update
// @Tags Posts Actions
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param params body models.UpdatePostRequest true "Post title and content"
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Failure 404 {object} models.Error
// @Security ApiKeyAuth
// @Router /pposts/{id} [put]
func (h *PostHandlerV1) UpdatePost(c echo.Context) error {
	req := new(models.UpdatePostRequest)
	id := c.Param("id")
	log := logger.GetLoggerForContext(c, h.server.Log)
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*models.JwtCustomClaims)

	if err := c.Bind(req); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	post := &models.Post{}

	if err := h.postService.GetById(post, id, claims.OrganizationId); err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	if post.Id == constants.EmptyString {
		return apiutils.MessageResponse(c, http.StatusNotFound, constants.PostNotFound)
	}

	err := h.postService.Update(req, post, claims.UserId)
	if err != nil {
		return apiutils.InternalServerErrorResponse(c, constants.EmptyString, log, err)
	}

	return apiutils.MessageResponse(c, http.StatusOK, constants.PostUpdated)
}
