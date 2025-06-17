package models

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	today     = time.Now()
	yesterday = today.Add(-24 * time.Hour)
	tomorrow  = today.Add(24 * time.Hour)
)

var bundleStateDraft = BundleStateDraft
var bundleStateInvalid = BundleState("Invalid")

var fullyPopulatedEvent = Event{
	CreatedAt: &today,
	RequestedBy: &RequestedBy{
		ID:    "user123",
		Email: "user123@ons.gov.uk",
	},
	Action:   ActionCreate,
	Resource: "/bundles/123/contents/item1",
	ContentItem: &ContentItem{
		ID:          "item1",
		BundleID:    "bundle123",
		ContentType: ContentTypeDataset,
		Metadata: Metadata{
			DatasetID: "dataset123",
			EditionID: "edition123",
			Title:     "Test Dataset",
			VersionID: 1,
		},
	},
	Bundle: &EventBundle{
		ID:         "bundle123",
		BundleType: BundleTypeManual,
		CreatedBy: &User{
			Email: "user123@ons.gov.uk",
		},
		CreatedAt: &yesterday,
		LastUpdatedBy: &User{
			Email: "user123@ons.gov.uk",
		},
		PreviewTeams: &[]PreviewTeam{
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
		ManagedBy:   ManagedByDataAdmin,
	},
}

var minimallyPopulatedEvent = Event{
	Action:   ActionCreate,
	Resource: "/bundles/123",
	Bundle: &EventBundle{
		BundleType: BundleTypeManual,
		CreatedBy: &User{
			Email: "user123@ons.gov.uk",
		},
		PreviewTeams: &[]PreviewTeam{
			{
				ID: "team1",
			},
			{
				ID: "team2",
			},
		},
		Title:     "Test Bundle",
		ManagedBy: ManagedByDataAdmin,
	},
}

func TestCreateEvent_Success(t *testing.T) {
	Convey("Given an event with all fields populated", t, func() {
		Convey("When CreateEvent is called", func() {
			b, err := json.Marshal(fullyPopulatedEvent)
			So(err, ShouldBeNil)

			reader := bytes.NewReader(b)
			event, err := CreateEvent(reader)

			Convey("Then it should return the event without any errors", func() {
				So(err, ShouldBeNil)
				So(event, ShouldNotBeNil)
				So(event.CreatedAt.Equal(*fullyPopulatedEvent.CreatedAt), ShouldBeTrue)
				So(event.RequestedBy, ShouldResemble, fullyPopulatedEvent.RequestedBy)
				So(event.Action, ShouldEqual, fullyPopulatedEvent.Action)
				So(event.Resource, ShouldEqual, fullyPopulatedEvent.Resource)
				So(event.ContentItem, ShouldResemble, fullyPopulatedEvent.ContentItem)
				So(event.Bundle.ID, ShouldEqual, fullyPopulatedEvent.Bundle.ID)
				So(event.Bundle.BundleType, ShouldEqual, fullyPopulatedEvent.Bundle.BundleType)
				So(event.Bundle.CreatedBy, ShouldResemble, fullyPopulatedEvent.Bundle.CreatedBy)
				So(event.Bundle.CreatedAt.Equal(*fullyPopulatedEvent.Bundle.CreatedAt), ShouldBeTrue)
				So(event.Bundle.LastUpdatedBy, ShouldResemble, fullyPopulatedEvent.Bundle.LastUpdatedBy)
				So(event.Bundle.PreviewTeams, ShouldResemble, fullyPopulatedEvent.Bundle.PreviewTeams)
				So(event.Bundle.ScheduledAt.Equal(*fullyPopulatedEvent.Bundle.ScheduledAt), ShouldBeTrue)
				So(event.Bundle.State, ShouldEqual, fullyPopulatedEvent.Bundle.State)
				So(event.Bundle.Title, ShouldEqual, fullyPopulatedEvent.Bundle.Title)
				So(event.Bundle.UpdatedAt.Equal(*fullyPopulatedEvent.Bundle.UpdatedAt), ShouldBeTrue)
				So(event.Bundle.ManagedBy, ShouldEqual, fullyPopulatedEvent.Bundle.ManagedBy)
			})
		})
	})

	Convey("Given an event with only mandatory fields populated", t, func() {
		Convey("When CreateEvent is called", func() {
			b, err := json.Marshal(minimallyPopulatedEvent)
			So(err, ShouldBeNil)

			reader := bytes.NewReader(b)
			event, err := CreateEvent(reader)

			Convey("Then it should return the event without any errors", func() {
				So(err, ShouldBeNil)
				So(event, ShouldResemble, &minimallyPopulatedEvent)
				So(event.Bundle, ShouldResemble, minimallyPopulatedEvent.Bundle)
			})
		})
	})
}

func TestCreateEvent_Failure(t *testing.T) {
	Convey("Given an io.Reader that returns an error", t, func() {
		errorReader := &ErrorReader{}

		Convey("When CreateEvent is called", func() {
			_, err := CreateEvent(errorReader)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errs.ErrUnableToReadMessage.Error())
			})
		})
	})

	Convey("Given an invalid JSON string", t, func() {
		invalidJSON := `{invalue_json}`
		reader := bytes.NewReader([]byte(invalidJSON))

		Convey("When CreateEvent is called", func() {
			_, err := CreateEvent(reader)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to parse json body")
			})
		})
	})
}

