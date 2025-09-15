package models

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateEventModel(t *testing.T) {
	Convey("Given valid parameters for creating an Event", t, func() {
		id := "user-id"
		email := "user@example.com"
		action := ActionCreate
		bundle := fullyPopulatedBundle
		contentItem := fullyPopulatedContentItem

		Convey("When only a Bundle is provided", func() {
			event, err := CreateEventModel(id, email, action, &bundle, nil)

			Convey("Then the returned Event contains the expected values", func() {
				So(err, ShouldBeNil)
				So(event.RequestedBy.ID, ShouldEqual, id)
				So(event.RequestedBy.Email, ShouldEqual, email)
				So(event.Action, ShouldEqual, action)
				So(event.Resource, ShouldEqual, "/bundles/"+bundle.ID)
				So(event.Bundle, ShouldResemble, &bundle)
				So(event.ContentItem, ShouldBeNil)
			})
		})

		Convey("When only a ContentItem is provided", func() {
			event, err := CreateEventModel(id, email, action, nil, &contentItem)

			Convey("Then the returned Event contains the expected values", func() {
				So(err, ShouldBeNil)
				So(event.RequestedBy.ID, ShouldEqual, id)
				So(event.RequestedBy.Email, ShouldEqual, email)
				So(event.Action, ShouldEqual, action)
				So(event.Resource, ShouldEqual, "/bundles/"+contentItem.BundleID+"/contents/"+contentItem.ID)
				So(event.ContentItem, ShouldResemble, &contentItem)
				So(event.Bundle, ShouldBeNil)
			})
		})

		Convey("When both a Bundle and a ContentItem are provided", func() {
			event, err := CreateEventModel(id, email, action, &bundle, &contentItem)

			Convey("Then the expected error is returned", func() {
				So(event, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "only one of bundle or contentItem must be provided")
			})
		})

		Convey("When neither a Bundle or ContentItem are provided", func() {
			event, err := CreateEventModel(id, email, action, nil, nil)

			Convey("Then the expected error is returned", func() {
				So(event, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "only one of bundle or contentItem must be provided")
			})
		})
	})
}
