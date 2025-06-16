package events

import (
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-permissions-api/sdk"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRequestedBy(t *testing.T) {
	request := http.Request{}

	Convey("If parse is success", t, func() {
		mockEntityData := sdk.EntityData{
			UserID: "1234",
		}

		eventsManager := CreateTestBundleEvents(&mockEntityData)

		Convey("Should return RequestedBy with values from JWT", func() {
			expectedResult := &models.RequestedBy{
				ID:    mockEntityData.UserID,
				Email: mockEntityData.UserID,
			}
			requestedBy, err := eventsManager.createReqestedBy(&request)

			So(err, ShouldBeNil)
			So(requestedBy, ShouldEqual, expectedResult)
		})
	})

	Convey("If parse errors", t, func() {
		eventsManager := CreateTestBundleEvents(nil)

		Convey("Should return error", func() {
			requestedBy, err := eventsManager.createReqestedBy(&request)

			So(requestedBy, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCreateBundleResourceLocation(t *testing.T) {
	Convey("CreateBundleResourceLocation should return correct route", t, func() {
		bundle := models.Bundle{
			ID: "bundle-1234",
		}

		expectedResult := "/bundle/bundle-1234"

		location := createBundleResourceLocation(&bundle)

		So(location, ShouldEqual, expectedResult)
	})
}

func TestCreateBundleContentResourceLocation(t *testing.T) {
	Convey("CreateBundleContentResourceLocation should return correct route", t, func() {
		contentItem := models.ContentItem{
			ID:       "content-item-1234",
			BundleID: "bundle-999",
		}

		expectedResult := "/bundle/bundle-999/content/content-item-1234"

		location := createBundleContentResourceLocation(&contentItem)

		So(location, ShouldEqual, expectedResult)
	})
}
