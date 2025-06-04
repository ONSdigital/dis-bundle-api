package mongo

import (
	"context"

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
