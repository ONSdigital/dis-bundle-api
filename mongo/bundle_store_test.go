package mongo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func setupBundleTestData(ctx context.Context, mongo *Mongo) ([]*models.Bundle, error) {
	if err := mongo.Connection.DropDatabase(ctx); err != nil {
		return nil, err
	}

	now := time.Now()
	oneDayFromNow := now.Add(24 * time.Hour)
	twoDaysFromNow := now.Add(48 * time.Hour)
	draft := models.BundleStateDraft
	approved := models.BundleStateApproved
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
		{
			ID:            "bundle3",
			BundleType:    models.BundleTypeManual,
			CreatedBy:     &models.User{Email: "user2@ons.gov.uk"},
			CreatedAt:     &now,
			LastUpdatedBy: &models.User{Email: "user2@ons.gov.uk"},
			PreviewTeams:  &[]models.PreviewTeam{{ID: "team3"}},
			ScheduledAt:   nil,
			State:         &approved,
			Title:         "Manual Bundle 3",
			UpdatedAt:     &now,
			ManagedBy:     models.ManagedByWagtail,
		},
	}

	for _, b := range bundles {
		if err := mongo.CreateBundle(ctx, b); err != nil {
			return nil, err
		}
	}

	return bundles, nil
}

func TestListBundles_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		mockBundles, err := setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When ListBundles is called with nil filters", func() {
			bundles, totalCount, err := mongodb.ListBundles(ctx, 0, 10, nil)

			Convey("Then it should return the correct bundles and total count", func() {
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 3)
				So(len(bundles), ShouldEqual, len(mockBundles))

				So(bundles[0].ID, ShouldEqual, "bundle1")
				So(bundles[0].BundleType, ShouldEqual, models.BundleTypeScheduled)
				So(bundles[1].ID, ShouldEqual, "bundle2")
				So(bundles[1].BundleType, ShouldEqual, models.BundleTypeManual)
				So(bundles[2].ID, ShouldEqual, mockBundles[2].ID)
				So(bundles[2].BundleType, ShouldEqual, mockBundles[2].BundleType)
			})
		})

		Convey("When ListBundles is called with valid filters", func() {
			Convey("Then it should return the matching correct bundles and total count when matching bundles", func() {
				expectedBundle := mockBundles[0]
				scheduledAtDate := expectedBundle.ScheduledAt

				filters := filters.BundleFilters{
					PublishDate: scheduledAtDate,
				}

				bundles, totalCount, err := mongodb.ListBundles(ctx, 0, 10, &filters)

				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 1)
				So(len(bundles), ShouldEqual, 1)

				So(bundles[0].ID, ShouldEqual, expectedBundle.ID)
				So(bundles[0].BundleType, ShouldEqual, models.BundleTypeScheduled)
			})

			Convey("Then it should return no bundles when no bundles matching", func() {
				scheduledAtDate := time.Now()

				filters := filters.BundleFilters{
					PublishDate: &scheduledAtDate,
				}

				bundles, totalCount, err := mongodb.ListBundles(ctx, 0, 10, &filters)

				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 0)
				So(len(bundles), ShouldEqual, 0)
			})
		})
	})
}