func TestValidateEvent_Success(t *testing.T) {
	Convey("Given an event with all mandatory fields populated", t, func() {
		Convey("When ValidateEvent is called", func() {
			err := ValidateEvent(&minimallyPopulatedEvent)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestValidateEvent_Failure(t *testing.T) {
	Convey("Given an event without any fields populated", t, func() {
		invalidEvent := Event{}

		Convey("When ValidateEvent is called", func() {
			err := ValidateEvent(&invalidEvent)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "missing mandatory fields: [action resource]")
			})
		})
	})

	Convey("Given an event without all mandatory fields (RequestedBy must exist to validate RequestedBy.ID)", t, func() {
		invalidEvent := Event{RequestedBy: &RequestedBy{}}

		Convey("When ValidateEvent is called", func() {
			err := ValidateEvent(&invalidEvent)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "missing mandatory fields: [requested_by.id action resource]")
			})
		})
	})

	Convey("Given an event with invalid fields", t, func() {
		invalidEvent := fullyPopulatedEvent
		invalidEvent.Action = Action("INVALID")

		Convey("When ValidateEvent is called", func() {
			err := ValidateEvent(&invalidEvent)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid fields: [action]")
			})
		})
	})
}

func TestAction_IsValid_Success(t *testing.T) {
	Convey("When given a valid Action", t, func() {
		validActions := []Action{
			ActionCreate,
			ActionRead,
			ActionUpdate,
			ActionDelete,
		}

		Convey("Then IsValid should return true", func() {
			for _, action := range validActions {
				So(action.IsValid(), ShouldBeTrue)
			}
		})
	})
}

func TestAction_IsValid_Failure(t *testing.T) {
	Convey("When given an invalid Action", t, func() {
		invalidAction := Action("INVALID")

		Convey("Then IsValid should return false", func() {
			So(invalidAction.IsValid(), ShouldBeFalse)
		})
	})
}

func TestConvertBundleToBundleEvent_Success(t *testing.T) {
	Convey("Given a Bundle object with all fields populated", t, func() {
		createdAt := time.Now()
		updatedAt := createdAt.Add(1 * time.Hour)
		scheduledAt := createdAt.Add(2 * time.Hour)

		bundle := &Bundle{
			ID:         "bundle123",
			BundleType: BundleTypeScheduled,
			CreatedBy: &User{
				Email: "user@ons.gov.uk",
			},
			CreatedAt: &createdAt,
			LastUpdatedBy: &User{
				Email: "user@ons.gov.uk",
			},
			PreviewTeams: &[]PreviewTeam{
				{ID: "team1"},
				{ID: "team2"},
			},
			ScheduledAt: &scheduledAt,
			State:       &bundleStateDraft,
			Title:       "Test Bundle",
			UpdatedAt:   &updatedAt,
			ManagedBy:   ManagedByDataAdmin,
		}

		Convey("When ConvertBundleToBundleEvent is called", func() {
			eventBundle, err := ConvertBundleToBundleEvent(bundle)

			Convey("Then it should return no error and an EventBundle with matching fields", func() {
				So(err, ShouldBeNil)
				So(eventBundle.ID, ShouldEqual, bundle.ID)
				So(eventBundle.BundleType, ShouldEqual, bundle.BundleType)
				So(eventBundle.CreatedBy, ShouldResemble, bundle.CreatedBy)
				So(eventBundle.CreatedAt.Equal(*bundle.CreatedAt), ShouldBeTrue)
				So(eventBundle.LastUpdatedBy, ShouldResemble, bundle.LastUpdatedBy)
				So(eventBundle.PreviewTeams, ShouldResemble, bundle.PreviewTeams)
				So(eventBundle.ScheduledAt.Equal(*bundle.ScheduledAt), ShouldBeTrue)
				So(eventBundle.State, ShouldEqual, bundle.State)
				So(eventBundle.Title, ShouldEqual, bundle.Title)
				So(eventBundle.UpdatedAt.Equal(*bundle.UpdatedAt), ShouldBeTrue)
				So(eventBundle.ManagedBy, ShouldEqual, bundle.ManagedBy)
			})
		})
	})
}

func TestConvertBundleToBundleEvent_Failure(t *testing.T) {
	Convey("Given a nil Bundle object", t, func() {
		var bundle *Bundle

		Convey("When ConvertBundleToBundleEvent is called", func() {
			eventBundle, err := ConvertBundleToBundleEvent(bundle)

			Convey("Then it should return an error and a nil EventBundle", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "input bundle cannot be nil")
				So(eventBundle, ShouldBeNil)
			})
		})
	})
}
