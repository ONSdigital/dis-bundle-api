package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func setupBundleTestData(ctx context.Context, mongo *Mongo) error {
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
			ScheduledAt:   &oneDayFromNow,
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
			ScheduledAt:   &twoDaysFromNow,
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

func TestListBundles_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When ListBundles is called", func() {
			bundles, totalCount, err := mongodb.ListBundles(ctx, 0, 10)

			Convey("Then it should return the correct bundles and total count", func() {
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 2)
				So(len(bundles), ShouldEqual, 2)

				So(bundles[0].ID, ShouldEqual, "bundle1")
				So(bundles[0].BundleType, ShouldEqual, models.BundleTypeScheduled)
				So(bundles[1].ID, ShouldEqual, "bundle2")
				So(bundles[1].BundleType, ShouldEqual, models.BundleTypeManual)
			})
		})
	})
}

func TestListBundles_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When ListBundles is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			bundles, totalCount, err := mongodb.ListBundles(ctx, 0, 10)

			Convey("Then it should return an error and no bundles", func() {
				So(err, ShouldNotBeNil)
				So(bundles, ShouldBeEmpty)
				So(totalCount, ShouldEqual, 0)
			})
		})
	})
}

func TestBuildListBundlesQuery(t *testing.T) {
	t.Parallel()

	Convey("When we call buildListBundlesQuery", t, func() {
		filter, sort := buildListBundlesQuery()

		Convey("Then it should return an empty filter and sort by updated_at descending", func() {
			expectedFilter := bson.M{}
			expectedSort := bson.M{"updated_at": -1}

			So(filter, ShouldResemble, expectedFilter)
			So(sort, ShouldResemble, expectedSort)
		})
	})
}

func TestGetBundle_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When GetBundle is called with an existing bundle ID", func() {
			bundle, err := mongodb.GetBundle(ctx, "bundle1")

			Convey("Then it should return the correct bundle without error", func() {
				So(err, ShouldBeNil)
				So(bundle.ID, ShouldEqual, "bundle1")
			})
		})
	})
}

func TestGetBundle_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When GetBundle is called with a non-existent bundle ID", func() {
			_, err := mongodb.GetBundle(ctx, "non-existent-id")

			Convey("Then it should return a bundle not found error", func() {
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})
		})

		Convey("When GetBundle is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			_, err := mongodb.GetBundle(ctx, "bundle1")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldNotEqual, apierrors.ErrBundleNotFound)
			})
		})
	})
}

func TestBuildGetBundleQuery(t *testing.T) {
	t.Parallel()

	Convey("Given a bundle ID", t, func() {
		bundleID := "abc123"

		Convey("When we call buildGetBundleQuery", func() {
			query := buildGetBundleQuery(bundleID)

			Convey("Then it should return a query with the correct ID", func() {
				expected := bson.M{"id": bundleID}
				So(query, ShouldResemble, expected)
			})
		})
	})
}

func TestCreateBundle_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CreateBundle is called with a new bundle", func() {
			newBundle := &models.Bundle{
				ID:           "NewBundle",
				BundleType:   models.BundleTypeManual,
				PreviewTeams: &[]models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
				Title:        "New Bundle",
				ManagedBy:    models.ManagedByWagtail,
			}
			err = mongodb.CreateBundle(ctx, newBundle)

			Convey("Then it should create the bundle without error", func() {
				So(err, ShouldBeNil)

				returnedBundle, err := mongodb.GetBundle(ctx, "NewBundle")
				So(err, ShouldBeNil)
				So(returnedBundle.ID, ShouldEqual, "NewBundle")
				So(returnedBundle.BundleType, ShouldEqual, models.BundleTypeManual)
				So(returnedBundle.PreviewTeams, ShouldResemble, &[]models.PreviewTeam{{ID: "team1"}, {ID: "team2"}})
				So(returnedBundle.Title, ShouldEqual, "New Bundle")
				So(returnedBundle.ManagedBy, ShouldEqual, models.ManagedByWagtail)
			})
		})
	})
}

