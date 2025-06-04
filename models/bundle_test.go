package models

import (
	"bytes"
	"encoding/json"
	"testing"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	. "github.com/smartystreets/goconvey/convey"
)

var fullyPopulatedBundle = Bundle{
	ID:            "123",
	BundleType:    BundleTypeManual,
	CreatedBy:     &User{Email: "example@example.com"},
	CreatedAt:     &today,
	LastUpdatedBy: &User{Email: "example@example.com"},
	PreviewTeams:  &[]PreviewTeam{{ID: "team1"}, {ID: "team2"}},
	ScheduledAt:   &tomorrow,
	State:         &bundleStateDraft,
	Title:         "Fully Populated Bundle",
	UpdatedAt:     &today,
	ManagedBy:     ManagedByWagtail,
}

var minimallyPopulatedBundle = Bundle{
	ID:           "456",
	BundleType:   BundleTypeManual,
	PreviewTeams: &[]PreviewTeam{{ID: "team1"}, {ID: "team2"}},
	Title:        "Minimally Populated Bundle",
	ManagedBy:    ManagedByWagtail,
}

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a fully populated bundle", t, func() {
		b, err := json.Marshal(fullyPopulatedBundle)
		So(err, ShouldBeNil)

		reader := bytes.NewReader(b)

		Convey("When CreateBundle is called", func() {
			bundle, err := CreateBundle(reader)
			So(err, ShouldBeNil)

			Convey("Then it should return a bundle with the expected values", func() {
				So(bundle.ID, ShouldNotBeEmpty)
				So(bundle.BundleType, ShouldEqual, fullyPopulatedBundle.BundleType)
				So(bundle.CreatedBy, ShouldResemble, fullyPopulatedBundle.CreatedBy)
				So(bundle.CreatedAt.Equal(*fullyPopulatedBundle.CreatedAt), ShouldBeTrue)
				So(bundle.LastUpdatedBy.Email, ShouldEqual, fullyPopulatedBundle.LastUpdatedBy.Email)
				So(bundle.PreviewTeams, ShouldResemble, fullyPopulatedBundle.PreviewTeams)
				So(bundle.ScheduledAt.Equal(*fullyPopulatedBundle.ScheduledAt), ShouldBeTrue)
				So(bundle.State, ShouldEqual, fullyPopulatedBundle.State)
				So(bundle.Title, ShouldEqual, fullyPopulatedBundle.Title)
				So(bundle.UpdatedAt.Equal(*fullyPopulatedBundle.UpdatedAt), ShouldBeTrue)
				So(bundle.ManagedBy, ShouldEqual, fullyPopulatedBundle.ManagedBy)
			})
		})
	})

	Convey("Given a minimally populated bundle", t, func() {
		b, err := json.Marshal(minimallyPopulatedBundle)
		So(err, ShouldBeNil)

		reader := bytes.NewReader(b)

		Convey("When CreateBundle is called", func() {
			bundle, err := CreateBundle(reader)
			So(err, ShouldBeNil)

			Convey("Then it should return a bundle with the expected values", func() {
				So(bundle.ID, ShouldNotBeEmpty)
				So(bundle.BundleType, ShouldEqual, minimallyPopulatedBundle.BundleType)
				So(bundle.CreatedBy, ShouldBeNil)
				So(bundle.CreatedAt, ShouldBeNil)
				So(bundle.LastUpdatedBy, ShouldBeNil)
				So(bundle.PreviewTeams, ShouldResemble, minimallyPopulatedBundle.PreviewTeams)
				So(bundle.ScheduledAt, ShouldBeNil)
				So(bundle.State, ShouldBeNil)
				So(bundle.Title, ShouldEqual, minimallyPopulatedBundle.Title)
				So(bundle.UpdatedAt, ShouldBeNil)
				So(bundle.ManagedBy, ShouldEqual, minimallyPopulatedBundle.ManagedBy)
			})
		})
	})
}

func TestCreateBundle_Failure(t *testing.T) {
	Convey("Given a reader that fails to read", t, func() {
		reader := &ErrorReader{}
		Convey("When CreateBundle is called", func() {
			_, err := CreateBundle(reader)
			Convey("Then it should return an unable to read message error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errs.ErrUnableToReadMessage.Error())
			})
		})
	})

	Convey("Given a reader with invalid json", t, func() {
		b := `{"state":"invalid-body}`
		reader := bytes.NewReader([]byte(b))
		Convey("When CreateBundle is called", func() {
			_, err := CreateBundle(reader)
			Convey("Then it should return an unable to parse json error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
			})
		})
	})
}

func TestValidateBundle_Success(t *testing.T) {
	Convey("Given a minimally populated bundle", t, func() {
		Convey("When ValidateBundle is called", func() {
			err := ValidateBundle(&minimallyPopulatedBundle)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestValidateBundle_Failure(t *testing.T) {
	Convey("Given a bundle with missing mandatory fields (CreatedBy and LastUpdatedBy email only checked if field exists)", t, func() {
		bundle := Bundle{
			ID:            "",
			BundleType:    "",
			CreatedBy:     &User{Email: ""},
			LastUpdatedBy: &User{Email: ""},
			PreviewTeams:  &[]PreviewTeam{},
			Title:         "",
			ManagedBy:     "",
		}

		Convey("When ValidateBundle is called", func() {
			err := ValidateBundle(&bundle)

			Convey("Then it should return an error indicating the missing fields", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "missing mandatory fields: [id bundle_type created_by.email last_updated_by.email preview_teams title managed_by]")
			})
		})
	})

	Convey("Given a bundle with invalid fields (State only checked if field exists)", t, func() {
		bundle := fullyPopulatedBundle
		bundle.BundleType = BundleType("invalid-type")
		bundle.State = &bundleStateInvalid
		bundle.ManagedBy = ManagedBy("invalid-managed-by")

		Convey("When ValidateBundle is called", func() {
			err := ValidateBundle(&bundle)

			Convey("Then it should return an error indicating the invalid fields", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid fields: [bundle_type state managed_by]")
			})
		})
	})
}

func TestBundleState_IsValid_Success(t *testing.T) {
	Convey("Given a valid bundle state", t, func() {
		state := BundleStateDraft

		Convey("When IsValid is called", func() {
			valid := state.IsValid()

			Convey("Then it should return true", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})
}

func TestBundleState_IsValid_Failure(t *testing.T) {
	Convey("Given an invalid bundle state", t, func() {
		state := BundleState("invalid-state")

		Convey("When IsValid is called", func() {
			valid := state.IsValid()

			Convey("Then it should return false", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}

func TestBundleState_String_Success(t *testing.T) {
	Convey("Given a valid bundle state", t, func() {
		state := BundleStateDraft

		Convey("When String is called", func() {
			str := state.String()

			Convey("Then it should return the correct string representation", func() {
				So(str, ShouldEqual, "DRAFT")
			})
		})
	})
}
