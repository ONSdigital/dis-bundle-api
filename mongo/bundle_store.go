package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

// ListBundles retrieves all bundles based on the provided offset and limit
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
	sort = bson.M{"updated_at": -1}
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
		return nil, err
	}
	return &result, nil
}

func buildGetBundleQuery(bundleID string) bson.M {
	return bson.M{"id": bundleID}
}

// CreateBundle inserts a new bundle
func (m *Mongo) CreateBundle(ctx context.Context, bundle *models.Bundle) error {
	now := time.Now()
	bundle.CreatedAt = &now
	collectionName := m.ActualCollectionName(config.BundlesCollection)

	_, err := m.Connection.Collection(collectionName).InsertOne(ctx, bundle)

	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) UpdateBundle(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
	collectionName := m.ActualCollectionName(config.BundlesCollection)
	filter := bson.M{"id": id}

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

// UpdateBundleETag updates the ETag, last_updated_by, and updated_at fields of a bundle
func (m *Mongo) UpdateBundleETag(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
	bundleUpdate, err := m.GetBundle(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	bundleUpdate.LastUpdatedBy.Email = email
	bundleUpdate.UpdatedAt = &now

	bundleUpdateJSON, err := json.Marshal(bundleUpdate)
	if err != nil {
		return nil, err
	}

	etag := dpresponse.GenerateETag(bundleUpdateJSON, false)

	filter := bson.M{"_id": bundleID}

	updateData := bson.M{
		"$set": bson.M{
			"last_updated_by.email": email,
			"updated_at":            now,
			"e_tag":                 etag,
		},
	}

	_, err = m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).UpdateOne(ctx, filter, updateData)
	if err != nil {
		return nil, err
	}

	return m.GetBundle(ctx, bundleID)
}

// DeleteBundle deletes a bundle by ID
func (m *Mongo) DeleteBundle(ctx context.Context, id string) (err error) {
	if _, err = m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Must().DeleteOne(ctx, bson.D{{Key: "id", Value: id}}); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return apierrors.ErrBundleNotFound
		}
		return err
	}

	log.Info(ctx, "bundle deleted", log.Data{"id": id})
	return nil
}

// CheckBundleExists checks if a bundle exists by ID
func (m *Mongo) CheckBundleExists(ctx context.Context, bundleID string) (bool, error) {
	filter := bson.M{"_id": bundleID}
	count, err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
