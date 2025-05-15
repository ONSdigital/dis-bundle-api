package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock read error")
}

type ErrorUUIDGenerator struct{}

func (e *ErrorUUIDGenerator) NewV4() (uuid.UUID, error) {
	return uuid.UUID{}, fmt.Errorf("mock UUID generation error")
}

const invalid = "INVALID"

func TestCreateContentItem(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the content item has all fields", func() {
			testContentItem := ContentItem{
				ID:          "123",
				BundleID:    "456",
				ContentType: ContentTypeDataset,
				Metadata: Metadata{
					DatasetID: "dataset-id",
					EditionID: "edition-id",
					Title:     "title",
					VersionID: 1,
				},
				State: StateApproved,
				Links: Links{
					Edit:    "/edit",
					Preview: "/preview",
				},
			}

			b, err := json.Marshal(testContentItem)
			if err != nil {
				t.Logf("failed to marshal test data into bytes, error: %v", err)
				t.FailNow()
			}

			reader := bytes.NewReader(b)
			result, err := CreateContentItem(reader, DefaultUUIDGenerator{})
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ID, ShouldNotBeEmpty)
			_, err = uuid.FromString(result.ID)
			So(err, ShouldBeNil)
			So(result.BundleID, ShouldEqual, testContentItem.BundleID)
			So(result.ContentType, ShouldEqual, testContentItem.ContentType)
			So(result.Metadata.DatasetID, ShouldEqual, testContentItem.Metadata.DatasetID)
			So(result.Metadata.EditionID, ShouldEqual, testContentItem.Metadata.EditionID)
			So(result.Metadata.Title, ShouldEqual, testContentItem.Metadata.Title)
			So(result.Metadata.VersionID, ShouldEqual, testContentItem.Metadata.VersionID)
			So(result.State, ShouldEqual, testContentItem.State)
			So(result.Links.Edit, ShouldEqual, testContentItem.Links.Edit)
			So(result.Links.Preview, ShouldEqual, testContentItem.Links.Preview)
		})
	})

	Convey("Return error when unable to read message", t, func() {
		Convey("when the reader returns an error", func() {
			reader := &ErrorReader{}
			_, err := CreateContentItem(reader, DefaultUUIDGenerator{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToReadMessage.Error())
		})
	})

	Convey("Return error when unable to parse JSON", t, func() {
		Convey("when the JSON is invalid", func() {
			b := `{"bundle_id": "123}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateContentItem(reader, DefaultUUIDGenerator{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
		})
	})

	Convey("Return error when unable to generate UUID", t, func() {
		Convey("when the UUID generator returns an error", func() {
			b := `{"bundle_id": "123"}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateContentItem(reader, &ErrorUUIDGenerator{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "mock UUID generation error")
		})
	})
}

func TestMarshalJSON(t *testing.T) {
	Convey("Given a ContentItem", t, func() {
		contentItem := ContentItem{
			ID:          "123",
			BundleID:    "456",
			ContentType: ContentTypeDataset,
			Metadata: Metadata{
				DatasetID: "dataset-id",
				EditionID: "edition-id",
				Title:     "title",
				VersionID: 1,
			},
			State: StateApproved,
			Links: Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}

		Convey("When the content type is invalid", func() {
			contentItem.ContentType = invalid

			Convey("Then it should return an error", func() {
				_, err := json.Marshal(contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid ContentType")
			})
		})

		Convey("When the state is invalid", func() {
			contentItem.State = invalid

			Convey("Then it should return an error", func() {
				_, err := json.Marshal(contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid State")
			})
		})
	})
}

func TestUnmarshalJSON_InvalidContentType(t *testing.T) {
	Convey("Given invalid JSON for ContentType", t, func() {
		invalidJSON := []byte(`123`) // Invalid JSON for a string
		var contentType ContentType

		Convey("When UnmarshalJSON is called", func() {
			err := contentType.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given invalid JSON for ContentType", t, func() {
		invalidJSON := []byte(`"INVALID"`) // Invalid value for ContentType
		var contentType ContentType

		Convey("When UnmarshalJSON is called", func() {
			err := contentType.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid ContentType: INVALID")
			})
		})
	})
}

func TestUnmarshalJSON_InvalidState(t *testing.T) {
	Convey("Given invalid JSON for State", t, func() {
		invalidJSON := []byte(`123`) // Invalid JSON for a string
		var state State

		Convey("When UnmarshalJSON is called", func() {
			err := state.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given invalid JSON for State", t, func() {
		invalidJSON := []byte(`"INVALID"`) // Invalid value for State
		var state State

		Convey("When UnmarshalJSON is called", func() {
			err := state.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid State: INVALID")
			})
		})
	})
}

func TestValidateContentItem(t *testing.T) {
	Convey("Given a ContentItem with all required fields", t, func() {
		contentItem := ContentItem{
			ID:          "123",
			BundleID:    "456",
			ContentType: ContentTypeDataset,
			Metadata: Metadata{
				DatasetID: "dataset-id",
				EditionID: "edition-id",
				Title:     "title",
				VersionID: 1,
			},
			State: StateApproved,
			Links: Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}

		Convey("When Validate is called with a valid ContentItem", func() {
			err := ValidateContentItem(&contentItem)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When Validate is called and ContentType is empty", func() {
			contentItem.ContentType = ""
			Convey("Then it should return an error", func() {
				err := ValidateContentItem(&contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"content_type"}))
			})
		})

		Convey("When Validate is called and Metadata is empty", func() {
			contentItem.Metadata = Metadata{}
			Convey("Then it should return an error", func() {
				err := ValidateContentItem(&contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"dataset_id", "edition_id"}))
			})
		})

		Convey("When Validate is called and Links is empty", func() {
			contentItem.Links = Links{}
			Convey("Then it should return an error", func() {
				err := ValidateContentItem(&contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"edit", "preview"}))
			})
		})

		Convey("When Validate is called and VersionID is invalid", func() {
			contentItem.Metadata.VersionID = -1
			Convey("Then it should return an error", func() {
				err := ValidateContentItem(&contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("invalid fields: %v", []string{"version_id"}))
			})
		})
	})
}
