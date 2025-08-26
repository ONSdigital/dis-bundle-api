package mongo

import (
	"context"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	collectionNames = []string{
		"bundles",
		"bundle_events",
		"bundle_contents",
	}
)

// getTestMongoDB initializes a MongoDB connection for use in tests
func getTestMongoDB(ctx context.Context) (*Mongo, *mim.Server, error) {
	mongoVersion := "4.4.8"

	cfg, err := config.Get()
	if err != nil {
		return nil, nil, err
	}

	mongoServer, err := mim.Start(ctx, mongoVersion)
	if err != nil {
		return nil, nil, err
	}
	mongoConfig := getTestMongoDriverConfig(mongoServer, cfg.Database, cfg.Collections)
	conn, err := mongoDriver.Open(mongoConfig)
	if err != nil {
		return nil, nil, err
	}

	return &Mongo{
		MongoConfig: cfg.MongoConfig,
		Connection:  conn,
	}, mongoServer, nil
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

func SetupIndexes(ctx context.Context, mimServer *mim.Server) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mimServer.URI()))
	if err != nil {
		return err
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Error(ctx, "failed to disconnect mongo client", err)
		}
	}(client, ctx)

	for _, collectionName := range collectionNames {
		collection := client.Database("bundles").Collection(collectionName)

		indexModel := mongo.IndexModel{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		}

		_, err = collection.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			return err
		}
	}

	return nil
}
