package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	. "github.com/smartystreets/goconvey/convey"
)

const invalid = "INVALID"

func TestCreateBundle(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the bundle has all fields", func() {
			bundleID := "12345"
			user := User{
				Email: "example@example.com",
			}
			now := time.Now()
			previewTeam := PreviewTeam{
				ID: "team1",
			}
			state := BundleStateDraft
			testBundle := Bundle{
				ID:            bundleID,
				BundleType:    BundleTypeManual,
				CreatedBy:     &user,
				CreatedAt:     &now,
				LastUpdatedBy: &user,
				PreviewTeams: &[]PreviewTeam{
					previewTeam,
				},
				ScheduledAt: &now,
				State:       &state,
				Title:       "Test Bundle",
				UpdatedAt:   &now,
				ManagedBy:   ManagedByWagtail,
			}
			b, err := json.Marshal(testBundle)
			if err != nil {
				t.Logf("failed to marshal test data into bytes, error: %v", err)
				t.FailNow()
			}
			reader := bytes.NewReader(b)
			bundle, err := CreateBundle(reader)
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleID)
			So(bundle.BundleType, ShouldEqual, BundleTypeManual)
			So(bundle.CreatedBy, ShouldEqual, &user)
			So(bundle.CreatedAt.Equal(now), ShouldBeTrue)
			So(bundle.LastUpdatedBy, ShouldEqual, &user)
			So(bundle.PreviewTeams, ShouldEqual, &[]PreviewTeam{previewTeam})
			So(bundle.ScheduledAt.Equal(now), ShouldBeTrue)
			So(bundle.State, ShouldEqual, &state)
			So(bundle.Title, ShouldEqual, "Test Bundle")
			So(bundle.UpdatedAt.Equal(now), ShouldBeTrue)
			So(bundle.ManagedBy, ShouldEqual, ManagedByWagtail)
		})
	})

	Convey("Return error when unable to read message", t, func() {
		Convey("when the reader returns an error", func() {
			reader := &ErrorReader{}
			_, err := CreateBundle(reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToReadMessage.Error())
		})
	})

	Convey("Return error when unable to parse json", t, func() {
		Convey("when the json is invalid", func() {
			b := `{"state":"invalid-body}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateBundle(reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
		})
	})
}

func TestMarshalJSONForBundle(t *testing.T) {
	Convey("Given a bundle", t, func() {
		bundleID := "12345"
		user := User{
			Email: "example@example.com",
		}
		now := time.Now()
		previewTeam := PreviewTeam{
			ID: "team1",
		}
		state := BundleStateDraft
		testBundle := Bundle{
			ID:            bundleID,
			BundleType:    BundleTypeManual,
			CreatedBy:     &user,
			CreatedAt:     &now,
			LastUpdatedBy: &user,
			PreviewTeams: &[]PreviewTeam{
				previewTeam,
			},
			ScheduledAt: &now,
			State:       &state,
			Title:       "Test Bundle",
			UpdatedAt:   &now,
			ManagedBy:   ManagedByWagtail,
		}
		Convey("when the bundle type is invalid", func() {
			testBundle.BundleType = invalid
			Convey("then it should return an error", func() {
				_, err := json.Marshal(testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid BundleType")
			})
		})
		Convey("when the bundle state is invalid", func() {
			state = invalid
			testBundle.State = &state
			Convey("then it should return an error", func() {
				_, err := json.Marshal(testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid BundleState")
			})
		})
		Convey("when the managed by is invalid", func() {
			testBundle.ManagedBy = invalid
			Convey("then it should return an error", func() {
				_, err := json.Marshal(testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid ManagedBy")
			})
		})
	})
}

func TestUnmarshalJSON_InvalidBundleType(t *testing.T) {
	Convey("Given invalid JSON for BundleState", t, func() {
		invalidJSON := []byte(`123`) // Invalid JSON for a string
		var bandleType BundleType

		Convey("When UnmarshalJSON is called", func() {
			err := bandleType.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given invalid JSON for BundleType", t, func() {
		invalidJSON := []byte(`"INVALID"`) // Invalid value for ManagedBy
		var bundleType BundleType

		Convey("When UnmarshalJSON is called", func() {
			err := bundleType.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid BundleType: INVALID")
			})
		})
	})
}

func TestUnmarshalJSON_InvalidBundleState(t *testing.T) {
	Convey("Given invalid JSON for BundleState", t, func() {
		invalidJSON := []byte(`123`) // Invalid JSON for a string
		var bandleState BundleState

		Convey("When UnmarshalJSON is called", func() {
			err := bandleState.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given invalid JSON for BundleState", t, func() {
		invalidJSON := []byte(`"INVALID"`) // Invalid value for ManagedBy
		var bundleState BundleState

		Convey("When UnmarshalJSON is called", func() {
			err := bundleState.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid BundleState: INVALID")
			})
		})
	})
}

func TestUnmarshalJSON_InvalidManagedBy(t *testing.T) {
	Convey("Given invalid JSON for ManagedBy", t, func() {
		invalidJSON := []byte(`123`) // Invalid JSON for a string
		var managedBy ManagedBy

		Convey("When UnmarshalJSON is called", func() {
			err := managedBy.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given invalid JSON for ManagedBy", t, func() {
		invalidJSON := []byte(`"INVALID"`) // Invalid value for ManagedBy
		var managedBy ManagedBy

		Convey("When UnmarshalJSON is called", func() {
			err := managedBy.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid ManagedBy: INVALID")
			})
		})
	})
}

func TestValidateBundle(t *testing.T) {
	Convey("Given a valid bundle", t, func() {
		bundleID := "12345"
		user := User{
			Email: "example@example.com",
		}
		now := time.Now()
		previewTeam := PreviewTeam{
			ID: "team1",
		}
		state := BundleStateDraft
		testBundle := Bundle{
			ID:            bundleID,
			BundleType:    BundleTypeManual,
			CreatedBy:     &user,
			CreatedAt:     &now,
			LastUpdatedBy: &user,
			PreviewTeams: &[]PreviewTeam{
				previewTeam,
			},
			ScheduledAt: &now,
			State:       &state,
			Title:       "Test Bundle",
			UpdatedAt:   &now,
			ManagedBy:   ManagedByWagtail,
		}

		Convey("When Validate is called with a valid bundle", func() {
			err := ValidateBundle(&testBundle)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When Validate is called and BundleType is empty", func() {
			testBundle.BundleType = ""
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"bundle_type"}))
			})
		})

		Convey("When Validate is called and PreviewTeams is empty", func() {
			testBundle.PreviewTeams = &[]PreviewTeam{}
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"preview_teams"}))
			})
		})

		Convey("When Validate is called and Title is empty", func() {
			testBundle.Title = ""
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"title"}))
			})
		})

		Convey("When Validate is called and ManagedBy is empty", func() {
			testBundle.ManagedBy = ""
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"managed_by"}))
			})
		})

		Convey("When Validate is called and all mandatory fields are empty", func() {
			testBundle.BundleType = ""
			testBundle.PreviewTeams = &[]PreviewTeam{}
			testBundle.Title = ""
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"bundle_type", "preview_teams", "title"}))
			})
		})

		Convey("When Validate is called and User fields are empty", func() {
			testBundle.CreatedBy = &User{Email: ""}
			testBundle.LastUpdatedBy = &User{Email: ""}
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"created_by", "last_updated_by"}))
			})
		})

		Convey("When Validate is called and id is empty", func() {
			testBundle.ID = ""
			Convey("Then it should return an error", func() {
				err := ValidateBundle(&testBundle)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"id"}))
			})
		})
	})
}

func TestBundleIdOmitEmpty(t *testing.T) {
	Convey("Given a valid bundle with empty the non-mandatory fields", t, func() {
		previewTeam := PreviewTeam{
			ID: "team1",
		}
		testBundle := Bundle{
			BundleType: BundleTypeManual,
			PreviewTeams: &[]PreviewTeam{
				previewTeam,
			},
			Title:     "Test Bundle",
			ManagedBy: ManagedByWagtail,
		}

		Convey("When marshaling to JSON", func() {
			data, err := json.Marshal(testBundle)

			Convey("Then it should omit the ID field", func() {
				So(err, ShouldBeNil)
				So(string(data), ShouldNotContainSubstring, `"created_by"`)
				So(string(data), ShouldNotContainSubstring, `"created_at"`)
				So(string(data), ShouldNotContainSubstring, `"last_updated_by"`)
				So(string(data), ShouldNotContainSubstring, `"scheduled_at"`)
				So(string(data), ShouldNotContainSubstring, `"state"`)
				So(string(data), ShouldNotContainSubstring, `"updated_at"`)
			})
		})
	})
}
