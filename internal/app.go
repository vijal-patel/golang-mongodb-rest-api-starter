package application

import (
	"golang-mongodb-rest-api-starter/internal/config"
	"golang-mongodb-rest-api-starter/internal/server"
	"golang-mongodb-rest-api-starter/internal/server/routes"
)

func Start(cfg *config.Config) {
	app := server.NewServer(cfg)

	routes.ConfigureRoutes(app)

	app.Start()

}
