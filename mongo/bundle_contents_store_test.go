package mongo

import (
	"context"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	stateApproved    = models.StateApproved
	contentsTestData = []*models.ContentItem{
		{
			ID:          "f3ee8348-9956-44e1-9c83-55fd2d7b2fb1",
			BundleID:    "bundle1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset1",
				EditionID: "2025",
				VersionID: 1,
				Title:     "Test Dataset 1",
			},
			State: &stateApproved,
			Links: models.Links{
				Edit:    "/edit/datasets/dataset1/editions/2025/versions/1",
				Preview: "/preview/datasets/dataset1/editions/2025/versions/1",
			},
		},
		{
			ID:          "af8b48b0-d085-4ea7-8f12-524fa8e6b0a0",
			BundleID:    "bundle1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset2",
				EditionID: "2025",
				VersionID: 1,
				Title:     "Test Dataset 2",
			},
			State: &stateApproved,
			Links: models.Links{
				Edit:    "/edit/datasets/dataset2/editions/2025/versions/1",
				Preview: "/preview/datasets/dataset2/editions/2025/versions/1",
			},
		},
		{
			ID:          "e784dfe1-2026-421c-a184-0e5a4c551019",
			BundleID:    "bundle2",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset3",
				EditionID: "2025",
				VersionID: 1,
				Title:     "Test Dataset 3",
			},
			Links: models.Links{
				Edit:    "/edit/datasets/dataset3/editions/2025/versions/1",
				Preview: "/preview/datasets/dataset3/editions/2025/versions/1",
			},
		},
	}
)

func setupBundleContentsTestData(ctx context.Context, mongodb *Mongo) error {
	for _, data := range contentsTestData {
		if _, err := mongodb.Connection.Collection(mongodb.ActualCollectionName(config.BundleContentsCollection)).InsertOne(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

func TestGetContentItemByBundleIDAndContentItemID_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When GetContentItemByBundleIDAndContentItemID is called with a valid bundleID and ContentItemID", func() {
			contentItemID := contentsTestData[0].ID
			bundleID := contentsTestData[0].BundleID
			contentItem, err := mongodb.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)

			Convey("Then it returns the content item without error", func() {
				So(err, ShouldBeNil)
				So(contentItem, ShouldNotBeNil)
				So(contentItem, ShouldResemble, contentsTestData[0])
			})
		})
	})
}

func TestGetContentItemByBundleIDAndContentItemID_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When GetContentItemByBundleIDAndContentItemID is called with a non-existent content item ID", func() {
			contentItemID := "non-existent-id"
			bundleID := contentsTestData[0].BundleID
			contentItem, err := mongodb.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)

			Convey("Then it returns a content item not found error", func() {
				So(err, ShouldNotBeNil)
				So(contentItem, ShouldBeNil)
				So(err, ShouldEqual, apierrors.ErrContentItemNotFound)
			})
		})

		Convey("When GetContentItemByBundleIDAndContentItemID is called with a non-existent bundle ID", func() {
			contentItemID := contentsTestData[0].ID
			bundleID := "non-existent-bundle"
			contentItem, err := mongodb.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)

			Convey("Then it returns a content item not found error", func() {
				So(err, ShouldNotBeNil)
				So(contentItem, ShouldBeNil)
				So(err, ShouldEqual, apierrors.ErrContentItemNotFound)
			})
		})

		Convey("When GetContentItemByBundleIDAndContentItemID is called and the connection fails", func() {
			mongodb.Connection.Close(ctx)
			contentItemID := contentsTestData[0].ID
			bundleID := contentsTestData[0].BundleID
			contentItem, err := mongodb.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
				So(contentItem, ShouldBeNil)
				So(err, ShouldNotEqual, apierrors.ErrContentItemNotFound)
			})
		})
	})
}

