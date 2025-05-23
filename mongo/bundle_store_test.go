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
		now          = time.Now()
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
				So(returnedBundle.BundleType, ShouldEqual, models.BundleTypeScheduled)
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
					BundleType: models.BundleTypeManual,
					CreatedAt:  &now,
				}
				updatedBundle, err := mongodb.UpdateBundle(ctx, "bundle1", &myBundle)
				So(err, ShouldBeNil)
				So(updatedBundle.BundleType, ShouldEqual, myBundle.BundleType)
			})

			Convey("When the bundle returns error", func() {
				myBundle := models.Bundle{
					BundleType: models.BundleTypeManual,
					CreatedAt:  &now,
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

	now := time.Now()
	oneDayFromNow := now.Add(24 * time.Hour)
	twoDaysFromNow := now.Add(48 * time.Hour)
	draft := models.BundleStateDraft
	bundles := []*models.Bundle{
		{
			ID:            "bundle1",
			BundleType:    models.BundleTypeScheduled,
			CreatedBy:     &models.User{Email: "user1@ons.gov.uk"},
			CreatedAt:     &now,
			LastUpdatedBy: &models.User{Email: "user1@ons.gov.uk"},
			PreviewTeams:  &[]models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
			ScheduledAt:   &oneDayFromNow, // 1 day from now
			State:         &draft,
			Title:         "Scheduled Bundle 1",
			UpdatedAt:     &now,
			ManagedBy:     models.ManagedByDataAdmin,
		},
		{
			ID:            "bundle2",
			BundleType:    models.BundleTypeManual,
			CreatedBy:     &models.User{Email: "user2@ons.gov.uk"},
			CreatedAt:     &now,
			LastUpdatedBy: &models.User{Email: "user2@ons.gov.uk"},
			PreviewTeams:  &[]models.PreviewTeam{{ID: "team3"}},
			ScheduledAt:   &twoDaysFromNow, // 2 days from now
			State:         &draft,
			Title:         "Manual Bundle 2",
			UpdatedAt:     &now,
			ManagedBy:     models.ManagedByWagtail,
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
		expectedSort := bson.M{"updated_at": -1}

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
