package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

// ListBundles retrieves all bundles based on the provided offset, limit, and BundleFilters
func (m *Mongo) ListBundles(ctx context.Context, offset, limit int, filters *filters.BundleFilters) (bundles []*models.Bundle, totalCount int, err error) {
	bundles = []*models.Bundle{}

	filter, sort := buildListBundlesQuery(filters)

	totalCount, err = m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).
		Find(ctx, filter, &bundles, mongodriver.Sort(sort), mongodriver.Offset(offset), mongodriver.Limit(limit))

	if err != nil {
		return nil, 0, err
	}

	return bundles, totalCount, nil
}

// buildListBundlesQuery Builds the MongoDB filter query based on the supplied BundleFilters value
func buildListBundlesQuery(filters *filters.BundleFilters) (filter, sort bson.M) {
	filter = bson.M{}
	sort = bson.M{"updated_at": -1}

	if filters == nil {
		return filter, sort
	}

	if filters.PublishDate != nil {
		filter["scheduled_at"] = buildDateTimeFilter(*filters.PublishDate)
	}

	return filter, sort
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

	now := time.Now()

	update.UpdatedAt = &now
	bytes, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	update.ETag = update.GenerateETag(&bytes)

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
			"e_tag":           update.ETag,
		},
	}

	_, err = m.Connection.Collection(collectionName).UpdateOne(ctx, filter, updateData)
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

	etag := bundleUpdate.GenerateETag(&bundleUpdateJSON)

	filter := bson.M{"id": bundleID}

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
	filter := bson.M{"id": bundleID}
	count, err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *Mongo) CheckBundleExistsByTitleUpdate(ctx context.Context, title, excludeID string) (bool, error) {
	filter := bson.M{
		"title": title,
		"id":    bson.M{"$ne": excludeID},
	}

	count, err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Count(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckBundleExistsByTitle checks if a bundle exists by its title
func (m *Mongo) CheckBundleExistsByTitle(ctx context.Context, title string) (bool, error) {
	filter := bson.M{"title": title}

	count, err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Count(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
