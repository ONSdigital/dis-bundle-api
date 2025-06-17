package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

// Event represents details of a specific change event forming part of the change and audit log for a bundle
type Event struct {
	CreatedAt   *time.Time   `bson:"created_at,omitempty"   json:"created_at,omitempty"`
	RequestedBy *RequestedBy `bson:"requested_by,omitempty" json:"requested_by,omitempty"`
	Action      Action       `bson:"action"                 json:"action"`
	Resource    string       `bson:"resource"               json:"resource"`
	ContentItem *ContentItem `bson:"content_item,omitempty" json:"content_item,omitempty"`
	Bundle      *EventBundle `bson:"bundle,omitempty"       json:"bundle,omitempty"`
}

// EventBundle represents the Bundle response body when retrieving an Event
type EventBundle struct {
	ID            string         `bson:"id"                       json:"id"`
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

// RequestedBy represents the user who made the request
type RequestedBy struct {
	ID    string `bson:"id"              json:"id"`
	Email string `bson:"email,omitempty" json:"email,omitempty"`
}

// EventsList represents the list of change events which form the change and audit log for a bundle
type EventsList struct {
	PaginationFields
	Items *[]Event `bson:"items,omitempty" json:"items,omitempty"`
}

// CreateEvent creates an Event from a JSON request body
func CreateEvent(reader io.Reader) (*Event, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var event Event
	err = json.Unmarshal(b, &event)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	return &event, nil
}

// ValidateEvent validates that an Event has all required fields and values
func ValidateEvent(event *Event) error {
	missingFields, invalidFields := []string{}, []string{}

	if event.RequestedBy != nil && event.RequestedBy.ID == "" {
		missingFields = append(missingFields, "requested_by.id")
	}

	if event.Action == "" {
		missingFields = append(missingFields, "action")
	}

	if event.Action != "" && !event.Action.IsValid() {
		invalidFields = append(invalidFields, "action")
	}

	if event.Resource == "" {
		missingFields = append(missingFields, "resource")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if len(invalidFields) > 0 {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil
}

// Action enum type representing the action taken by a user
type Action string

// Define the possible values for the Action enum
const (
	ActionCreate Action = "CREATE"
	ActionRead   Action = "READ"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"
)

// String returns the string value of the Action
func (a Action) String() string {
	return string(a)
}

// IsValid validates that the Action is a valid enum value
func (a Action) IsValid() bool {
	switch a {
	case ActionCreate, ActionRead, ActionUpdate, ActionDelete:
		return true
	default:
		return false
	}
}

func ConvertBundleToBundleEvent(bundle *Bundle) (*EventBundle, error) {
	if bundle == nil {
		return nil, errors.New("input bundle cannot be nil")
	}
	return &EventBundle{
		ID:            bundle.ID,
		BundleType:    bundle.BundleType,
		CreatedBy:     bundle.CreatedBy,
		CreatedAt:     bundle.CreatedAt,
		LastUpdatedBy: bundle.LastUpdatedBy,
		PreviewTeams:  bundle.PreviewTeams,
		ScheduledAt:   bundle.ScheduledAt,
		State:         bundle.State,
		Title:         bundle.Title,
		UpdatedAt:     bundle.UpdatedAt,
		ManagedBy:     bundle.ManagedBy,
	}, nil
}
