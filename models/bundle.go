package models

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
)

// Bundle represents the response body when retrieving a bundle
type Bundle struct {
	ID            string         `bson:"id"                        json:"id"`
	BundleType    BundleType     `bson:"bundle_type"               json:"bundle_type"`
	CreatedBy     *User          `bson:"created_by,omitempty"      json:"created_by,omitempty"`
	CreatedAt     *time.Time     `bson:"created_at,omitempty"      json:"created_at,omitempty"`
	LastUpdatedBy *User          `bson:"last_updated_by,omitempty" json:"last_updated_by,omitempty"`
	PreviewTeams  *[]PreviewTeam `bson:"preview_teams,omitempty"   json:"preview_teams,omitempty"`
	ScheduledAt   *time.Time     `bson:"scheduled_at,omitempty"    json:"scheduled_at,omitempty"`
	State         BundleState    `bson:"state"                     json:"state"`
	Title         string         `bson:"title"                     json:"title"`
	UpdatedAt     *time.Time     `bson:"updated_at,omitempty"      json:"updated_at,omitempty"`
	ManagedBy     ManagedBy      `bson:"managed_by"                json:"managed_by"`
	ETag          string         `bson:"e_tag"                     json:"-"`
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

// CreateBundle creates a new Bundle from the provided reader
func CreateBundle(reader io.Reader, email string) (*Bundle, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var bundle Bundle

	err = json.Unmarshal(b, &bundle)
	if err != nil {
		if strings.Contains(err.Error(), "parsing time") {
			return nil, errs.ErrUnableToParseTime
		}
		return nil, errs.ErrUnableToParseJSON
	}

	bundle.CreatedBy = &User{
		Email: email,
	}

	bundle.LastUpdatedBy = &User{
		Email: email,
	}

	CleanBundle(&bundle)

	etag := dpresponse.GenerateETag(b, false)
	etag = strings.Trim(etag, "\"")
	bundle.ETag = etag

	id, err := newUUID()
	if err != nil {
		return nil, err
	}
	bundle.ID = id.String()

	return &bundle, nil
}

// CleanBundle removes any whitespace from the Bundle fields (except for the CreatedAt, UpdatedAt, ScheduledAt and etag fields)
func CleanBundle(bundle *Bundle) {
	bundle.ID = strings.TrimSpace(bundle.ID)

	bundle.BundleType = BundleType(strings.TrimSpace(bundle.BundleType.String()))

	if bundle.CreatedBy != nil {
		bundle.CreatedBy.Email = strings.TrimSpace(bundle.CreatedBy.Email)
	}

	if bundle.LastUpdatedBy != nil {
		bundle.LastUpdatedBy.Email = strings.TrimSpace(bundle.LastUpdatedBy.Email)
	}

	if bundle.PreviewTeams != nil {
		for _, preview := range *bundle.PreviewTeams {
			preview.ID = strings.TrimSpace(preview.ID)
		}
	}

	bundle.State = BundleState(strings.TrimSpace(bundle.State.String()))

	bundle.Title = strings.TrimSpace(bundle.Title)

	bundle.ManagedBy = ManagedBy(strings.TrimSpace(bundle.ManagedBy.String()))
}

// ValidateBundle checks that the Bundle has all mandatory fields and valid values
func ValidateBundle(bundle *Bundle) []*Error {
	var invalidOrMissingFields []*Error

	codeMissingParameters := CodeMissingParameters
	codeInvalidParameters := CodeInvalidParameters

	if bundle.ID == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeMissingParameters,
			Description: errs.ErrorDescriptionMissingParameters,
			Source: &Source{
				Field: "/id",
			},
		})
	}

	if bundle.BundleType == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeMissingParameters,
			Description: errs.ErrorDescriptionMissingParameters,
			Source: &Source{
				Field: "/bundle_type",
			},
		})
	}

	if bundle.BundleType != "" && !bundle.BundleType.IsValid() {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeInvalidParameters,
			Description: errs.ErrorDescriptionMalformedRequest,
			Source: &Source{
				Field: "/bundle_type",
			},
		})
	}

	if bundle.CreatedBy != nil && bundle.CreatedBy.Email == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeMissingParameters,
			Description: errs.ErrorDescriptionMissingParameters,
			Source: &Source{
				Field: "/created_by/email",
			},
		})
	}

	if bundle.LastUpdatedBy != nil && bundle.LastUpdatedBy.Email == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeInvalidParameters,
			Description: errs.ErrorDescriptionMalformedRequest,
			Source: &Source{
				Field: "/last_updated_by/email",
			},
		})
	}

	if bundle.PreviewTeams != nil {
		for _, team := range *bundle.PreviewTeams {
			if team.ID == "" {
				invalidOrMissingFields = append(invalidOrMissingFields, &Error{
					Code:        &codeMissingParameters,
					Description: errs.ErrorDescriptionMissingParameters,
					Source: &Source{
						Field: "/preview_teams/id",
					},
				})
				break
			}
		}
	}

	if bundle.State.String() == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeMissingParameters,
			Description: errs.ErrorDescriptionMissingParameters,
			Source: &Source{
				Field: "/state",
			},
		})
	}

	if bundle.State.String() != "" && !bundle.State.IsValid() {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeInvalidParameters,
			Description: errs.ErrorDescriptionMalformedRequest,
			Source: &Source{
				Field: "/state",
			},
		})
	}

	if bundle.Title == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeMissingParameters,
			Description: errs.ErrorDescriptionMissingParameters,
			Source: &Source{
				Field: "/title",
			},
		})
	}

	if bundle.ManagedBy == "" {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeMissingParameters,
			Description: errs.ErrorDescriptionMissingParameters,
			Source: &Source{
				Field: "/managed_by",
			},
		})
	}

	if bundle.ManagedBy != "" && !bundle.ManagedBy.IsValid() {
		invalidOrMissingFields = append(invalidOrMissingFields, &Error{
			Code:        &codeInvalidParameters,
			Description: errs.ErrorDescriptionMalformedRequest,
			Source: &Source{
				Field: "/managed_by",
			},
		})
	}

	if len(invalidOrMissingFields) > 0 {
		return invalidOrMissingFields
	}

	return nil
}

func (b *Bundle) GenerateETag(bytes *[]byte) string {
	etag := dpresponse.GenerateETag(*bytes, false)
	etag = strings.Trim(etag, "\"")
	b.ETag = etag

	return b.ETag
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
	ManagedByDataAdmin ManagedBy = "DATA-ADMIN"
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
