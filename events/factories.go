package events

import (
	"fmt"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
)

// createEventObject creates an Event + EventBundle object, and validates them before returning
func (em *EventsManager) createEventObject(r *http.Request, action models.Action, resource string, contentItem *models.ContentItem, bundle *models.Bundle) (*models.Event, *models.Error) {
	requestedBy, err := em.createReqestedBy(r)

	if err != nil {
		return nil, err
	}

	var eventBundle *models.EventBundle = nil

	if bundle != nil {
		mappedBundle, createEventBundleErr := models.CreateEventBundle(bundle)
		if createEventBundleErr != nil {
			return nil, models.CreateModelError(models.CodeInternalServerError, "failed to create EventBundle from Bundle")
		}
		eventBundle = mappedBundle
	}
	event := &models.Event{
		RequestedBy: requestedBy,
		Action:      action,
		Resource:    resource,
		ContentItem: contentItem,
		Bundle:      eventBundle,
	}

	validationErr := models.ValidateEvent(event)

	if validationErr != nil {
		return nil, models.CreateModelError(models.CodeInternalServerError, apierrors.ErrorDescriptionValidationEventFailed)
	}

	return event, nil
}

// createRequestedBy extracts the EntityData from the middleware and creates a RequestedBy object with the extracted data
func (em *EventsManager) createReqestedBy(r *http.Request) (*models.RequestedBy, *models.Error) {
	JWTEntityData, err := em.authMiddleware.GetJWTEntityData(r)

	if err != nil {
		return nil, err
	}

	return &models.RequestedBy{
		ID:    JWTEntityData.UserID,
		Email: JWTEntityData.UserID,
	}, nil
}

// createBundleResourceLocation creates a resource location for the provided bundle
func createBundleResourceLocation(bundle *models.Bundle) string {
	return fmt.Sprintf("/bundle/%s", bundle.ID)
}

// createBundleContentResourceLocation creates a resource location for the provided bundle
func createBundleContentResourceLocation(content *models.ContentItem) string {
	return fmt.Sprintf("/bundle/%s/content/%s", content.BundleID, content.ID)
}
