package datasets

import (
	"context"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	datasetmodels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type DatasetsApiVersionsClient struct {
	headers *datasetAPISDK.Headers
	client  datasetAPISDK.Clienter
}

type DatasetsVersionsClient interface {
	GetForContentItem(ctx context.Context, contentItem models.ContentItem) (*datasetmodels.Version, error)
	UpdateStateForContentItem(ctx context.Context, contentItem models.ContentItem, targetState models.BundleState) error
}

var _ DatasetsVersionsClient = (*DatasetsApiVersionsClient)(nil)

func createVersionsClient(client datasetAPISDK.Clienter, headers *datasetAPISDK.Headers) DatasetsVersionsClient {
	return &DatasetsApiVersionsClient{
		client:  client,
		headers: headers,
	}
}

func (c *DatasetsApiVersionsClient) GetForContentItem(ctx context.Context, contentItem models.ContentItem) (*datasetmodels.Version, error) {
	versionId := contentItem.VersionIdString()
	version, err := c.client.GetVersion(ctx, *c.headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, versionId)

	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (c *DatasetsApiVersionsClient) UpdateStateForContentItem(ctx context.Context, contentItem models.ContentItem, targetState models.BundleState) error {
	targetVersionState := strings.ToLower(targetState.String())
	err := c.client.PutVersionState(ctx, *c.headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, contentItem.VersionIdString(), targetVersionState)

	if err != nil {
		return err
	}

	return nil
}
