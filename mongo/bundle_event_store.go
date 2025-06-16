package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateBundleEvent inserts a new bundle event
func (m *Mongo) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	now := time.Now()
	event.CreatedAt = &now

	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundleEventsCollection)).
		InsertOne(ctx, event)

	return err
}

// ListBundleEvents retrieves all bundle events with optional filtering and pagination
func (m *Mongo) ListBundleEvents(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) (events []*models.Event, totalCount int, err error) {
	var results []*models.Event

	filter, sort := buildListBundleEventsQuery(bundleID, after, before)

	totalCount, err = m.Connection.Collection(m.ActualCollectionName(config.BundleEventsCollection)).
		Find(ctx, filter, &results, mongodriver.Sort(sort), mongodriver.Offset(offset), mongodriver.Limit(limit))

	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

func buildListBundleEventsQuery(bundleID string, after, before *time.Time) (filter, sort bson.M) {
	filter = bson.M{}

	if bundleID != "" {

		filter["$or"] = []bson.M{
			{"bundle.id": bundleID},
			{"content_item.bundle_id": bundleID},
		}
	}

	if after != nil || before != nil {
		dateFilter := bson.M{}
		if after != nil {
			dateFilter["$gte"] = *after
		}
		if before != nil {
			dateFilter["$lte"] = *before
		}
		filter["created_at"] = dateFilter
	}

	sort = bson.M{"created_at": -1}
	return
}

// GetBundleEvent retrieves an event by Bundle ID
func (m *Mongo) GetBundleEvent(ctx context.Context, bundleID string) (*models.Event, error) {
	filter := buildGetBundleEventQuery(bundleID)

	var result models.Event
	err := m.Connection.Collection(m.ActualCollectionName(config.BundleEventsCollection)).
		FindOne(ctx, filter, &result)

	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, errors.New("bundle event not found")
		}
		return nil, err
	}

	return &result, nil
}

func buildGetBundleEventQuery(bundleID string) bson.M {
	return bson.M{"_id": bundleID}
}
