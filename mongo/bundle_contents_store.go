package mongo

import (
	"context"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
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