func TestCreateBundle_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, mimServer, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		err = SetupIndexes(ctx, mimServer, cfg.Database, cfg.Collections[config.BundlesCollection])
		So(err, ShouldBeNil)

		Convey("When CreateBundle is called with an existing bundle ID", func() {
			existingBundle := &models.Bundle{
				ID:         "bundle1",
				BundleType: models.BundleTypeScheduled,
			}
			err = mongodb.CreateBundle(ctx, existingBundle)

			Convey("Then it should return a duplicate key error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "duplicate key error")
			})
		})

		Convey("When CreateBundle is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			newBundle := &models.Bundle{ID: "NewBundleFailure"}
			err = mongodb.CreateBundle(ctx, newBundle)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUpdateBundle_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When UpdateBundle is called with an existing bundle ID", func() {
			bundleUpdate := &models.Bundle{
				Title:     "Updated Bundle",
				ManagedBy: models.ManagedByWagtail,
			}
			updatedBundle, err := mongodb.UpdateBundle(ctx, "bundle1", bundleUpdate)

			Convey("Then it should update the bundle without error", func() {
				So(err, ShouldBeNil)
				So(updatedBundle.Title, ShouldEqual, bundleUpdate.Title)
				So(updatedBundle.ManagedBy, ShouldEqual, bundleUpdate.ManagedBy)
			})
		})
	})
}

func TestUpdateBundle_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When UpdateBundle is called with a non-existent bundle ID", func() {
			bundleUpdate := &models.Bundle{
				Title:     "Non-existent Bundle Update",
				ManagedBy: models.ManagedByWagtail,
			}
			returnedBundle, err := mongodb.UpdateBundle(ctx, "non-existent-id", bundleUpdate)

			Convey("Then it should return a bundle not found error", func() {
				So(returnedBundle, ShouldBeNil)
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})
		})

		Convey("When UpdateBundle is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			bundleUpdate := &models.Bundle{Title: "Connection Failure Update"}
			_, err := mongodb.UpdateBundle(ctx, "bundle1", bundleUpdate)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldNotEqual, apierrors.ErrBundleNotFound)
			})
		})
	})
}

func TestDeleteBundle_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When DeleteBundle is called with an existing bundle ID", func() {
			err = mongodb.DeleteBundle(ctx, "bundle1")

			Convey("Then it should delete the bundle without error", func() {
				So(err, ShouldBeNil)

				_, err := mongodb.GetBundle(ctx, "bundle1")
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})
		})
	})
}

func TestDeleteBundle_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When DeleteBundle is called with a non-existent bundle ID", func() {
			err := mongodb.DeleteBundle(ctx, "non-existent-id")

			Convey("Then it should return a bundle not found error", func() {
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})
		})

		Convey("When DeleteBundle is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			err := mongodb.DeleteBundle(ctx, "bundle1")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldNotEqual, apierrors.ErrBundleNotFound)
			})
		})
	})
}

// Test GetBundleByTitle
func TestGetBundleByTitle_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When GetBundleByTitle is called with an existing bundle title", func() {
			bundle, err := mongodb.GetBundleByTitle(ctx, "Scheduled Bundle 1")

			Convey("Then it should return the correct bundle without error", func() {
				So(err, ShouldBeNil)
				So(bundle.ID, ShouldEqual, "bundle1")
				So(bundle.Title, ShouldEqual, "Scheduled Bundle 1")
			})
		})
	})
}

func TestGetBundleByTitle_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When GetBundleByTitle is called with a non-existent bundle title", func() {
			_, err := mongodb.GetBundleByTitle(ctx, "non-existent-title")

			Convey("Then it should return a bundle not found error", func() {
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})
		})

		Convey("When GetBundleByTitle is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			_, err := mongodb.GetBundleByTitle(ctx, "Scheduled Bundle 1")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldNotEqual, apierrors.ErrBundleNotFound)
			})
		})
	})
}
