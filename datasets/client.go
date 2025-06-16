package datasets

import (
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type DatasetsAPIClient struct {
	client datasetAPISDK.Clienter

	versions DatasetsVersionsClient
}

type DatasetsClient interface {
	Versions() DatasetsVersionsClient
}

var _ DatasetsClient = (*DatasetsAPIClient)(nil)

func CreateDatasetsAPIClient(client datasetAPISDK.Clienter) DatasetsClient {
	versions := createVersionsClient(client)

	return &DatasetsAPIClient{
		client:   client,
		versions: versions,
	}
}

func (c *DatasetsAPIClient) Versions() DatasetsVersionsClient {
	return c.versions
}
