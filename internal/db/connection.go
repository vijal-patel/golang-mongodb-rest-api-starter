package db

import (
	"context"
	"fmt"
	"golang-mongodb-rest-api-starter/internal/config"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func Init(cfg *config.Config, log *zap.Logger) *mongo.Client {
	log.Info("Connecting to MongoDB...")
	mongoDbConnectionString := fmt.Sprintf("mongodb+srv://%s:%s@%s.%s/?retryWrites=true&w=majority",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.Host)
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 300 * time.Second,
	}

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongoDbConnectionString).
		SetServerAPIOptions(serverAPIOptions)
	clientOptions.SetMaxConnIdleTime(60 * time.Second)
	clientOptions.SetDialer(dialer)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB", zap.Error(err))
	}
	log.Info("Successfully connected to MongoDB!")

	return client
}

func Disconnect(client *mongo.Client) error {
	return client.Disconnect(context.TODO())
}

func Ping(client *mongo.Client) error {
	return client.Ping(context.TODO(), nil)
}
