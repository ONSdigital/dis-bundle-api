package models

import (
	"encoding/json"
	"fmt"
	"io"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/gofrs/uuid"
)

// ContentItem represents information about the datasets to be published as part of the bundle
type ContentItem struct {
	ID          string      `bson:"_id,omitempty" json:"id,omitempty"`
	BundleID    string      `bson:"bundle_id" json:"bundle_id"`
	ContentType ContentType `bson:"content_type" json:"content_type"`
	Metadata    Metadata    `bson:"metadata" json:"metadata"`
	State       *State      `bson:"state,omitempty" json:"state,omitempty"`
	Links       Links       `bson:"links" json:"links"`
}

// Metadata represents the metadata for the content item
type Metadata struct {
	DatasetID string `bson:"dataset_id" json:"dataset_id"`
	EditionID string `bson:"edition_id" json:"edition_id"`
	Title     string `bson:"title,omitempty" json:"title,omitempty"`
	VersionID int    `bson:"version_id" json:"version_id"`
}

// Links represents the navigational links for onward actions related to the content item
type Links struct {
	Edit    string `bson:"edit" json:"edit"`
	Preview string `bson:"preview" json:"preview"`
}

// Contents represents a list of contents related to a bundle
type Contents struct {
	PaginationFields
	Items []ContentItem `bson:"contents,omitempty" json:"contents,omitempty"`
}

// UnmarshalJSON unmarshals a string to ContentItem
func (c *ContentItem) UnmarshalJSON(data []byte) error {
	type Alias ContentItem
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if !aux.ContentType.IsValid() {
		return fmt.Errorf("invalid content type: %s", aux.ContentType)
	}

	return nil
}

// ContentType enum represents the type of content
type ContentType string

// Define the possible values for the contentType enum
const (
	ContentTypeDataset ContentType = "DATASET"
)

// IsValid validates that the ContentType is a valid enum value
func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypeDataset:
		return true
	default:
		return false
	}
}

// String returns the string value of the ContentType
func (ct ContentType) String() string {
	return string(ct)
}

// MarshalJSON marshals the ContentType to JSON
func (ct ContentType) MarshalJSON() ([]byte, error) {
	if !ct.IsValid() {
		return nil, fmt.Errorf("invalid ContentType: %s", ct)
	}
	return json.Marshal(string(ct))
}

// UnmarshalJSON unmarshals a string to ContentType
func (ct *ContentType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := ContentType(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid ContentType: %s", str)
	}
	*ct = converted
	return nil
}

// State enum represents the state of the content item
type State string

// Define the possible values for the state enum
const (
	StateApproved  State = "APPROVED"
	StatePublished State = "PUBLISHED"
)

// IsValid validates that the State is a valid enum value
func (s State) IsValid() bool {
	switch s {
	case StateApproved, StatePublished:
		return true
	default:
		return false
	}
}

// String returns the string value of the State
func (s State) String() string {
	return string(s)
}

// MarshalJSON marshals the State to JSON
func (s State) MarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid State: %s", s)
	}
	return json.Marshal(string(s))
}

// UnmarshalJSON unmarshals a string to State
func (s *State) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := State(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid State: %s", str)
	}
	*s = converted
	return nil
}

var newUUID = uuid.NewV4

func CreateContentItem(reader io.Reader) (*ContentItem, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var contentItem ContentItem

	err = json.Unmarshal(b, &contentItem)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	id, err := newUUID()
	if err != nil {
		return nil, err
	}
	contentItem.ID = id.String()

	return &contentItem, nil
}

func ValidateContentItem(contentItem *ContentItem) error {
	var missingFields []string
	var invalidFields []string

	if contentItem.BundleID == "" {
		missingFields = append(missingFields, "bundle_id")
	}

	if contentItem.ContentType == "" {
		missingFields = append(missingFields, "content_type")
	}

	if contentItem.Metadata.DatasetID == "" {
		missingFields = append(missingFields, "dataset_id")
	}
	if contentItem.Metadata.EditionID == "" {
		missingFields = append(missingFields, "edition_id")
	}
	if contentItem.Metadata.VersionID < 1 {
		invalidFields = append(invalidFields, "version_id")
	}

	if contentItem.Links.Edit == "" {
		missingFields = append(missingFields, "edit")
	}
	if contentItem.Links.Preview == "" {
		missingFields = append(missingFields, "preview")
	}

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil
}
