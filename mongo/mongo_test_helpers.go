package mongo

import (
	"context"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

// getTestMongoDB initializes a MongoDB connection for use in tests
func getTestMongoDB(ctx context.Context) (*Mongo, error) {
	mongoVersion := "4.4.8"

	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	mongoServer, err := mim.Start(ctx, mongoVersion)
	if err != nil {
		return nil, err
	}

	conn, err := mongoDriver.Open(getTestMongoDriverConfig(mongoServer, cfg.Database, cfg.Collections))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		MongoConfig: cfg.MongoConfig,
		Connection:  conn,
	}, nil
}

// Custom config to work with mongo in memory
func getTestMongoDriverConfig(mongoServer *mim.Server, database string, collections map[string]string) *mongoDriver.MongoDriverConfig {
	return &mongoDriver.MongoDriverConfig{
		ConnectTimeout:  5 * time.Second,
		QueryTimeout:    5 * time.Second,
		ClusterEndpoint: mongoServer.URI(),
		Database:        database,
		Collections:     collections,
	}
}
