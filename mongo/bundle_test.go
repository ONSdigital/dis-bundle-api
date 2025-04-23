package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMimMongo(t *testing.T) {
	// Config for test mongo in memory instance
	var (
		ctx          = context.Background()
		mongoVersion = "4.4.8"
		mongoServer  *mim.Server
		err          error
	)

	// Get the default app config to use when setting up mongo in memory
	cfg, _ := config.Get()

	Convey("", t, func() {

		mongoServer, err = mim.Start(ctx, mongoVersion)
		So(err, ShouldBeNil)
		defer mongoServer.Stop(ctx)

		conn, err := mongoDriver.Open(getMongoDriverConfig(mongoServer, cfg.Database, cfg.Collections))
		So(err, ShouldBeNil)
		So(conn, ShouldNotBeNil)

		mongodb := &Mongo{
			MongoConfig: cfg.MongoConfig,
			Connection:  conn,
		}
		So(mongodb, ShouldNotBeNil)
		// Try to create a bundle
		myBundle := models.Bundle{
			ID:         "test",
			BundleType: "test",
		}
		err = mongodb.CreateBundle(ctx, &myBundle)
		So(err, ShouldBeNil)

		returnedBundle, err := mongodb.GetBundle(ctx, myBundle.ID)
		So(err, ShouldBeNil)
		So(returnedBundle.BundleType, ShouldEqual, myBundle.BundleType)

	})
}

// Custom config to work with mongo in memory
func getMongoDriverConfig(mongoServer *mim.Server, database string, collections map[string]string) *mongoDriver.MongoDriverConfig {
	return &mongoDriver.MongoDriverConfig{
		ConnectTimeout:  5 * time.Second,
		QueryTimeout:    5 * time.Second,
		ClusterEndpoint: mongoServer.URI(),
		Database:        database,
		Collections:     collections,
	}
}

func TestBuildListBundlesQuery(t *testing.T) {
	t.Parallel()

	Convey("When building the list bundles query", t, func() {
		filter, sort := buildListBundlesQuery()

		expectedFilter := bson.M{}
		expectedSort := bson.M{"id": -1}

		So(filter, ShouldResemble, expectedFilter)
		So(sort, ShouldResemble, expectedSort)
	})
}

func TestBuildGetBundleQuery(t *testing.T) {
	t.Parallel()

	Convey("When building query for GetBundle", t, func() {
		bundleID := "abc123"
		query := buildGetBundleQuery(bundleID)
		expected := bson.M{"id": "abc123"}
		So(query, ShouldResemble, expected)
	})
}

func TestBundleUpdateQuery(t *testing.T) {
	t.Parallel()

	Convey("When a full bundle object is provided", t, func() {
		createdDate, _ := time.Parse(time.RFC3339, "2025-04-04T07:00:00.000Z")

		bundle := &models.Bundle{
			ID:              "9e4e3628-fc85-48cd-80ad-e005d9d283ff",
			BundleType:      "scheduled",
			Creator:         "publisher@ons.gov.uk",
			CreatedDate:     createdDate,
			LastUpdatedBy:   "publisher@ons.gov.uk",
			PreviewTeams:    []string{"string"},
			PublishDateTime: createdDate,
			State:           "approved",
			Title:           "CPI March 2025",
			UpdatedDate:     createdDate,
			WagtailManaged:  true,
			Contents: []models.BundleContent{
				{
					DatasetID: "cpih",
					EditionID: "march",
					ItemID:    "DE3BC0B6-D6C4-4E20-917E-95D7EA8C91DC",
					State:     "published",
					Title:     "Consumer Prices Index",
					URLPath:   "/datasets/cpih/editions/time-series/versions/1",
				},
			},
		}

		expected := bson.M{
			"$set":         bundle,
			"$setOnInsert": bson.M{"created_date": createdDate},
		}

		actual := bundleUpdateQuery(bundle)

		So(actual, ShouldNotBeNil)
		So(actual, ShouldResemble, expected)
	})
}
