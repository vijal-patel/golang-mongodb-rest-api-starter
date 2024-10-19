package main

import (
	"fmt"
	application "golang-mongodb-rest-api-starter/internal"
	"golang-mongodb-rest-api-starter/internal/config"

	"golang-mongodb-rest-api-starter/docs"
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath /
func main() {
	cfg := config.NewConfig()
	if cfg.Env == "dev" {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.ExposePort)
	}
	application.Start(cfg)
}
