package mongo

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	"go.mongodb.org/mongo-driver/bson"
)

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

func (m *Mongo) GetContentsForBundle(ctx context.Context, bundleID string) ([]models.ContentItem, error) {
	filter := bson.M{
		"bundle_id": bundleID,
	}

	var bundles []models.ContentItem
	_, err := m.Connection.Collection(m.ActualCollectionName(config.BundleContentsCollection)).Find(ctx, filter, &bundles)
	if err != nil {
		return nil, err
	}

	return bundles, nil
}

// UpdateBundleContentItemState updates the state for the
func (m *Mongo) UpdateBundleContentItemState(ctx context.Context, contentItemID string, state models.BundleState) error {
	filter := bson.M{"id": contentItemID}

	updateData := bson.M{
		"$set": bson.M{
			"state": state,
		},
	}

	collectionName := m.ActualCollectionName(config.BundleContentsCollection)

	updateResult, err := m.Connection.Collection(collectionName).UpdateOne(ctx, filter, updateData)
	if err != nil {
		return err
	}

	if updateResult.ModifiedCount == 0 {
		return fmt.Errorf("no content items were modified. %d items were matched for content item ID %s", updateResult.MatchedCount, contentItemID)
	}

	return nil
}
