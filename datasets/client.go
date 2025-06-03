package datasets

import (
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type DatasetsApiClient struct {
	headers *datasetAPISDK.Headers
	client  datasetAPISDK.Clienter

	versions DatasetsVersionsClient
}

type DatasetsClient interface {
	Versions() DatasetsVersionsClient
}

var _ DatasetsClient = (*DatasetsApiClient)(nil)

func CreateDatasetsApiClient(client datasetAPISDK.Clienter, headers datasetAPISDK.Headers) DatasetsClient {
	versions := createVersionsClient(client, &headers)

	return &DatasetsApiClient{
		client:   client,
		headers:  &headers,
		versions: versions,
	}
}

func (c *DatasetsApiClient) Versions() DatasetsVersionsClient {
	return c.versions
}
