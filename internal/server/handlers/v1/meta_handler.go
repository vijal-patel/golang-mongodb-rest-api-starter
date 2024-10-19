package handlers

import (
	"context"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/utils/apiutils"

	"golang-mongodb-rest-api-starter/internal/logger"

	s "golang-mongodb-rest-api-starter/internal/server"
	"golang-mongodb-rest-api-starter/internal/vcs"

	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap/zapcore"
)

type MetaHandlerV1 struct {
	server *s.Server
}

type InfoResponse struct {
	Version string `json:"version"`
	Env     string `json:"env"`
}

func NewMetaHandlerV1(server *s.Server) *MetaHandlerV1 {
	return &MetaHandlerV1{server: server}
}

// Register godoc
// @Summary Health Check
// @Description health check
// @ID health
// @Tags Meta
// @Accept json
// @Produce json
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Router /meta/health [get]
func (h *MetaHandlerV1) Health(c echo.Context) error {
	if err := h.server.DB.Ping(context.TODO(), nil); err != nil {
		return apiutils.InternalServerErrorResponse(c, "DB connection error", logger.GetLoggerForContext(c, h.server.Log), err)
	}
	return apiutils.MessageResponse(c, http.StatusOK, "ok")
}

// Register godoc
// @Summary Info
// @Description version and env
// @ID info
// @Tags Meta
// @Accept json
// @Produce json
// @Success 200 {object} InfoResponse
// @Failure 400 {object} models.Error
// @Router /meta/info [get]
func (h *MetaHandlerV1) Info(c echo.Context) error {
	info := InfoResponse{
		Version: vcs.Version(),
		Env:     h.server.Config.Env,
	}
	return apiutils.Response(c, http.StatusOK, info)

}

// Register godoc
// @Summary Log level
// @Description log level
// @ID meta
// @Tags Meta
// @Accept json
// @Produce json
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Router /meta/log/level [get]
func (h *MetaHandlerV1) GetLogLevel(c echo.Context) error {
	return apiutils.MessageResponse(c, http.StatusOK, h.server.LogLevel.Level().String())
}

// Register godoc
// @Summary Log level
// @Description log level
// @ID meta
// @Tags Meta
// @Accept json
// @Produce json
// @Param params body models.PatchLogLevelRequest true "Log Level"
// @Success 200 {object} models.Data
// @Failure 400 {object} models.Error
// @Router /meta/log/level [patch]
func (h *MetaHandlerV1) PatchLogLevel(c echo.Context) error {
	req := new(models.PatchLogLevelRequest)

	if err := c.Bind(req); err != nil {
		return apiutils.BadRequestResponse(c, constants.EmptyString)
	}

	if err := h.server.Validator.Struct(req); err != nil {
		return apiutils.ValidationErrorResponse(c, err, constants.EmptyString)
	}

	log := logger.GetLoggerForContext(c, h.server.Log)
	log.Infof("setting log level to %s", req.Level)
	switch req.Level {
	case "info":
		h.server.LogLevel.SetLevel(zapcore.InfoLevel)
	case "debug":
		h.server.LogLevel.SetLevel(zapcore.DebugLevel)
	case "error":
		h.server.LogLevel.SetLevel(zapcore.ErrorLevel)
	default:
		return apiutils.BadRequestResponse(c, "not supported")
	}

	return apiutils.MessageResponse(c, http.StatusOK, h.server.LogLevel.Level().String())
}
