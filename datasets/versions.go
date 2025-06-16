package datasets

import (
	"context"
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	datasetmodels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type DatasetsAPIVersionsClient struct {
	client datasetAPISDK.Clienter
}

type DatasetsVersionsClient interface {
	GetForContentItem(ctx context.Context, r *http.Request, contentItem *models.ContentItem) (*datasetmodels.Version, error)
	UpdateStateForContentItem(ctx context.Context, r *http.Request, contentItem *models.ContentItem, targetState models.BundleState) error
}

var _ DatasetsVersionsClient = (*DatasetsAPIVersionsClient)(nil)

func createVersionsClient(client datasetAPISDK.Clienter) DatasetsVersionsClient {
	return &DatasetsAPIVersionsClient{
		client: client,
	}
}

func (c *DatasetsAPIVersionsClient) GetForContentItem(ctx context.Context, r *http.Request, contentItem *models.ContentItem) (*datasetmodels.Version, error) {
	versionID := contentItem.VersionIDString()
	version, err := c.client.GetVersion(ctx, CreateAuthHeaders(r), contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, versionID)

	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (c *DatasetsAPIVersionsClient) UpdateStateForContentItem(ctx context.Context, r *http.Request, contentItem *models.ContentItem, targetState models.BundleState) error {
	targetVersionState := convertToDatasetAPIState(targetState)
	err := c.client.PutVersionState(ctx, CreateAuthHeaders(r), contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, contentItem.VersionIDString(), targetVersionState)

	if err != nil {
		return err
	}

	return nil
}

func convertToDatasetAPIState(targetState models.BundleState) string {
	return strings.ToLower(targetState.String())
}
