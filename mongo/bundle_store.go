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

// ListBundles retrieves all bundles
func (m *Mongo) ListBundles(ctx context.Context, offset, limit int) (bundles []*models.Bundle, totalCount int, err error) {
	bundles = []*models.Bundle{}

	filter, sort := buildListBundlesQuery()

	totalCount, err = m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).
		Find(ctx, filter, &bundles, mongodriver.Sort(sort), mongodriver.Offset(offset), mongodriver.Limit(limit))

	if err != nil {
		return nil, 0, err
	}

	return bundles, totalCount, nil
}

func buildListBundlesQuery() (filter, sort bson.M) {
	filter = bson.M{}
	sort = bson.M{"_id": 1}
	return
}

// GetBundle retrieves a single bundle by ID
func (m *Mongo) GetBundle(ctx context.Context, bundleID string) (*models.Bundle, error) {
	filter := buildGetBundleQuery(bundleID)

	var result models.Bundle
	err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).
		FindOne(ctx, filter, &result)

	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, apierrors.ErrBundleNotFound
		}
	}
	return &result, nil
}

func buildGetBundleQuery(bundleID string) bson.M {
	return bson.M{"_id": bundleID}
}

// CreateBundle inserts a new bundle
func (m *Mongo) CreateBundle(ctx context.Context, bundle *models.Bundle) error {
	now := time.Now()
	bundle.CreatedAt = &now
	collectionName := m.ActualCollectionName("BundlesCollection")

	_, err := m.Connection.Collection(collectionName).Insert(ctx, bundle)

	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) UpdateBundle(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
	collectionName := m.ActualCollectionName("BundlesCollection")
	filter := bson.M{"_id": id}

	updateData := bson.M{
		"$set": bson.M{
			"bundle_type":     update.BundleType,
			"created_by":      update.CreatedBy,
			"created_at":      update.CreatedAt,
			"last_updated_by": update.LastUpdatedBy,
			"preview_teams":   update.PreviewTeams,
			"scheduled_at":    update.ScheduledAt,
			"state":           update.State,
			"title":           update.Title,
			"updated_at":      update.UpdatedAt,
			"managed_by":      update.ManagedBy,
		},
	}

	_, err := m.Connection.Collection(collectionName).UpdateOne(ctx, filter, updateData)
	if err != nil {
		return nil, err
	}

	// Re-fetch updated bundle to return full latest version
	return m.GetBundle(ctx, id)
}

// DeleteBundle deletes a bundle by ID
func (m *Mongo) DeleteBundle(ctx context.Context, id string) (err error) {
	if _, err = m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Must().Delete(ctx, bson.D{{Key: "_id", Value: id}}); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return apierrors.ErrBundleNotFound
		}
		return err
	}

	log.Info(ctx, "bundle deleted", log.Data{"_id": id})
	return nil
}
