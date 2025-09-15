package models

import (
	"errors"
	"time"
)

// Event represents details of a specific change event forming part of the change and audit log for a bundle
type Event struct {
	CreatedAt   *time.Time   `bson:"created_at,omitempty"   json:"created_at,omitempty"`
	RequestedBy *RequestedBy `bson:"requested_by,omitempty" json:"requested_by,omitempty"`
	Action      Action       `bson:"action"                 json:"action"`
	Resource    string       `bson:"resource"               json:"resource"`
	ContentItem *ContentItem `bson:"content_item,omitempty" json:"content_item,omitempty"`
	Bundle      *Bundle      `bson:"bundle,omitempty"       json:"bundle,omitempty"`
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

// Action enum type representing the action taken by a user
type Action string

// Define the possible values for the Action enum
const (
	ActionCreate Action = "CREATE"
	ActionRead   Action = "READ"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"
)

// CreateEventModel creates an Event model for either a Bundle or a ContentItem
func CreateEventModel(id, email string, action Action, bundle *Bundle, contentItem *ContentItem) (*Event, error) {
	if (bundle == nil && contentItem == nil) || (bundle != nil && contentItem != nil) {
		return nil, errors.New("only one of bundle or contentItem must be provided")
	}

	// CreatedAt will be set within Mongo.CreateEvent
	event := &Event{
		RequestedBy: &RequestedBy{
			ID:    id,
			Email: email,
		},
		Action: action,
	}

	if bundle != nil {
		event.Resource = "/bundles/" + bundle.ID
		event.Bundle = bundle
	} else {
		event.Resource = "/bundles/" + contentItem.BundleID + "/contents/" + contentItem.ID
		event.ContentItem = contentItem
	}

	return event, nil
}
