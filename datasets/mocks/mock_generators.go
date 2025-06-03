package datasetsmocks

import (
	"github.com/ONSdigital/dis-bundle-api/datasets"
)

func CreateDatasetsClientMock() datasets.DatasetsClient {
	return &DatasetsClientMock{}
}

func CreateDatasetsVersionsMock() datasets.DatasetsVersionsClient {
	return &DatasetsVersionsClientMock{}
}
