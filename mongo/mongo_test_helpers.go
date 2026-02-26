package mongo

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/testcontainers/testcontainers-go"
	testMongoContainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
)

// getTestMongoDB initializes a MongoDB connection for use in tests
func getTestMongoDB(ctx context.Context, t *testing.T) (*Mongo, error) {
	t.Helper()

	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	mongoContainer, err := testMongoContainer.Run(ctx, "mongo:4.4.8")
	if err != nil {
		return nil, fmt.Errorf("failed to start MongoDB container: %w", err)
	}
	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, mongoContainer)
	})

	mongoConfig, err := getTestMongoDriverConfig(ctx, mongoContainer, cfg.Database, cfg.Collections)
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB driver config: %w", err)
	}

	conn, err := mongoDriver.Open(mongoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open MongoDB connection: %w", err)
	}

	return &Mongo{
		MongoConfig: cfg.MongoConfig,
		Connection:  conn,
	}, nil
}

func getTestMongoDriverConfig(ctx context.Context, mongoContainer *testMongoContainer.MongoDBContainer, database string, collections map[string]string) (*mongoDriver.MongoDriverConfig, error) {
	connectionString, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB connection string: %w", err)
	}

	connectionStringURL, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MongoDB connection string: %w", err)
	}

	return &mongoDriver.MongoDriverConfig{
		ConnectTimeout:  5 * time.Second,
		QueryTimeout:    5 * time.Second,
		ClusterEndpoint: connectionStringURL.Host,
		Database:        database,
		Collections:     collections,
	}, nil
}
