package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

// ListBundleEvents retrieves all bundle events
func (m *Mongo) ListBundleEvents(ctx context.Context, offset, limit int) (bundles []*models.Event, totalCount int, err error) {
	var events []*models.Event

	filter, sort := buildListBundleEventsQuery()

	totalCount, err = m.Connection.Collection(m.ActualCollectionName(config.BundleEventsCollection)).
		Find(ctx, filter, &events, mongodriver.Sort(sort), mongodriver.Offset(offset), mongodriver.Limit(limit))

	if err != nil {
		return nil, 0, err
	}

	return events, totalCount, nil
}

func buildListBundleEventsQuery() (filter, sort bson.M) {
	filter = bson.M{}
	sort = bson.M{"_id": 1}
	return
}

// GetBundleEvent retrieves a single bundle event by ID
func (m *Mongo) GetBundleEvent(ctx context.Context, eventID string) (*models.Event, error) {
	filter := buildGetBundleEventQuery(eventID)

	var result models.Event
	err := m.Connection.Collection(m.ActualCollectionName(config.BundleEventsCollection)).
		FindOne(ctx, filter, &result)

	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, apierrors.ErrBundleEventNotFound
		}
	}
	return &result, nil
}

func buildGetBundleEventQuery(eventID string) bson.M {
	return bson.M{"_id": eventID}
}

// CreateBundleEvent inserts a new bundle event
func (m *Mongo) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	now := time.Now()
	event.CreatedAt = &now

	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundleEventsCollection)).
		InsertOne(ctx, event)

	if err != nil {
		return err
	}

	return nil
}

// UpdateBundleEvent updates an existing bundle event
func (m *Mongo) UpdateBundleEvent(ctx context.Context, id string, update *models.Event) (*models.Event, error) {
	collectionName := m.ActualCollectionName(config.BundleEventsCollection)
	filter := buildGetBundleEventQuery(id)

	updateData := bson.M{
		"$set": bson.M{
			"created_at":   update.CreatedAt,
			"requested_by": update.RequestedBy,
			"action":       update.Action,
			"resource":     update.Resource,
			"content_item": update.ContentItem,
			"bundle":       update.Bundle,
		},
	}

	_, err := m.Connection.Collection(collectionName).UpdateOne(ctx, filter, updateData)
	if err != nil {
		return nil, err
	}

	// Re-fetch updated event to return full latest version
	return m.GetBundleEvent(ctx, id)
}

// DeleteBundleEvent deletes a bundle event by ID
func (m *Mongo) DeleteBundleEvent(ctx context.Context, id string) error {
	collectionName := m.ActualCollectionName(config.BundleEventsCollection)
	filter := buildGetBundleEventQuery(id)

	_, err := m.Connection.Collection(collectionName).DeleteOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return apierrors.ErrBundleEventNotFound
		}
		return err
	}

	log.Info(ctx, "bundle event deleted", log.Data{"event_id": id})
	return nil
}
