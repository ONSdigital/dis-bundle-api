package models

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/gofrs/uuid"
)

// ContentItem represents information about the datasets to be published as part of the bundle
type ContentItem struct {
	ID          string      `bson:"id,omitempty"    json:"id,omitempty"`
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

func CleanContentItem(contentItem *ContentItem) {
	contentItem.ID = strings.TrimSpace(contentItem.ID)

	contentItem.BundleID = strings.TrimSpace(contentItem.BundleID)

	contentItem.ContentType = ContentType(strings.TrimSpace(contentItem.ContentType.String()))

	contentItem.Metadata.DatasetID = strings.TrimSpace(contentItem.Metadata.DatasetID)
	contentItem.Metadata.EditionID = strings.TrimSpace(contentItem.Metadata.EditionID)
	contentItem.Metadata.Title = strings.TrimSpace(contentItem.Metadata.Title)

	if contentItem.State != nil {
		state := State(strings.TrimSpace(contentItem.State.String()))
		contentItem.State = &state
	}

	contentItem.Links.Edit = strings.TrimSpace(contentItem.Links.Edit)
	contentItem.Links.Preview = strings.TrimSpace(contentItem.Links.Preview)
}

func ValidateContentItem(contentItem *ContentItem) []*Error {
	var (
		invalidOrMissingFields = []*Error{}
	)

	codeMissingParameters := CodeMissingParameters
	codeInvalidParameters := CodeInvalidParameters

	if contentItem.BundleID == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeMissingParameters, Description: errs.ErrorDescriptionMissingParameters, Source: &Source{Field: "/bundle_id"}})
	}

	if contentItem.ContentType == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeMissingParameters, Description: errs.ErrorDescriptionMissingParameters, Source: &Source{Field: "/content_type"}})
	}

	if contentItem.ContentType != "" && !contentItem.ContentType.IsValid() {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeInvalidParameters, Description: errs.ErrorDescriptionMalformedRequest, Source: &Source{Field: "/content_type"}})
	}

	if contentItem.Metadata.DatasetID == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeMissingParameters, Description: errs.ErrorDescriptionMissingParameters, Source: &Source{Field: "/metadata/dataset_id"}})
	}
	if contentItem.Metadata.EditionID == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeMissingParameters, Description: errs.ErrorDescriptionMissingParameters, Source: &Source{Field: "/metadata/edition_id"}})
	}
	if contentItem.Metadata.VersionID < 1 {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeInvalidParameters, Description: errs.ErrorDescriptionMalformedRequest, Source: &Source{Field: "/metadata/version_id"}})
	}

	if contentItem.State != nil && !contentItem.State.IsValid() {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeInvalidParameters, Description: errs.ErrorDescriptionMalformedRequest, Source: &Source{Field: "/state"}})
	}

	if contentItem.Links.Edit == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeMissingParameters, Description: errs.ErrorDescriptionMissingParameters, Source: &Source{Field: "/links/edit"}})
	}
	if contentItem.Links.Preview == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{Code: &codeMissingParameters, Description: errs.ErrorDescriptionMissingParameters, Source: &Source{Field: "/links/preview"}})
	}

	if len(invalidOrMissingFields) > 0 {
		return invalidOrMissingFields
	}

	return nil
}

func (ci *ContentItem) VersionIdString() string {
	versionId := strconv.Itoa(ci.Metadata.VersionID)
	return versionId
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

// State enum represents the state of the content item
type State string

// Define the possible values for the state enum
const (
	StateApproved  State = "APPROVED"
	StatePublished State = "PUBLISHED"
	StateDraft     State = "DRAFT"
	StateInReview  State = "IN_REVIEW"
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

func GetMatchingStateForBundleState(bundleState BundleState) (*State, error) {
	str := bundleState.String()
	var state State
	switch str {
	case StateApproved.String():
		state = StateApproved
		return &state, nil
	case StatePublished.String():
		state = StatePublished
		return &state, nil
	case StateDraft.String():
		state = StateDraft
		return &state, nil
	case StateInReview.String():
		state = StateInReview
		return &state, nil
	default:
		return nil, errors.New("not found state")
	}
}
