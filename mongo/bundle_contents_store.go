package mongo

import (
	"context"
	"errors"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// GetContentItemByBundleIDAndContentItemID retrieves a content item by bundle ID and content item ID
func (m *Mongo) GetContentItemByBundleIDAndContentItemID(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
	filter := bson.M{
		"id":        contentItemID,
		"bundle_id": bundleID,
	}

	var result models.ContentItem
	err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).
		FindOne(ctx, filter, &result)

	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, apierrors.ErrContentItemNotFound
		}
		return nil, err
	}
	return &result, nil
}

// CreateContentItem inserts a new content item into the database
func (m *Mongo) CreateContentItem(ctx context.Context, contentItem *models.ContentItem) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).InsertOne(ctx, contentItem)
	return err
}

// CheckAllBundleContentsAreApproved checks if all contents of a bundle are in the approved state
func (m *Mongo) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	filter := bson.M{
		"bundle_id": bundleID,
		"state":     bson.M{"$ne": "APPROVED"},
	}

	count, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).Count(ctx, filter)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return false, nil
	}
	return true, nil
}

// CheckContentItemExistsByDatasetEditionVersion checks if a content item exists with the specified dataset, edition, and version
func (m *Mongo) CheckContentItemExistsByDatasetEditionVersion(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
	filter := bson.M{
		"metadata.dataset_id": datasetID,
		"metadata.edition_id": editionID,
		"metadata.version_id": versionID,
	}

	count, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).Count(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetBundleContentsForBundle retrieves all ContentItems for the specified bundle
func (m *Mongo) GetBundleContentsForBundle(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
	filter := bson.M{
		"bundle_id": bundleID,
	}

	var contents []models.ContentItem
	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).Find(ctx, filter, &contents)
	if err != nil {
		return nil, err
	}

	return &contents, nil
}

// DeleteContentItem removes a content item by its ID
func (m *Mongo) DeleteContentItem(ctx context.Context, contentItemID string) error {
	result, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).
		DeleteOne(ctx, bson.M{"id": contentItemID})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return apierrors.ErrContentItemNotFound
	}

	return nil
}

// UpdateContentItemState updates the state for the content item matching the contentItemID
func (m *Mongo) UpdateContentItemState(ctx context.Context, contentItemID, state string) error {
	filter := bson.M{
		"id": contentItemID,
	}

	updateData := bson.M{
		"$set": bson.M{
			"state": state,
		},
	}

	result, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).UpdateOne(ctx, filter, updateData)

	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("no documents were modified")
	}

	return nil
}

func (m *Mongo) ListBundleContents(ctx context.Context, bundleID string, offset, limit int) (contents []*models.ContentItem, totalCount int, err error) {
	var results []*models.ContentItem

	filter, sort := buildListBundleContentsQuery(bundleID)

	totalCount, err = m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).
		Find(ctx, filter, &results, mongodriver.Sort(sort), mongodriver.Offset(offset), mongodriver.Limit(limit))

	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

func buildListBundleContentsQuery(bundleID string) (filter, sort bson.M) {
	filter = bson.M{}

	if bundleID != "" {
		filter["bundle_id"] = bundleID
	}

	sort = bson.M{"id": -1}
	return
}