func TestListBundles_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When ListBundles is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			bundles, totalCount, err := mongodb.ListBundles(ctx, 0, 10, nil)

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

	Convey("When we call buildListBundlesQuery with a nil filter", t, func() {
		filter, sort := buildListBundlesQuery(nil)

		Convey("Then it should return an empty filter and sort by updated_at descending", func() {
			expectedFilter := bson.M{}
			expectedSort := bson.M{"updated_at": -1}

			So(filter, ShouldResemble, expectedFilter)
			So(sort, ShouldResemble, expectedSort)
		})
	})

	Convey("When we call buildListBundlesQuery with a non-nil filter", t, func() {
		Convey("Then it should return an empty filter if publishDate is nil", func() {
			bundleFilters := filters.BundleFilters{
				PublishDate: nil,
			}
			filter, sort := buildListBundlesQuery(&bundleFilters)

			expectedFilter := bson.M{}
			expectedSort := bson.M{"updated_at": -1}

			So(filter, ShouldResemble, expectedFilter)
			So(sort, ShouldResemble, expectedSort)
		})

		Convey("Then it should return an appropriate scheduled_at filter if publishDate is not nil", func() {
			publishDateFilter := time.Date(2025, 01, 01, 10, 30, 30, 0, time.UTC)
			bundleFilters := filters.BundleFilters{
				PublishDate: &publishDateFilter,
			}

			filter, sort := buildListBundlesQuery(&bundleFilters)

			scheduledAtFilter := bson.M{
				"$gte": publishDateFilter.Add(time.Second * -2),
				"$lte": publishDateFilter.Add(time.Second * 2),
			}

			expectedFilter := bson.M{
				"scheduled_at": scheduledAtFilter,
			}
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

		_, err = setupBundleTestData(ctx, mongodb)
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

		_, err = setupBundleTestData(ctx, mongodb)
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

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CreateBundle is called with a new bundle", func() {
			newBundle := &models.Bundle{
				ID:           "NewBundle",
				BundleType:   models.BundleTypeManual,
				PreviewTeams: &[]models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
				Title:        "New Bundle",
				ManagedBy:    models.ManagedByWagtail,
				ETag:         "some-etag",
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
				So(returnedBundle.ETag, ShouldEqual, "some-etag")
			})
		})
	})
}

func TestCreateBundle_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, mimServer, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		err = SetupIndexes(ctx, mimServer)
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

		_, err = setupBundleTestData(ctx, mongodb)
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

		_, err = setupBundleTestData(ctx, mongodb)
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

func TestUpdateBundleETag_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When UpdateBundleETag is called with an existing bundle ID", func() {
			oldBundle, err := mongodb.GetBundle(ctx, "bundle1")
			So(err, ShouldBeNil)

			bundleUpdate, err := mongodb.UpdateBundleETag(ctx, "bundle1", "new-email")

			Convey("Then it should update the ETag, last_updated_by, and updated_at fields without error", func() {
				So(err, ShouldBeNil)
				So(bundleUpdate.ETag, ShouldNotEqual, oldBundle.ETag)
				So(bundleUpdate.LastUpdatedBy.Email, ShouldEqual, "new-email")
				So(bundleUpdate.UpdatedAt, ShouldNotEqual, oldBundle.UpdatedAt)
			})
		})
	})
}

func TestUpdateBundleETag_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When UpdateBundleETag is called with a non-existent bundle ID", func() {
			_, err := mongodb.UpdateBundleETag(ctx, "non-existent-id", "new-email")

			Convey("Then it should return a bundle not found error", func() {
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
			})
		})
	})
}

func TestDeleteBundle_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
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

		_, err = setupBundleTestData(ctx, mongodb)
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

func TestCheckBundleExists_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckBundleExists is called with an existing bundle ID", func() {
			exists, err := mongodb.CheckBundleExists(ctx, "bundle1")

			Convey("Then it should return true without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
			})
		})

		Convey("When CheckBundleExists is called with a non-existent bundle ID", func() {
			exists, err := mongodb.CheckBundleExists(ctx, "non-existent-id")

			Convey("Then it should return false without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestCheckBundleExists_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		_, err = setupBundleTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckBundleExists is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			_, err := mongodb.CheckBundleExists(ctx, "bundle1")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func (m *Mongo) TestUpdateBundleState(ctx context.Context, bundleID string, state models.BundleState) error {
	bundle, err := m.GetBundle(ctx, bundleID)

	if err != nil {
		return err
	}
	bundle.State = &state
	filter := bson.M{"id": bundleID}

	etag, err := getEtagForBundle(bundle)
	if err != nil {
		return err
	}
	updateData := bson.M{
		"$set": bson.M{
			"e_tag": etag,
			"state": state,
		},
	}

	collectionName := m.ActualCollectionName(config.BundlesCollection)

	_, err = m.Connection.Collection(collectionName).UpdateOne(ctx, filter, updateData)
	if err != nil {
		return err
	}

	return nil
}

func TestGetETagForBundle(bundle *models.Bundle) (*string, error) {
	bundleUpdateJSON, err := json.Marshal(&bundle)
	if err != nil {
		return nil, err
	}

	etag := dpresponse.GenerateETag(bundleUpdateJSON, false)
	return &etag, nil
}
