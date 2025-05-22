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

type ErrorUUIDGenerator struct{}

func (e *ErrorUUIDGenerator) NewV4() (uuid.UUID, error) {
	return uuid.UUID{}, fmt.Errorf("mock UUID generation error")
}

func TestCreateContentItem(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the content item has all fields", func() {
			state := StateApproved
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
				State: &state,
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
			b := `{
			"id": "123",
			"bundle_id": "456",
			"content_type": "DATASET",
			"metadata": {
				"dataset_id": "dataset-id",
				"edition_id": "edition-id",
				"title": "title",
				"version_id": 1
			},
			"state": "APPROVED",
			"links": {
				"edit": "/edit",
				"preview": "/preview"
			}
			}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateContentItem(reader, &ErrorUUIDGenerator{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "mock UUID generation error")
		})
	})
}

func TestMarshalJSONForContentItem(t *testing.T) {
	Convey("Given a ContentItem", t, func() {
		state := StateApproved
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
			State: &state,
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
			state = invalid
			contentItem.State = &state

			Convey("Then it should return an error", func() {
				_, err := json.Marshal(contentItem)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid State")
			})
		})

		Convey("When the ID is empty", func() {
			contentItem.ID = ""

			Convey("Then the ID should not be marshalled", func() {
				b, err := json.Marshal(contentItem)
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)

				bStr := string(b)
				So(bStr, ShouldNotContainSubstring, `"id":""`)
			})
		})

		Convey("When the bundle ID is empty", func() {
			contentItem.BundleID = ""

			Convey("Then the bundle ID should not be marshalled", func() {
				b, err := json.Marshal(contentItem)
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)

				bStr := string(b)
				So(bStr, ShouldNotContainSubstring, `"bundle_id":""`)
			})
		})

		Convey("When the title is empty", func() {
			contentItem.Metadata.Title = ""

			Convey("Then the title should not be marshalled", func() {
				b, err := json.Marshal(contentItem)
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)

				bStr := string(b)
				So(bStr, ShouldNotContainSubstring, `"title":""`)
			})
		})

		Convey("When the state is empty", func() {
			contentItem.State = nil

			Convey("Then the state should not be marshalled", func() {
				b, err := json.Marshal(contentItem)
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)

				bStr := string(b)
				So(bStr, ShouldNotContainSubstring, `"state":""`)
			})
		})
	})
}

func TestUnmarshalJSON(t *testing.T) {
	Convey("Given a valid JSON for ContentItem", t, func() {
		validJSON := []byte(`{
			"id": "123",
			"bundle_id": "456",
			"content_type": "DATASET",
			"metadata": {
				"dataset_id": "dataset-id",
				"edition_id": "edition-id",
				"title": "title",
				"version_id": 1
			},
			"state": "APPROVED",
			"links": {
				"edit": "/edit",
				"preview": "/preview"
			}
		}`)

		Convey("When we unmarshall it", func() {
			var contentItem ContentItem
			err := json.Unmarshal(validJSON, &contentItem)

			Convey("Then it should not return an error", func() {
				state := StateApproved
				So(err, ShouldBeNil)
				So(contentItem.ID, ShouldEqual, "123")
				So(contentItem.BundleID, ShouldEqual, "456")
				So(contentItem.ContentType, ShouldEqual, ContentTypeDataset)
				So(contentItem.Metadata.DatasetID, ShouldEqual, "dataset-id")
				So(contentItem.Metadata.EditionID, ShouldEqual, "edition-id")
				So(contentItem.Metadata.Title, ShouldEqual, "title")
				So(contentItem.Metadata.VersionID, ShouldEqual, 1)
				So(contentItem.State, ShouldEqual, &state)
				So(contentItem.Links.Edit, ShouldEqual, "/edit")
				So(contentItem.Links.Preview, ShouldEqual, "/preview")
			})
		})
	})

	Convey("Given a valid JSON for ContentItem", t, func() {
		Convey("And the ID is empty", func() {
			validJSON := []byte(`{
				"bundle_id": "456",
				"content_type": "DATASET",
				"metadata": {
					"dataset_id": "dataset-id",
					"edition_id": "edition-id",
					"title": "title",
					"version_id": 1
				},
				"state": "APPROVED",
				"links": {
					"edit": "/edit",
					"preview": "/preview"
				}
			}`)

			Convey("When we unmarshall it", func() {
				var contentItem ContentItem
				err := json.Unmarshal(validJSON, &contentItem)

				Convey("Then it should not return an error and the ID should be empty", func() {
					state := StateApproved
					So(err, ShouldBeNil)
					So(contentItem.ID, ShouldEqual, "")
					So(contentItem.BundleID, ShouldEqual, "456")
					So(contentItem.ContentType, ShouldEqual, ContentTypeDataset)
					So(contentItem.Metadata.DatasetID, ShouldEqual, "dataset-id")
					So(contentItem.Metadata.EditionID, ShouldEqual, "edition-id")
					So(contentItem.Metadata.Title, ShouldEqual, "title")
					So(contentItem.Metadata.VersionID, ShouldEqual, 1)
					So(contentItem.State, ShouldEqual, &state)
					So(contentItem.Links.Edit, ShouldEqual, "/edit")
					So(contentItem.Links.Preview, ShouldEqual, "/preview")
				})
			})
		})
	})

	Convey("Given a valid JSON for ContentItem", t, func() {
		Convey("And the bundle ID is empty", func() {
			validJSON := []byte(`{
				"id": "123",
				"content_type": "DATASET",
				"metadata": {
					"dataset_id": "dataset-id",
					"edition_id": "edition-id",
					"title": "title",
					"version_id": 1
				},
				"state": "APPROVED",
				"links": {
					"edit": "/edit",
					"preview": "/preview"
				}
			}`)

			Convey("When we unmarshall it", func() {
				var contentItem ContentItem
				err := json.Unmarshal(validJSON, &contentItem)

				Convey("Then it should not return an error and the bundle ID should be empty", func() {
					state := StateApproved
					So(err, ShouldBeNil)
					So(contentItem.ID, ShouldEqual, "123")
					So(contentItem.BundleID, ShouldEqual, "")
					So(contentItem.ContentType, ShouldEqual, ContentTypeDataset)
					So(contentItem.Metadata.DatasetID, ShouldEqual, "dataset-id")
					So(contentItem.Metadata.EditionID, ShouldEqual, "edition-id")
					So(contentItem.Metadata.Title, ShouldEqual, "title")
					So(contentItem.Metadata.VersionID, ShouldEqual, 1)
					So(contentItem.State, ShouldEqual, &state)
					So(contentItem.Links.Edit, ShouldEqual, "/edit")
					So(contentItem.Links.Preview, ShouldEqual, "/preview")
				})
			})
		})
	})

	Convey("Given a valid JSON for ContentItem", t, func() {
		Convey("And the state is empty", func() {
			validJSON := []byte(`{
				"id": "123",
				"bundle_id": "456",
				"content_type": "DATASET",
				"metadata": {
					"dataset_id": "dataset-id",
					"edition_id": "edition-id",
					"title": "title",
					"version_id": 1
				},
				"links": {
					"edit": "/edit",
					"preview": "/preview"
				}
			}`)

			Convey("When we unmarshall it", func() {
				var contentItem ContentItem
				err := json.Unmarshal(validJSON, &contentItem)

				Convey("Then it should not return an error and the state should be empty", func() {
					So(err, ShouldBeNil)
					So(contentItem.ID, ShouldEqual, "123")
					So(contentItem.BundleID, ShouldEqual, "456")
					So(contentItem.ContentType, ShouldEqual, ContentTypeDataset)
					So(contentItem.Metadata.DatasetID, ShouldEqual, "dataset-id")
					So(contentItem.Metadata.EditionID, ShouldEqual, "edition-id")
					So(contentItem.Metadata.Title, ShouldEqual, "title")
					So(contentItem.Metadata.VersionID, ShouldEqual, 1)
					So(contentItem.State, ShouldBeNil)
					So(contentItem.Links.Edit, ShouldEqual, "/edit")
					So(contentItem.Links.Preview, ShouldEqual, "/preview")
				})
			})
		})
	})

	Convey("Given a valid JSON for ContentItem", t, func() {
		Convey("And the content type is empty", func() {
			validJSON := []byte(`{
				"id": "123",
				"bundle_id": "456",
				"metadata": {
					"dataset_id": "dataset-id",
					"edition_id": "edition-id",
					"title": "title",
					"version_id": 1
				},
				"state": "APPROVED",
				"links": {
					"edit": "/edit",
					"preview": "/preview"
				}
			}`)

			Convey("When we unmarshall it", func() {
				var contentItem ContentItem
				err := json.Unmarshal(validJSON, &contentItem)

				Convey("Then we should get an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid content type: ")
				})
			})
		})
	})

	Convey("Given an  invalid JSON for ContentItem", t, func() {
		invalidJSON := []byte(`{
			"id": "123,
			"bundle_id": "456",
			"content_type": "DATASET",
			"metadata": {
				"dataset_id": "dataset-id",
				"edition_id": "edition-id",
				"title": "title",
				"version_id": 1
			},
			"state": "APPROVED",
			"links": {
				"edit": "/edit",
				"preview": "/preview"
			}
		}`)

		Convey("When we unmarshall it", func() {
			var contentItem ContentItem
			err := json.Unmarshal(invalidJSON, &contentItem)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid character")
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

func TestUnmarshalJSON_InvalidContentItem(t *testing.T) {
	Convey("Given invalid JSON for ContentItem", t, func() {
		invalidJSON := []byte(`{
			"id": "123,
			"bundle_id": "456",
			"content_type": "DATASET",
			"metadata": {
				"dataset_id": "dataset-id",
				"edition_id": "edition-id",
				"title": "title",
				"version_id": 1
			},
			"state": "APPROVED",
			"links": {
				"edit": "/edit",
				"preview": "/preview"
			}
		}`)

		Convey("When UnmarshalJSON is called", func() {
			var contentItem ContentItem
			err := contentItem.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid character")
			})
		})
	})
}

func TestValidateContentItem(t *testing.T) {
	Convey("Given a ContentItem with all required fields", t, func() {
		state := StateApproved
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
			State: &state,
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

		Convey("When Validate is called and Links has blank fields", func() {
			contentItem.Links = Links{Edit: "", Preview: ""}
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