func TestCreateContentItem_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CreateContentItem is called with a valid content item", func() {
			stateApproved := models.StateApproved
			contentItem := &models.ContentItem{
				ID:          "12345678-1234-5678-1234-567812345678",
				BundleID:    "bundle3",
				ContentType: models.ContentTypeDataset,
				Metadata: models.Metadata{
					DatasetID: "dataset4",
					EditionID: "2025",
					VersionID: 1,
					Title:     "Test Dataset 4",
				},
				State: &stateApproved,
				Links: models.Links{
					Edit:    "/edit/datasets/dataset4/editions/2025/versions/1",
					Preview: "/preview/datasets/dataset4/editions/2025/versions/1",
				},
			}

			err := mongodb.CreateContentItem(ctx, contentItem)

			Convey("Then it returns no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCreateContentItem_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, minServer, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = SetupIndexes(ctx, minServer)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CreateContentItem is called with an existing ID", func() {
			stateApproved := models.StateApproved
			contentItem := &models.ContentItem{
				ID:          "f3ee8348-9956-44e1-9c83-55fd2d7b2fb1",
				BundleID:    "bundle3",
				ContentType: models.ContentTypeDataset,
				Metadata: models.Metadata{
					DatasetID: "dataset4",
					EditionID: "2025",
					VersionID: 1,
					Title:     "Test Dataset 4",
				},
				State: &stateApproved,
				Links: models.Links{
					Edit:    "/edit/datasets/dataset4/editions/2025/versions/1",
					Preview: "/preview/datasets/dataset4/editions/2025/versions/1",
				},
			}

			err := mongodb.CreateContentItem(ctx, contentItem)

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "duplicate key error")
			})
		})
	})
}

func TestCheckAllBundleContentsAreApproved_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckAllBundleContentsAreApproved is called and all contents are approved", func() {
			isApproved, err := mongodb.CheckAllBundleContentsAreApproved(ctx, "bundle1")

			Convey("Then it returns true without error", func() {
				So(err, ShouldBeNil)
				So(isApproved, ShouldBeTrue)
			})
		})

		Convey("When CheckAllBundleContentsAreApproved is called and not all contents are approved", func() {
			isApproved, err := mongodb.CheckAllBundleContentsAreApproved(ctx, "bundle2")

			Convey("Then it returns false without error", func() {
				So(err, ShouldBeNil)
				So(isApproved, ShouldBeFalse)
			})
		})
	})
}

func TestCheckAllBundleContentsAreApproved_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckAllBundleContentsAreApproved is called and the connection is closed", func() {
			mongodb.Connection.Close(ctx)
			_, err := mongodb.CheckAllBundleContentsAreApproved(ctx, "bundle1")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "client is disconnected")
			})
		})
	})
}

func TestCheckContentItemExistsByDatasetEditionVersion_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckContentItemExistsByDatasetEditionVersion is called with an existing dataset edition version", func() {
			exists, err := mongodb.CheckContentItemExistsByDatasetEditionVersion(ctx, "dataset1", "2025", 1)

			Convey("Then it returns true without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
			})
		})

		Convey("When CheckContentItemExistsByDatasetEditionVersion is called with a non-existent dataset edition version", func() {
			exists, err := mongodb.CheckContentItemExistsByDatasetEditionVersion(ctx, "dataset4", "2025", 1)

			Convey("Then it returns false without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestCheckContentItemExistsByDatasetEditionVersion_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckContentItemExistsByDatasetEditionVersion is called and the connection is closed", func() {
			mongodb.Connection.Close(ctx)
			_, err := mongodb.CheckContentItemExistsByDatasetEditionVersion(ctx, "dataset1", "2025", 1)

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "client is disconnected")
			})
		})
	})
}

func TestDeleteContentItem_Success(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When DeleteContentItem is called with a valid content item ID", func() {
			contentItemID := contentsTestData[0].ID
			bundleID := contentsTestData[0].BundleID
			err := mongodb.DeleteContentItem(ctx, contentItemID)

			Convey("Then it deletes the content item without error", func() {
				So(err, ShouldBeNil)

				deletedContentItem, err := mongodb.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)
				So(err, ShouldNotBeNil)
				So(deletedContentItem, ShouldBeNil)
				So(err, ShouldEqual, apierrors.ErrContentItemNotFound)
			})
		})
	})
}

func TestDeleteContentItem_Failure(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, _, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When DeleteContentItem is called with a non-existent content item ID", func() {
			contentItemID := "non-existent-id"
			err := mongodb.DeleteContentItem(ctx, contentItemID)

			Convey("Then it returns a content item not found error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, apierrors.ErrContentItemNotFound)
			})
		})

		Convey("When DeleteContentItem is called and the connection is closed", func() {
			mongodb.Connection.Close(ctx)
			contentItemID := contentsTestData[0].ID
			err := mongodb.DeleteContentItem(ctx, contentItemID)

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldNotEqual, apierrors.ErrContentItemNotFound)
			})
		})
	})
}
