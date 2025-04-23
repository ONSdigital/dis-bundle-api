package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoCRUD(t *testing.T) {
	// Config for test mongo in memory instance
	var (
		ctx          = context.Background()
		mongoVersion = "4.4.8"
		mongoServer  *mim.Server
		err          error
	)

	// Get the default app config to use when setting up mongo in memory
	cfg, _ := config.Get()

	Convey("When the dbconnection is initialized correctly", t, func() {
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

		if err := setupTestData(ctx, mongodb); err != nil {
			t.Fatalf("failed to insert test data, skipping tests: %v", err)
		}

		// GetBundle
		Convey("When the GetBudle is called", func() {
			Convey("When the bundle is fetched successfully", func() {
				returnedBundle, err := mongodb.GetBundle(ctx, "bundle1")
				So(err, ShouldBeNil)
				So(returnedBundle.BundleType, ShouldEqual, "scheduled")
			})

			Convey("When the bundle is not found", func() {
				_, err := mongodb.GetBundle(ctx, "bundle-not-exist")
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})

			Convey("When GetBundle returns a generic error", func() {
				mongodb.Connection.DropDatabase(ctx)
				_, err := mongodb.GetBundle(ctx, "bundle1")
				So(err, ShouldNotBeNil)
			})
		})

		// ListBundles
		Convey("When the ListBundle is called", func() {
			Convey("When the bundles list is fetched successfully", func() {
				bundlesList, TotalCount, err := mongodb.ListBundles(ctx, 0, 10)
				So(err, ShouldBeNil)
				So(TotalCount, ShouldEqual, 2)
				So(len(bundlesList), ShouldEqual, 2)

				So(bundlesList[0].ID, ShouldEqual, "bundle1")
				So(bundlesList[1].ID, ShouldEqual, "bundle2")
			})

			Convey("When the ListBundles returns an error", func() {
				mongodb.Connection.Close(ctx)
				_, _, err := mongodb.ListBundles(ctx, 0, 10)
				So(err, ShouldNotBeNil)
			})
		})

		// Create Bundle
		Convey("When the CreateBundle is called", func() {
			Convey("When the bundle is created successfully", func() {
				myBundle := models.Bundle{
					ID:         "bundle3",
					BundleType: "scheduled",
				}
				err = mongodb.CreateBundle(ctx, &myBundle)
				So(err, ShouldBeNil)
			})

			Convey("When CreateBundle returns an error", func() {
				b := &models.Bundle{ID: "bundle1"}
				err := mongodb.CreateBundle(ctx, b) // ID already exists
				So(err, ShouldNotBeNil)
			})
		})

		// Update Bundle
		Convey("When the UpdateBundle is called", func() {
			Convey("When the bundle is updated successfully", func() {
				myBundle := models.Bundle{
					BundleType:  "UpdatedType",
					CreatedDate: time.Now(),
				}
				updatedBundle, err := mongodb.UpdateBundle(ctx, "bundle1", &myBundle)
				So(err, ShouldBeNil)
				So(updatedBundle.BundleType, ShouldEqual, myBundle.BundleType)
			})

			Convey("When the bundle returns error", func() {
				myBundle := models.Bundle{
					BundleType:  "UpdatedType",
					CreatedDate: time.Now(),
				}
				mongodb.Connection.Close(ctx)
				_, err := mongodb.UpdateBundle(ctx, "bundle1", &myBundle)
				So(err, ShouldNotBeNil)
			})
		})

		// Delete Bundle
		Convey("When the DeleteBundle is called", func() {
			Convey("When the bundle is deleted successfully", func() {
				err = mongodb.DeleteBundle(ctx, "bundle1")
				So(err, ShouldBeNil)

				// verify if  the bundle is deleted
				_, err = mongodb.GetBundle(ctx, "bundle1")
				So(err, ShouldResemble, apierrors.ErrBundleNotFound)
			})

			Convey("When trying to delete a non-existent bundle and returns error", func() {
				err := mongodb.DeleteBundle(ctx, "non-existent-id")
				So(err, ShouldResemble, apierrors.ErrBundleNotFound)
			})

			Convey("When trying to delete a bundle and returns generic error", func() {
				mongodb.Connection.Close(ctx)
				err := mongodb.DeleteBundle(ctx, "bundle1")
				So(err, ShouldResemble, err)
			})
		})
	})
}

func setupTestData(ctx context.Context, mongo *Mongo) error {
	if err := mongo.Connection.DropDatabase(ctx); err != nil {
		return err
	}

	bundles := []*models.Bundle{
		{
			ID:         "bundle1",
			BundleType: "scheduled",
			Contents: []models.BundleContent{
				{
					DatasetID: "dataset1",
					EditionID: "edition1",
					ItemID:    "item1",
					State:     "published",
					Title:     "Dataset 1",
					URLPath:   "/dataset1/edition1/item1",
				},
				{
					DatasetID: "dataset2",
					EditionID: "edition2",
					ItemID:    "item2",
					State:     "draft",
					Title:     "Dataset 2",
					URLPath:   "/dataset2/edition2/item2",
				},
			},
			Creator:         "user1",
			CreatedDate:     time.Now(),
			LastUpdatedBy:   "user1",
			PreviewTeams:    []string{"team1", "team2"},
			PublishDateTime: time.Now().Add(24 * time.Hour), // 1 day from now
			State:           "active",
			Title:           "Scheduled Bundle 1",
			UpdatedDate:     time.Now(),
			WagtailManaged:  false,
		},
		{
			ID:         "bundle2",
			BundleType: "manual",
			Contents: []models.BundleContent{
				{
					DatasetID: "dataset3",
					EditionID: "edition3",
					ItemID:    "item3",
					State:     "draft",
					Title:     "Dataset 3",
					URLPath:   "/dataset3/edition3/item3",
				},
			},
			Creator:         "user2",
			CreatedDate:     time.Now(),
			LastUpdatedBy:   "user2",
			PreviewTeams:    []string{"team3"},
			PublishDateTime: time.Now().Add(48 * time.Hour), // 2 days from now
			State:           "inactive",
			Title:           "Manual Bundle 2",
			UpdatedDate:     time.Now(),
			WagtailManaged:  true,
		},
	}

	for _, b := range bundles {
		if err := mongo.CreateBundle(ctx, b); err != nil {
			return err
		}
	}

	return nil
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
		expectedSort := bson.M{"_id": 1}

		So(filter, ShouldResemble, expectedFilter)
		So(sort, ShouldResemble, expectedSort)
	})
}

func TestBuildGetBundleQuery(t *testing.T) {
	t.Parallel()

	Convey("When building query for GetBundle", t, func() {
		bundleID := "abc123"
		query := buildGetBundleQuery(bundleID)
		expected := bson.M{"_id": "abc123"}
		So(query, ShouldResemble, expected)
	})
}
