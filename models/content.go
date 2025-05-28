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
	ID          string      `bson:"_id,omitempty"   json:"id,omitempty"`
	BundleID    string      `bson:"bundle_id"       json:"bundle_id"`
	ContentType ContentType `bson:"content_type"    json:"content_type"`
	Metadata    Metadata    `bson:"metadata"        json:"metadata"`
	State       *State      `bson:"state,omitempty" json:"state,omitempty"`
	Links       Links       `bson:"links"           json:"links"`
}

// Metadata represents the metadata for the content item
type Metadata struct {
	DatasetID string `bson:"dataset_id"      json:"dataset_id"`
	EditionID string `bson:"edition_id"      json:"edition_id"`
	Title     string `bson:"title,omitempty" json:"title,omitempty"`
	VersionID int    `bson:"version_id"      json:"version_id"`
}

// Links represents the navigational links for onward actions related to the content item
type Links struct {
	Edit    string `bson:"edit"    json:"edit"`
	Preview string `bson:"preview" json:"preview"`
}

// Contents represents a list of contents related to a bundle
type Contents struct {
	PaginationFields
	Items []ContentItem `bson:"contents,omitempty" json:"contents,omitempty"`
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
	missingFields, invalidFields := []string{}, []string{}

	if contentItem.BundleID == "" {
		missingFields = append(missingFields, "bundle_id")
	}

	if contentItem.ContentType == "" {
		missingFields = append(missingFields, "content_type")
	}

	if !contentItem.ContentType.IsValid() {
		invalidFields = append(invalidFields, "content_type")
	}

	if contentItem.Metadata.DatasetID == "" {
		missingFields = append(missingFields, "metadata.dataset_id")
	}
	if contentItem.Metadata.EditionID == "" {
		missingFields = append(missingFields, "metadata.edition_id")
	}
	if contentItem.Metadata.VersionID < 1 {
		invalidFields = append(invalidFields, "metadata.version_id")
	}

	if contentItem.State != nil && !contentItem.State.IsValid() {
		invalidFields = append(invalidFields, "state")
	}

	if contentItem.Links.Edit == "" {
		missingFields = append(missingFields, "links.edit")
	}
	if contentItem.Links.Preview == "" {
		missingFields = append(missingFields, "links.preview")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if len(invalidFields) > 0 {
		return fmt.Errorf("invalid fields: %v", invalidFields)
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
