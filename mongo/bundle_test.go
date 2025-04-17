package mongo

import (
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

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
