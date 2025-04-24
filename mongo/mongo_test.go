package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	. "github.com/smartystreets/goconvey/convey"

	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

func TestMongoInit(t *testing.T) {
	Convey("Given a MongoDB server is running", t, func() {
		var (
			mongoVersion = "4.4.8"
		)
		cfg, _ := config.Get()

		ctx := context.Background()
		mongoServer, err := mim.Start(ctx, mongoVersion)
		if err != nil {
			t.Fatalf("failed to start mongo server: %v", err)
		}
		defer mongoServer.Stop(ctx)

		// Successful Init scenario
		Convey("When Init is called on the Mongo wrapper with a valid config", func() {
			conn, err := mongoDriver.Open(getMongoDriverConfig(mongoServer, cfg.Database, cfg.Collections))
			So(err, ShouldBeNil)

			mongodb := &Mongo{
				MongoConfig: cfg.MongoConfig,
				Connection:  conn,
			}

			err = mongodb.Init(ctx)

			Convey("Then Init should succeed", func() {
				So(err, ShouldBeNil)
				So(mongodb.Connection, ShouldNotBeNil)
				So(mongodb.healthClient, ShouldNotBeNil)
			})

			Convey("And the connection should ping successfully", func() {
				pingErr := mongodb.Connection.Ping(ctx, 2*time.Second)
				So(pingErr, ShouldBeNil)
			})
		})

		// Error scenario
		Convey("When Init is called with an invalid MongoDB address", func() {
			badCfg := cfg.MongoConfig
			badCfg.MongoDriverConfig.ClusterEndpoint = "invalidhost:9999" // Invalid address

			mongodb := &Mongo{
				MongoConfig: badCfg,
			}

			err := mongodb.Init(ctx)

			Convey("Then Init should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})

	})
}

func TestMongoClose(t *testing.T) {
	Convey("Given a MongoDB connection is established", t, func() {
		ctx := context.Background()
		cfg, _ := config.Get()

		mongoServer, err := mim.Start(ctx, "4.4.8")
		if err != nil {
			t.Fatalf("failed to start mongo server: %v", err)
		}
		defer mongoServer.Stop(ctx)

		conn, err := mongoDriver.Open(getMongoDriverConfig(mongoServer, cfg.Database, cfg.Collections))
		So(err, ShouldBeNil)

		mongodb := &Mongo{
			MongoConfig: cfg.MongoConfig,
			Connection:  conn,
		}

		Convey("When Close is called", func() {
			err := mongodb.Close(ctx)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And further use of the connection should fail", func() {
				pingErr := mongodb.Connection.Ping(ctx, 2*time.Second)
				So(pingErr, ShouldNotBeNil)
			})
		})
	})
}
