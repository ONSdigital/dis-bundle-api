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

type BundleEventsManager interface {
	//Bundle events
	InsertBundleUpdatedEvent(ctx context.Context, r *http.Request, bundle *models.Bundle) *models.Error

	//ContentItem events
	InsertContentItemAddedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error
	InsertContentItemUpdatedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error
}

func CreateBundleEventsManager(store store.Datastore, authMiddleware auth.AuthorisationMiddleware) BundleEventsManager {
	return &EventsManager{
		datastore:      store,
		authMiddleware: authMiddleware,
	}
}

var _ BundleEventsManager = (*EventsManager)(nil)

func (em *EventsManager) CreateBundleEvent(ctx context.Context, r *http.Request, action models.Action, bundle *models.Bundle, contentItem *models.ContentItem) *models.Error {
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

func (em *EventsManager) InsertBundleUpdatedEvent(ctx context.Context, r *http.Request, bundle *models.Bundle) *models.Error {
	return em.CreateBundleEvent(ctx, r, models.ActionUpdate, bundle, nil)
}

func (em *EventsManager) InsertContentItemAddedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error {
	return em.CreateBundleEvent(ctx, r, models.ActionCreate, nil, contentItem)
}

func (em *EventsManager) InsertContentItemUpdatedEvent(ctx context.Context, r *http.Request, contentItem *models.ContentItem) *models.Error {
	return em.CreateBundleEvent(ctx, r, models.ActionUpdate, nil, contentItem)
}
