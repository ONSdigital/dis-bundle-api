package models

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

// Action enum type representing the action
type Action string

// Define the possible values for the Action enum
const (
	ActionCreate Action = "CREATE"
	ActionRead   Action = "READ"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"
)

// IsValid validates that the Action is a valid enum value
func (a Action) IsValid() bool {
	switch a {
	case ActionCreate, ActionRead, ActionUpdate, ActionDelete:
		return true
	default:
		return false
	}
}

// String returns the string value of the Action
func (a Action) String() string {
	return string(a)
}

// MarshalJSON marshals the Action to JSON
func (a Action) MarshalJSON() ([]byte, error) {
	if !a.IsValid() {
		return nil, fmt.Errorf("invalid Action: %s", a)
	}
	return json.Marshal(string(a))
}

// UnmarshalJSON unmarshals a string to Action
func (a *Action) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := Action(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid Action: %s", str)
	}
	*a = converted
	return nil
}

// RequestedBy represents the user who made the request
type RequestedBy struct {
	ID    string `bson:"id" json:"id"`
	Email string `bson:"email,omitempty" json:"email,omitempty"`
}

// Data represents the state of a resource following a change action
type Data struct {
	DatasetID string `bson:"dataset_id" json:"dataset_id"`
	EditionID string `bson:"edition_id" json:"edition_id"`
	ItemID    string `bson:"item_id" json:"item_id"`
	State     string `bson:"state" json:"state"`
	URLPath   string `bson:"url_path" json:"url_path"`
	Title     string `bson:"title,omitempty" json:"title,omitempty"`
}

// Event details a specific event
type Event struct {
	CreatedAt   time.Time   `bson:"created_at" json:"created_at"`
	RequestedBy RequestedBy `bson:"requested_by" json:"requested_by"`
	Action      Action      `bson:"action" json:"action"`
	Resource    string      `bson:"resource" json:"resource"`
	Data        *Data       `bson:"data,omitempty" json:"data,omitempty"`
}

// EventsList represents a list of events
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
	var missingFields []string

	if event.RequestedBy.ID == "" {
		missingFields = append(missingFields, "requested_by.id")
	}

	if event.Action == "" {
		missingFields = append(missingFields, "action")
	}

	if event.Resource == "" {
		missingFields = append(missingFields, "resource")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	return nil
}
