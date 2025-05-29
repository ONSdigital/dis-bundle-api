package mongo

import (
	"context"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
)

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
