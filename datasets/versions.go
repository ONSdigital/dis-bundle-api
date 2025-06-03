package datasets

import (
	"context"
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	datasetmodels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type DatasetsApiVersionsClient struct {
	client datasetAPISDK.Clienter
}

type DatasetsVersionsClient interface {
	GetForContentItem(ctx context.Context, r *http.Request, contentItem models.ContentItem) (*datasetmodels.Version, error)
	UpdateStateForContentItem(ctx context.Context, r *http.Request, contentItem models.ContentItem, targetState models.BundleState) error
}

var _ DatasetsVersionsClient = (*DatasetsApiVersionsClient)(nil)

func createVersionsClient(client datasetAPISDK.Clienter) DatasetsVersionsClient {
	return &DatasetsApiVersionsClient{
		client: client,
	}
}

func (c *DatasetsApiVersionsClient) GetForContentItem(ctx context.Context, r *http.Request, contentItem models.ContentItem) (*datasetmodels.Version, error) {
	versionId := contentItem.VersionIdString()
	version, err := c.client.GetVersion(ctx, CreateAuthHeaders(r), contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, versionId)

	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (c *DatasetsApiVersionsClient) UpdateStateForContentItem(ctx context.Context, r *http.Request, contentItem models.ContentItem, targetState models.BundleState) error {
	targetVersionState := convertToDatasetAPIState(targetState)
	err := c.client.PutVersionState(ctx, CreateAuthHeaders(r), contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, contentItem.VersionIdString(), targetVersionState)

	if err != nil {
		return err
	}

	return nil
}

func convertToDatasetAPIState(targetState models.BundleState) string {
	return strings.ToLower(targetState.String())
}
