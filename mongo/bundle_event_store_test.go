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
)

var (
	today            = time.Now()
	yesterday        = today.Add(-24 * time.Hour)
	tomorrow         = today.Add(24 * time.Hour)
	bundleStateDraft = models.BundleStateDraft

	bundleEvent = &models.Event{
		CreatedAt: &today,
		RequestedBy: &models.RequestedBy{
			ID:    "user123",
			Email: "user123@ons.gov.uk",
		},
		Action:   models.ActionCreate,
		Resource: "/bundles/123/contents/item1",
		ContentItem: &models.ContentItem{
			ID:          "item1",
			BundleID:    "bundle123",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset123",
				EditionID: "edition123",
				Title:     "Test Dataset",
				VersionID: 1,
			},
		},
		Bundle: &models.Bundle{
			ID:         "bundle123",
			BundleType: models.BundleTypeManual,
			CreatedBy: &models.User{
				Email: "user123@ons.gov.uk",
			},
			CreatedAt: &yesterday,
			LastUpdatedBy: &models.User{
				Email: "user123@ons.gov.uk",
			},
			PreviewTeams: &[]models.PreviewTeam{
				{
					ID: "team1",
				},
				{
					ID: "team2",
				},
			},
			ScheduledAt: &tomorrow,
			State:       &bundleStateDraft,
			Title:       "Test Bundle",
			UpdatedAt:   &today,
			ManagedBy:   models.ManagedByDataAdmin,
		},
	}
)

func TestMongoCRUDForEvents(t *testing.T) {
	// Config for test mongo in memory instance
	var (
		ctx          = context.Background()
		mongoVersion = "4.4.8"
		mongoServer  *mim.Server
		err          error
	)

	// Get the default app config to use when setting up mongo in memory
	cfg, _ := config.Get()

	Convey("When the db connection is initialized correctly", t, func() {
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

		if err := setupTestDataForEvents(ctx, mongodb); err != nil {
			t.Fatalf("failed to insert test data for events, skipping tests: %v", err)
		}

		// GetBundleEvent
		Convey("When GetBundleEvent is called", func() {
			Convey("When the bundle event is not found", func() {
				_, err := mongodb.GetBundleEvent(ctx, "event-not-exist")
				So(err, ShouldEqual, apierrors.ErrBundleEventNotFound)
			})

			Convey("When GetBundleEvent returns a generic error", func() {
				mongodb.Connection.DropDatabase(ctx)
				_, err := mongodb.GetBundleEvent(ctx, "bundleEvent1")
				So(err, ShouldNotBeNil)
			})
		})

		// ListBundleEvents
		Convey("When ListBundleEvents is called", func() {
			Convey("When the bundle event list is fetched successfully", func() {
				returnedEvents, totalCount, err := mongodb.ListBundleEvents(ctx, 0, 10)
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 2)
				So(len(returnedEvents), ShouldEqual, 2)
				So(returnedEvents[0].ContentItem.ID, ShouldEqual, "content1")
				So(returnedEvents[1].ContentItem.ID, ShouldEqual, "content2")
			})

			Convey("When ListBundleEvents returns an error", func() {
				mongodb.Connection.Close(ctx)
				_, _, err := mongodb.ListBundleEvents(ctx, 0, 10)
				So(err, ShouldNotBeNil)
			})
		})

		// CreateBundleEvent
		Convey("When CreateBundleEvent is called", func() {
			Convey("When the bundle event is created successfully", func() {
				err := mongodb.CreateBundleEvent(ctx, bundleEvent)
				So(err, ShouldBeNil)
			})
		})

		// Update Bundle
		Convey("When the UpdateBundle is called", func() {
			Convey("When the bundle event is not found for update", func() {
				_, err := mongodb.UpdateBundleEvent(ctx, "event-not-exist", bundleEvent)
				So(err, ShouldEqual, apierrors.ErrBundleEventNotFound)
			})

			Convey("When UpdateBundleEvent returns a generic error", func() {
				mongodb.Connection.DropDatabase(ctx)
				_, err := mongodb.UpdateBundleEvent(ctx, "bundleEvent1", bundleEvent)
				So(err, ShouldNotBeNil)
			})
		})

		// Delete Bundle
		Convey("When DeleteBundleEvent is called", func() {
			Convey("When trying to delete a non-existent bundle event and returns error", func() {
				err := mongodb.DeleteBundle(ctx, "non-existent-id")
				So(err, ShouldResemble, apierrors.ErrBundleNotFound)
			})

			Convey("When trying to delete a bundle and returns generic error", func() {
				mongodb.Connection.Close(ctx)
				err := mongodb.DeleteBundleEvent(ctx, "bundleEvent1")
				So(err, ShouldResemble, err)
			})
		})
	})
}

func setupTestDataForEvents(ctx context.Context, mongo *Mongo) error {
	if err := mongo.Connection.DropDatabase(ctx); err != nil {
		return err
	}

	now := time.Now()
	approved := models.StateApproved
	bundleEvents := []*models.Event{
		{
			CreatedAt:   &now,
			RequestedBy: &models.RequestedBy{ID: "user1", Email: "user1@ons.gov.uk"},
			Action:      models.ActionCreate,
			Resource:    "content1",
			ContentItem: &models.ContentItem{
				ID:          "content1",
				BundleID:    "bundle1",
				ContentType: models.ContentTypeDataset,
				Metadata: models.Metadata{
					DatasetID: "dataset1",
					EditionID: "edition1",
					Title:     "Dataset 1",
					VersionID: 1,
				},
				State: &approved,
				Links: models.Links{
					Edit:    "edit_link",
					Preview: "preview_link",
				},
			},
		},
		{
			CreatedAt:   &now,
			RequestedBy: &models.RequestedBy{ID: "user2", Email: "user2@ons.gov.uk"},
			Action:      models.ActionCreate,
			Resource:    "content2",
			ContentItem: &models.ContentItem{
				ID:          "content2",
				BundleID:    "bundle2",
				ContentType: models.ContentTypeDataset,
				Metadata: models.Metadata{
					DatasetID: "dataset2",
					EditionID: "edition2",
					Title:     "Dataset 2",
					VersionID: 2,
				},
				State: &approved,
				Links: models.Links{
					Edit:    "edit_link2",
					Preview: "preview_link2",
				},
			},
		},
	}

	for _, event := range bundleEvents {
		if err := mongo.CreateBundleEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
