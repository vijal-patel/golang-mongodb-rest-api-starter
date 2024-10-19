package routes

import (
	s "golang-mongodb-rest-api-starter/internal/server"
	"golang-mongodb-rest-api-starter/internal/server/handlers/v1"
	customMiddlware "golang-mongodb-rest-api-starter/internal/server/middleware"
	"golang-mongodb-rest-api-starter/internal/services"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

func ConfigureRoutes(server *s.Server) {
	authHandlerV1 := handlers.NewCustomAuthHandlerV1(server)
	registerHandlerV1 := handlers.NewRegisterHandlerV1(server)
	metaHandlerV1 := handlers.NewMetaHandlerV1(server)
	userHandlerV1 := handlers.NewUserHandlerV1(server)
	organizationHandlerV1 := handlers.NewOrganizationHandlerV1(server)
	postHandlerV1 := handlers.NewPostHandlerV1(server)

	server.Echo.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
	}))
	server.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     server.Config.HTTP.AllowedOrigins,
		AllowHeaders:     []string{echo.HeaderAccessControlAllowHeaders, echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderSetCookie},
		ExposeHeaders:    []string{echo.HeaderXRequestID, echo.HeaderSetCookie},
		AllowCredentials: true,
	}))
	server.Echo.Use(middleware.BodyLimit("2M"))
	server.Echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(30)))
	server.Echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			id := c.Response().Header().Get(echo.HeaderXRequestID)
			server.Log.Info(id,
				zap.String("method", c.Request().Method),
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
				zap.String("remote_ip", c.RealIP()),
				zap.Int64("latency", time.Since(v.StartTime).Milliseconds()),
			)
			return nil
		},
	}))

	v1Public := server.Echo.Group("/api/v1")
	v1Docs := server.Echo.Group("/docs/v1")

	if server.Config.Env == "dev" {
		v1Docs.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	v1Public.POST("/auth/logout", authHandlerV1.Logout)
	v1Public.POST("/auth/login", authHandlerV1.Login)
	v1Public.GET("/auth/refresh", authHandlerV1.RefreshToken)

	v1Public.GET("/health", metaHandlerV1.Health)
	v1Public.GET("/info", metaHandlerV1.Info) // TODO make private

	v1Public.POST("/register", registerHandlerV1.RegisterOrganization)
	v1Public.POST("/auth/password/recover", authHandlerV1.PasswordRecover)
	v1Public.POST("/auth/password/reset", authHandlerV1.PasswordReset)

	v1Private := server.Echo.Group("/api/v1")
	v1Private.Use(customMiddlware.CustomAuthMiddleware(server.Config.Auth.AccessSecret, server.Rbac, services.NewAuthzService()))
	v1Private.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(echo.Context) bool {
			return server.LogLevel.Level() == zap.ErrorLevel
		},
		Handler: func(c echo.Context, reqBody, resBody []byte) {
			id := c.Response().Header().Get(echo.HeaderXRequestID)
			server.Log.Info(id, zap.String("request", string(reqBody[:])), zap.String("response", string(resBody[:])))
		},
	}))

	v1Private.POST("/auth/password", authHandlerV1.PasswordChange)

	v1Private.PATCH("/logs/level", metaHandlerV1.PatchLogLevel)
	v1Private.GET("/logs/level", metaHandlerV1.GetLogLevel)

	v1Private.GET("/users", userHandlerV1.GetUsers)
	v1Private.POST("/users", userHandlerV1.CreateUser)
	v1Private.POST("/users/bulk", userHandlerV1.BulkCreateUsers)
	v1Private.GET("/users/:id", userHandlerV1.GetUser)
	v1Private.GET("/users/me", userHandlerV1.GetMyUser)
	v1Private.PATCH("/users/:id", userHandlerV1.UpdateUser)
	v1Private.DELETE("/users/:id", userHandlerV1.DeleteUser)
	v1Private.DELETE("/users/me", userHandlerV1.DeleteMyUser)

	v1Private.POST("/auth/otp/validate", authHandlerV1.ValidateOtp)
	v1Private.POST("/auth/otp/send", authHandlerV1.SendOtp)

	v1Private.DELETE("/organizations/me", organizationHandlerV1.DeleteOrganization)
	v1Private.GET("/organizations/me", organizationHandlerV1.GetOrganization) //,custom_middlware.AuthMiddleware(server.Config.Auth.AccessSecret, constants.EmptyString)
	v1Private.PATCH("/organizations/me", organizationHandlerV1.UpdateOrganization)

	v1Private.GET("/posts", postHandlerV1.GetPosts)
	v1Private.GET("/posts/:id", postHandlerV1.GetPost)
	v1Private.POST("/posts", postHandlerV1.CreatePost)
	v1Private.DELETE("/posts/:id", postHandlerV1.DeletePost)
	v1Private.PATCH("/posts/:id", postHandlerV1.UpdatePost)

}
