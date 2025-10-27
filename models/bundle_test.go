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
	today    = time.Now()
	tomorrow = today.Add(24 * time.Hour)
)

var fullyPopulatedBundle = Bundle{
	ID:            "123",
	BundleType:    BundleTypeManual,
	CreatedBy:     &User{Email: "example@example.com"},
	CreatedAt:     &today,
	LastUpdatedBy: &User{Email: "example@example.com"},
	PreviewTeams:  &[]PreviewTeam{{ID: "team1"}, {ID: "team2"}},
	ScheduledAt:   &tomorrow,
	State:         BundleStateDraft,
	Title:         "Fully Populated Bundle",
	UpdatedAt:     &today,
	ManagedBy:     ManagedByWagtail,
	ETag:          "f9226b8eb338ac139b1c39d2bb69f5abad8bea09",
}

var minimallyPopulatedBundle = Bundle{
	ID:           "456",
	BundleType:   BundleTypeManual,
	PreviewTeams: &[]PreviewTeam{{ID: "team1"}, {ID: "team2"}},
	State:        BundleStateDraft,
	Title:        "Minimally Populated Bundle",
	ManagedBy:    ManagedByWagtail,
	ETag:         "3c897c1081faa19bff0c20ffd1ca99cc54640f0e",
}

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a fully populated bundle", t, func() {
		b, err := json.Marshal(fullyPopulatedBundle)
		So(err, ShouldBeNil)

		reader := bytes.NewReader(b)

		Convey("When CreateBundle is called", func() {
			bundle, err := CreateBundle(reader, "example@example.com")
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
				So(bundle.ETag, ShouldNotBeEmpty)
			})
		})
	})

	Convey("Given a minimally populated bundle", t, func() {
		b, err := json.Marshal(minimallyPopulatedBundle)
		So(err, ShouldBeNil)

		reader := bytes.NewReader(b)

		Convey("When CreateBundle is called", func() {
			bundle, err := CreateBundle(reader, "example@example.com")
			So(err, ShouldBeNil)

			Convey("Then it should return a bundle with the expected values", func() {
				So(bundle.ID, ShouldNotBeEmpty)
				So(bundle.BundleType, ShouldEqual, minimallyPopulatedBundle.BundleType)
				So(bundle.CreatedBy.Email, ShouldEqual, "example@example.com")
				So(bundle.CreatedAt, ShouldBeNil)
				So(bundle.LastUpdatedBy.Email, ShouldEqual, "example@example.com")
				So(bundle.PreviewTeams, ShouldResemble, minimallyPopulatedBundle.PreviewTeams)
				So(bundle.ScheduledAt, ShouldBeNil)
				So(bundle.State, ShouldResemble, minimallyPopulatedBundle.State)
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
			_, err := CreateBundle(reader, "example@example.com")
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
			_, err := CreateBundle(reader, "example@example.com")
			Convey("Then it should return an unable to parse json error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
			})
		})
	})
}

func TestCleanBundle_Success(t *testing.T) {
	Convey("Given a Bundle with leading and trailing whitespace", t, func() {
		bundle := &Bundle{
			ID:            " 123 ",
			BundleType:    BundleType(" MANUAL "),
			CreatedBy:     &User{Email: " test@example.com "},
			LastUpdatedBy: &User{Email: " test@example.com "},
			PreviewTeams:  &[]PreviewTeam{{ID: " team1 "}, {ID: " team2 "}},
			State:         BundleState(" DRAFT "),
			Title:         " Bundle Title ",
			ManagedBy:     ManagedBy(" WAGTAIL "),
		}

		Convey("When CleanBundle is called", func() {
			CleanBundle(bundle)

			Convey("Then it should remove the whitespace from all fields", func() {
				So(bundle.ID, ShouldEqual, "123")
				So(bundle.BundleType, ShouldEqual, BundleTypeManual)
				So(bundle.CreatedBy.Email, ShouldEqual, "test@example.com")
				So(bundle.LastUpdatedBy.Email, ShouldEqual, "test@example.com")

				So((*bundle.PreviewTeams)[0].ID, ShouldEqual, " team1 ")
				So((*bundle.PreviewTeams)[1].ID, ShouldEqual, " team2 ")
				So(bundle.State, ShouldEqual, BundleStateDraft)
				So(bundle.Title, ShouldEqual, "Bundle Title")
				So(bundle.ManagedBy, ShouldEqual, ManagedByWagtail)
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
			CreatedBy:     &User{Email: ""},
			LastUpdatedBy: &User{Email: ""},
		}

		Convey("When ValidateBundle is called", func() {
			err := ValidateBundle(&bundle)

			Convey("Then it should return an error indicating the missing fields", func() {
				So(err, ShouldNotBeNil)
				So(err[0].Source.Field, ShouldEqual, "/id")
				So(err[1].Source.Field, ShouldEqual, "/bundle_type")
				So(err[2].Source.Field, ShouldEqual, "/created_by/email")
				So(err[3].Source.Field, ShouldEqual, "/last_updated_by/email")
				So(err[4].Source.Field, ShouldEqual, "/state")
				So(err[5].Source.Field, ShouldEqual, "/title")
				So(err[6].Source.Field, ShouldEqual, "/managed_by")
			})
		})
	})

	Convey("Given a bundle with invalid fields (State only checked if field exists)", t, func() {
		bundle := fullyPopulatedBundle
		bundle.BundleType = BundleType("invalid-type")
		bundle.State = BundleState("Invalid")
		bundle.ManagedBy = ManagedBy("invalid-managed-by")

		Convey("When ValidateBundle is called", func() {
			err := ValidateBundle(&bundle)

			Convey("Then it should return an error indicating the invalid fields", func() {
				So(err, ShouldNotBeNil)
				So(err[0].Source.Field, ShouldEqual, "/bundle_type")
				So(err[1].Source.Field, ShouldEqual, "/state")
				So(err[2].Source.Field, ShouldEqual, "/managed_by")
			})
		})
	})
}

func TestValidateBundle_Failure_MissingPreviewTeamsID(t *testing.T) {
	Convey("Given a bundle with missing PreviewTeams ID", t, func() {
		bundle := Bundle{
			ID:           "789",
			BundleType:   BundleTypeManual,
			PreviewTeams: &[]PreviewTeam{{ID: ""}},
			Title:        "Bundle with Missing Preview Teams ID",
			ManagedBy:    ManagedByWagtail,
		}

		Convey("When ValidateBundle is called", func() {
			err := ValidateBundle(&bundle)

			Convey("Then it should return an error indicating the id is missing", func() {
				So(err, ShouldNotBeNil)
				So(err[0].Source.Field, ShouldEqual, "/preview_teams/id")
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
