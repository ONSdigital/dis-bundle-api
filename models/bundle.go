package models

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

// Bundle is the model for the response body when retrieving a bundle
type Bundle struct {
	ID            string         `bson:"_id,omitempty" json:"id,omitempty"`
	BundleType    BundleType     `bson:"bundle_type" json:"bundle_type"` //
	CreatedBy     *User          `bson:"created_by,omitempty" json:"created_by,omitempty"`
	CreatedAt     *time.Time     `bson:"created_at,omitempty" json:"created_at,omitempty"`
	LastUpdatedBy *User          `bson:"last_updated_by,omitempty" json:"last_updated_by,omitempty"`
	PreviewTeams  *[]PreviewTeam `bson:"preview_teams" json:"preview_teams"` //
	ScheduledAt   *time.Time     `bson:"scheduled_at,omitempty" json:"scheduled_at,omitempty"`
	State         *BundleState   `bson:"state,omitempty" json:"state,omitempty"`
	Title         string         `bson:"title" json:"title"` //
	UpdatedAt     *time.Time     `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	ManagedBy     ManagedBy      `bson:"managed_by" json:"managed_by"` //
}

// Bundles represents a list of bundles
type Bundles struct {
	Items []Bundle `bson:"items" json:"items"`
}

type User struct {
	Email string `bson:"email" json:"email"`
}

type PreviewTeam struct {
	ID string `bson:"id" json:"id"`
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

// String returns the string value of the BundleType
func (bt BundleType) String() string {
	return string(bt)
}

// MarshalJSON marshals the BundleType to JSON
func (bt BundleType) MarshalJSON() ([]byte, error) {
	if !bt.IsValid() {
		return nil, fmt.Errorf("invalid BundleType: %s", bt)
	}
	return json.Marshal(string(bt))
}

// UnmarshalJSON unmarshals a string to BundleType
func (bt *BundleType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := BundleType(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid BundleType: %s", str)
	}
	*bt = converted
	return nil
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

// MarshalJSON marshals the BundleState to JSON
func (bs BundleState) MarshalJSON() ([]byte, error) {
	if !bs.IsValid() {
		return nil, fmt.Errorf("invalid BundleState: %s", bs)
	}
	return json.Marshal(string(bs))
}

// UnmarshalJSON unmarshals a string to BundleState
func (bs *BundleState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := BundleState(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid BundleState: %s", str)
	}
	*bs = converted
	return nil
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

// String returns the string value of the ManagedBy
func (mb ManagedBy) String() string {
	return string(mb)
}

// MarshalJSON marshals the ManagedBy to JSON
func (mb ManagedBy) MarshalJSON() ([]byte, error) {
	if !mb.IsValid() {
		return nil, fmt.Errorf("invalid ManagedBy: %s", mb)
	}
	return json.Marshal(string(mb))
}

// UnmarshalJSON unmarshals a string to ManagedBy
func (mb *ManagedBy) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := ManagedBy(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid ManagedBy: %s", str)
	}
	*mb = converted
	return nil
}

func ValidateBundle(bundle *Bundle) error {
	var missingFields []string

	if bundle.BundleType == "" {
		missingFields = append(missingFields, "bundle_type")
	}

	if len(*bundle.PreviewTeams) == 0 {
		missingFields = append(missingFields, "preview_teams")
	}

	if bundle.Title == "" {
		missingFields = append(missingFields, "title")
	}

	if bundle.ManagedBy == "" {
		missingFields = append(missingFields, "managed_by")
	}

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	return nil
}
