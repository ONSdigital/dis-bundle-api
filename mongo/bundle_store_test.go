package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoCRUD(t *testing.T) {
	var (
		ctx = context.Background()
		now = time.Now()
	)

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		if err := setupBundleTestData(ctx, mongodb); err != nil {
			t.Fatalf("failed to insert test data, skipping tests: %v", err)
		}

		// GetBundle
		Convey("When GetBundle is called", func() {
			Convey("And the bundle is fetched successfully", func() {
				returnedBundle, err := mongodb.GetBundle(ctx, "bundle1")

				Convey("Then the bundle is returned without error", func() {
					So(err, ShouldBeNil)
					So(returnedBundle.BundleType, ShouldEqual, models.BundleTypeScheduled)
				})
			})

			Convey("And the bundle is not found", func() {
				Convey("Then a bundle not found error is returned", func() {
					_, err := mongodb.GetBundle(ctx, "bundle-not-exist")
					So(err, ShouldEqual, apierrors.ErrBundleNotFound)
				})
			})

			Convey("And GetBundle returns a generic error", func() {
				mongodb.Connection.DropDatabase(ctx)

				Convey("Then an error is returned", func() {
					_, err := mongodb.GetBundle(ctx, "bundle1")
					So(err, ShouldNotBeNil)
				})
			})
		})

		// ListBundles
		Convey("When ListBundles is called", func() {
			Convey("And the bundles list is fetched successfully", func() {
				bundlesList, TotalCount, err := mongodb.ListBundles(ctx, 0, 10)

				Convey("Then the bundles are returned without error", func() {
					So(err, ShouldBeNil)
					So(TotalCount, ShouldEqual, 2)
					So(len(bundlesList), ShouldEqual, 2)

					So(bundlesList[0].ID, ShouldEqual, "bundle1")
					So(bundlesList[1].ID, ShouldEqual, "bundle2")
				})
			})

			Convey("And ListBundles returns an error", func() {
				mongodb.Connection.Close(ctx)

				Convey("Then an error is returned", func() {
					_, _, err := mongodb.ListBundles(ctx, 0, 10)
					So(err, ShouldNotBeNil)
				})
			})
		})

		// Create Bundle
		Convey("When CreateBundle is called", func() {
			Convey("And the bundle is created successfully", func() {
				myBundle := models.Bundle{
					ID:         "bundle3",
					BundleType: "scheduled",
				}
				err = mongodb.CreateBundle(ctx, &myBundle)

				Convey("Then the bundle is created without error", func() {
					So(err, ShouldBeNil)
					_, err := mongodb.GetBundle(ctx, "bundle3")
					So(err, ShouldBeNil)
				})
			})

			Convey("And CreateBundle returns an error", func() {
				b := &models.Bundle{ID: "bundle1"}
				err := mongodb.CreateBundle(ctx, b) // ID already exists

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "duplicate key error")
				})
			})
		})

		// Update Bundle
		Convey("When UpdateBundle is called", func() {
			Convey("And the bundle is updated successfully", func() {
				myBundle := models.Bundle{
					BundleType: models.BundleTypeManual,
					CreatedAt:  &now,
				}
				updatedBundle, err := mongodb.UpdateBundle(ctx, "bundle1", &myBundle)
				So(err, ShouldBeNil)

				Convey("Then the bundle is updated without error", func() {
					So(updatedBundle.BundleType, ShouldEqual, myBundle.BundleType)
				})
			})

			Convey("And UpdateBundle returns an error", func() {
				myBundle := models.Bundle{
					BundleType: models.BundleTypeManual,
					CreatedAt:  &now,
				}
				mongodb.Connection.Close(ctx)
				_, err := mongodb.UpdateBundle(ctx, "bundle1", &myBundle)

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		// Delete Bundle
		Convey("When DeleteBundle is called", func() {
			Convey("And the bundle is deleted successfully", func() {
				err = mongodb.DeleteBundle(ctx, "bundle1")
				So(err, ShouldBeNil)

				Convey("Then the bundle should not be found", func() {
					_, err = mongodb.GetBundle(ctx, "bundle1")
					So(err, ShouldResemble, apierrors.ErrBundleNotFound)
				})
			})

			Convey("And we try to delete a non-existent bundle", func() {
				err := mongodb.DeleteBundle(ctx, "non-existent-id")

				Convey("Then a bundle not found error is returned", func() {
					So(err, ShouldResemble, apierrors.ErrBundleNotFound)
				})
			})

			Convey("And DeleteBundle returns a generic error", func() {
				mongodb.Connection.Close(ctx)
				err := mongodb.DeleteBundle(ctx, "bundle1")

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

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

func TestBuildListBundlesQuery(t *testing.T) {
	t.Parallel()

	Convey("When we call buildListBundlesQuery", t, func() {
		filter, sort := buildListBundlesQuery()

		expectedFilter := bson.M{}
		expectedSort := bson.M{"updated_at": -1}

		So(filter, ShouldResemble, expectedFilter)
		So(sort, ShouldResemble, expectedSort)
	})
}

func TestBuildGetBundleQuery(t *testing.T) {
	t.Parallel()

	Convey("Given a bundle ID", t, func() {
		bundleID := "abc123"

		Convey("When we call buildGetBundleQuery", func() {
			query := buildGetBundleQuery(bundleID)

			Convey("Then it should return a query with the correct ID", func() {
				expected := bson.M{"_id": bundleID}
				So(query, ShouldResemble, expected)
			})
		})
	})
}
