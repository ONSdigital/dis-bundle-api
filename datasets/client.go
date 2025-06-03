package datasets

import (
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type DatasetsApiClient struct {
	client datasetAPISDK.Clienter

	versions DatasetsVersionsClient
}

type DatasetsClient interface {
	Versions() DatasetsVersionsClient
}

var _ DatasetsClient = (*DatasetsApiClient)(nil)

func CreateDatasetsApiClient(client datasetAPISDK.Clienter) DatasetsClient {
	versions := createVersionsClient(client)

	return &DatasetsApiClient{
		client:   client,
		versions: versions,
	}
}

func (c *DatasetsApiClient) Versions() DatasetsVersionsClient {
	return c.versions
}
