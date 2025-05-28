package mongo

import (
	"context"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func setupBundleContentsTestData(ctx context.Context, mongodb *Mongo) error {
	testData := []map[string]interface{}{
		{
			"_id":          "f3ee8348-9956-44e1-9c83-55fd2d7b2fb1",
			"bundle_id":    "bundle1",
			"content_type": "DATASET",
			"metadata": map[string]interface{}{
				"dataset_id": "dataset1",
				"edition_id": "2025",
				"version_id": 1,
				"title":      "Test Dataset 1",
			},
			"state": "APPROVED",
			"links": map[string]interface{}{
				"edit":    "/edit/datasets/dataset1/editions/2025/versions/1",
				"preview": "/preview/datasets/dataset1/editions/2025/versions/1",
			},
		},
		{
			"_id":          "af8b48b0-d085-4ea7-8f12-524fa8e6b0a0",
			"bundle_id":    "bundle1",
			"content_type": "DATASET",
			"metadata": map[string]interface{}{
				"dataset_id": "dataset2",
				"edition_id": "2025",
				"version_id": 1,
				"title":      "Test Dataset 2",
			},
			"state": "APPROVED",
			"links": map[string]interface{}{
				"edit":    "/edit/datasets/dataset2/editions/2025/versions/1",
				"preview": "/preview/datasets/dataset2/editions/2025/versions/1",
			},
		},
		{
			"_id":          "e784dfe1-2026-421c-a184-0e5a4c551019",
			"bundle_id":    "bundle2",
			"content_type": "DATASET",
			"metadata": map[string]interface{}{
				"dataset_id": "dataset3",
				"edition_id": "2025",
				"version_id": 1,
				"title":      "Test Dataset 3",
			},
			"links": map[string]interface{}{
				"edit":    "/edit/datasets/dataset3/editions/2025/versions/1",
				"preview": "/preview/datasets/dataset3/editions/2025/versions/1",
			},
		},
	}

	for _, data := range testData {
		if _, err := mongodb.Connection.Collection(mongodb.ActualCollectionName(config.BundleContentsCollection)).InsertOne(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

func TestCheckAllBundleContentsAreApproved(t *testing.T) {
	ctx := context.Background()

	Convey("Given the db connection is initialized correctly", t, func() {
		mongodb, err := getTestMongoDB(ctx)
		So(err, ShouldBeNil)

		err = setupBundleContentsTestData(ctx, mongodb)
		So(err, ShouldBeNil)

		Convey("When CheckAllBundleContentsAreApproved is called", func() {
			Convey("And all contents are approved", func() {
				isApproved, err := mongodb.CheckAllBundleContentsAreApproved(ctx, "bundle1")

				Convey("Then it returns true without error", func() {
					So(err, ShouldBeNil)
					So(isApproved, ShouldBeTrue)
				})
			})

			Convey("And some contents are not approved", func() {
				isApproved, err := mongodb.CheckAllBundleContentsAreApproved(ctx, "bundle2")

				Convey("Then it returns false without error", func() {
					So(err, ShouldBeNil)
					So(isApproved, ShouldBeFalse)
				})
			})

			Convey("And function returns an error", func() {
				mongodb.Connection.Close(ctx)
				_, err := mongodb.CheckAllBundleContentsAreApproved(ctx, "bundle1")

				Convey("Then it returns an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "client is disconnected")
				})
			})
		})
	})
}
