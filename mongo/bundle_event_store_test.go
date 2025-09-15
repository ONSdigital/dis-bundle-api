package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	today = time.Now()

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
	}
)

func TestCreateEvent_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupTestDataForEvents(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CreateEvent is called with a new event", func() {
			err := mongodb.CreateEvent(ctx, bundleEvent)

			Convey("Then it should create the event successfully without error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCreateEvent_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupTestDataForEvents(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CreateEvent is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			err := mongodb.CreateEvent(ctx, bundleEvent)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func setupTestDataForEvents(ctx context.Context, mongo *Mongo) error {
	if err := mongo.Connection.DropDatabase(ctx); err != nil {
		return err
	}

	approved := models.StateApproved
	events := []*models.Event{
		{
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

	for _, event := range events {
		if err := mongo.CreateEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
