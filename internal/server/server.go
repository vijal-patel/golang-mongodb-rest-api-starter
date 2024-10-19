package server

import (
	"context"
	"fmt"
	"golang-mongodb-rest-api-starter/internal/cache"
	"golang-mongodb-rest-api-starter/internal/config"
	"golang-mongodb-rest-api-starter/internal/db"
	"golang-mongodb-rest-api-starter/internal/logger"
	"golang-mongodb-rest-api-starter/internal/validators"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mailjet/mailjet-apiv3-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Server struct {
	Config    *config.Config
	Cache     *redis.Client
	DB        *mongo.Client
	Echo      *echo.Echo
	Log       *zap.Logger
	LogLevel  zap.AtomicLevel
	Mailer    *mailjet.Client
	Rbac      *casbin.Enforcer
	Validator *validator.Validate
}

func NewServer(cfg *config.Config) *Server {
	log, atom := logger.NewLogger()
	enforcer, err := casbin.NewEnforcer(fmt.Sprintf("%s/keymatch_model.conf", os.Getenv("KO_DATA_PATH")), fmt.Sprintf("%s/keymatch_policy.csv", os.Getenv("KO_DATA_PATH")))
	if err != nil {
		log.Fatal("Could not load RBAC enforcer", zap.Error(err))
	}

	v := validator.New()
	validators.RegisterValidators(v)
	return &Server{
		Config:    cfg,
		Cache:     cache.Init(cfg, log),
		DB:        db.Init(cfg, log),
		Echo:      echo.New(),
		Log:       log,
		LogLevel:  atom,
		Mailer:    mailjet.NewMailjetClient(cfg.Email.ApiKey, cfg.Email.ApiSecret),
		Rbac:      enforcer,
		Validator: v,
	}
}

func (server *Server) Shutdown(ctx context.Context) {
	if err := db.Disconnect(server.DB); err != nil {
		server.Log.Error("Failed to disconnect from DB", zap.Error(err))
	}
	server.Log.Info("Disconnected from DB")

	// TODO Uncomment this when using a cacha
	// if err := cache.Disconnect(server.Cache); err != nil {
	// 	server.Log.Error("Failed to disconnect from Cache", zap.Error(err))
	// }
	// server.Log.Info("Disconnected from Cache")

	if err := server.Echo.Shutdown(ctx); err != nil {
		server.Log.Fatal("Failed to shutdown server", zap.Error(err))
	}
}

func (server *Server) Start() {

	go func() {
		if err := server.Echo.Start(":" + server.Config.HTTP.Port); err != nil && err != http.ErrServerClosed {
			server.Log.Fatal("shutting down the server", zap.Error(err))
		}
	}()

	// Create a quit channel which carries os.Signal values. Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)

	// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and relay
	// them to the quit channel. Any other signal will not be caught by signal.Notify()
	// and will retain their default behavior.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Read the signal from the quit channel. This code will block until a signal is
	// received.
	s := <-quit

	// Log a message to say we caught the signal. Notice that we also call the
	// String() method on the signal to get the signal name and include it in the log
	// entry properties.
	server.Log.Info("caught signal", zap.String("signal", s.String()))

	// Create a context with a 5-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Shutdown(ctx)

}
