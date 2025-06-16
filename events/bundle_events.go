package events

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/auth"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type EventsManager struct {
	datastore      store.Datastore
	authMiddleware auth.AuthorisationMiddleware
}

// BundleEventsManager defines the interface for managing bundle and content item events.
type BundleEventsManager interface {
	// Bundle events

	// InsertBundleUpdatedEvent creates and stores an event for when a bundle is updated
	InsertBundleUpdatedEvent(ctx context.Context, r *http.Request, bundle *models.Bundle) *models.Error

	// ContentItem events

	// InsertContentItemAddedEvent creates and stores an event for when a content item is added.
	InsertContentItemAddedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error

	// InsertContentItemUpdatedEvent creates and stores an event for when a content item is updated
	InsertContentItemUpdatedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error
}

func InsertEventsManager(store store.Datastore, authMiddleware auth.AuthorisationMiddleware) BundleEventsManager {
	return &EventsManager{
		datastore:      store,
		authMiddleware: authMiddleware,
	}
}

var _ BundleEventsManager = (*EventsManager)(nil)

// InsertBundleUpdatedEvent creates and stores an event
func (em *EventsManager) InsertEvent(ctx context.Context, r *http.Request, action models.Action, bundle *models.Bundle, contentItem *models.ContentItem) *models.Error {
	event, err := em.createEventObject(r, action, createBundleResourceLocation(bundle), contentItem, bundle)

	if err != nil {
		return err
	}

	e := em.datastore.CreateBundleEvent(ctx, event)

	if e != nil {
		return models.CreateModelError(models.CodeInternalServerError, e.Error())
	}

	return nil
}

// InsertBundleUpdatedEvent creates and stores an event for when a bundle is updated.
func (em *EventsManager) InsertBundleUpdatedEvent(ctx context.Context, r *http.Request, bundle *models.Bundle) *models.Error {
	return em.InsertEvent(ctx, r, models.ActionUpdate, bundle, nil)
}

// InsertContentItemAddedEvent creates and stores an event for when a content item is added.
func (em *EventsManager) InsertContentItemAddedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error {
	return em.InsertEvent(ctx, r, models.ActionCreate, nil, contentItem)
}

// InsertContentItemUpdatedEvent creates and stores an event for when a content item is updated
func (em *EventsManager) InsertContentItemUpdatedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error {
	return em.InsertEvent(ctx, r, models.ActionUpdate, nil, contentItem)
}
