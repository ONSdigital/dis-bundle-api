package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
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

func buildListBundlesQuery() (filter bson.M, sort bson.M) {
	filter = bson.M{} // No filters yet; future: filter by state/title/etc.
	sort = bson.M{"id": -1}
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
	bundle.CreatedDate = time.Now()
	collectionName := m.ActualCollectionName("BundlesCollection")

	_, err := m.Connection.Collection(collectionName).Insert(ctx, bundle)

	if err != nil {
		return err
	}
	return nil
}

// UpdateBundle updates or inserts a bundle
func (m *Mongo) UpdateBundle(ctx context.Context, id string, bundle *models.Bundle) error {
	update := bundleUpdateQuery(bundle)

	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).UpsertById(ctx, id, update)

	if err != nil {
		return err
	}
	return nil
}

func bundleUpdateQuery(bundle *models.Bundle) bson.M {
	return bson.M{
		"$set":         bundle,
		"$setOnInsert": bson.M{"created_date": bundle.CreatedDate},
	}
}

// DeleteBundle deletes a bundle by ID
func (m *Mongo) DeleteBundle(ctx context.Context, id string) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundlesCollection)).Must().DeleteById(ctx, id)

	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return apierrors.ErrBundleNotFound
		}
		return err
	}
	return nil
}
