package models

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

// Bundle represents the response body when retrieving a bundle
type Bundle struct {
	ID            string         `bson:"_id"                       json:"id"`
	BundleType    BundleType     `bson:"bundle_type"               json:"bundle_type"`
	CreatedBy     *User          `bson:"created_by,omitempty"      json:"created_by,omitempty"`
	CreatedAt     *time.Time     `bson:"created_at,omitempty"      json:"created_at,omitempty"`
	LastUpdatedBy *User          `bson:"last_updated_by,omitempty" json:"last_updated_by,omitempty"`
	PreviewTeams  *[]PreviewTeam `bson:"preview_teams"             json:"preview_teams"`
	ScheduledAt   *time.Time     `bson:"scheduled_at,omitempty"    json:"scheduled_at,omitempty"`
	State         *BundleState   `bson:"state,omitempty"           json:"state,omitempty"`
	Title         string         `bson:"title"                     json:"title"`
	UpdatedAt     *time.Time     `bson:"updated_at,omitempty"      json:"updated_at,omitempty"`
	ManagedBy     ManagedBy      `bson:"managed_by"                json:"managed_by"`
}

// Bundles represents a list of bundles
type Bundles struct {
	Items *[]Bundle `bson:"items" json:"items"`
}

// User represents the user who created or updated the bundle
type User struct {
	Email string `bson:"email" json:"email"`
}

// PreviewTeam represents a team who have permissions to view the dataset series in the bundle
type PreviewTeam struct {
	ID string `bson:"id" json:"id"`
}

// BundleContent represents the content of the bundle
type BundleContent struct {
	DatasetID string `bson:"dataset_id" json:"dataset_id"`
	EditionID string `bson:"edition_id" json:"edition_id"`
	ItemID    string `bson:"item_id" json:"item_id"`
	State     string `bson:"state" json:"state"`
	Title     string `bson:"title" json:"title"`
	URLPath   string `bson:"url_path" json:"url_path"`
}

func CreateBundle(reader io.Reader) (*Bundle, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var bundle Bundle

	err = json.Unmarshal(b, &bundle)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	return &bundle, nil
}

func ValidateBundle(bundle *Bundle) error {
	missingFields, invalidFields := []string{}, []string{}

	if bundle.ID == "" {
		missingFields = append(missingFields, "id")
	}

	if bundle.BundleType == "" {
		missingFields = append(missingFields, "bundle_type")
	}

	if bundle.BundleType != "" && !bundle.BundleType.IsValid() {
		invalidFields = append(invalidFields, "bundle_type")
	}

	if bundle.CreatedBy != nil && bundle.CreatedBy.Email == "" {
		missingFields = append(missingFields, "created_by.email")
	}

	if bundle.LastUpdatedBy != nil && bundle.LastUpdatedBy.Email == "" {
		missingFields = append(missingFields, "last_updated_by.email")
	}

	if len(*bundle.PreviewTeams) == 0 {
		missingFields = append(missingFields, "preview_teams")
	}

	if bundle.State != nil && !bundle.State.IsValid() {
		invalidFields = append(invalidFields, "state")
	}

	if bundle.Title == "" {
		missingFields = append(missingFields, "title")
	}

	if bundle.ManagedBy == "" {
		missingFields = append(missingFields, "managed_by")
	}

	if bundle.ManagedBy != "" && !bundle.ManagedBy.IsValid() {
		invalidFields = append(invalidFields, "managed_by")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if len(invalidFields) > 0 {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil
}

// BundleType enum type representing the type of the bundle
type BundleType string

// Define the possible values for the BundleType enum
const (
	BundleTypeManual    BundleType = "MANUAL"
	BundleTypeScheduled BundleType = "SCHEDULED"
)

// IsValid validates that the BundleType is a valid enum value
func (bt BundleType) IsValid() bool {
	switch bt {
	case BundleTypeManual, BundleTypeScheduled:
		return true
	default:
		return false
	}
}

// BundleState enum type representing the state of the bundle
type BundleState string

// Define the possible values for the BundleState enum
const (
	BundleStateDraft     BundleState = "DRAFT"
	BundleStateInReview  BundleState = "IN_REVIEW"
	BundleStateApproved  BundleState = "APPROVED"
	BundleStatePublished BundleState = "PUBLISHED"
)

// IsValid validates that the BundleState is a valid enum value
func (bs BundleState) IsValid() bool {
	switch bs {
	case BundleStateDraft, BundleStateInReview, BundleStateApproved, BundleStatePublished:
		return true
	default:
		return false
	}
}

// String returns the string value of the BundleState
func (bs BundleState) String() string {
	return string(bs)
}

// ManagedBy enum type representing the system that created and manages the bundle
type ManagedBy string

// Define the possible values for the ManagedBy enum
const (
	ManagedByWagtail   ManagedBy = "WAGTAIL"
	ManagedByDataAdmin ManagedBy = "DATA_ADMIN"
)

// IsValid validates that the ManagedBy is a valid enum value
func (mb ManagedBy) IsValid() bool {
	switch mb {
	case ManagedByWagtail, ManagedByDataAdmin:
		return true
	default:
		return false
	}
}
