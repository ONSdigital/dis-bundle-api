package datasetstests

import (
	"github.com/ONSdigital/dis-bundle-api/datasets"
	datasetsmocks "github.com/ONSdigital/dis-bundle-api/datasets/mocks"
)

func CreateDatasetsClientMock() datasets.DatasetsClient {
	return &datasetsmocks.DatasetsClientMock{}
}

func CreateDatasetsVersionsMock() datasets.DatasetsVersionsClient {
	return &datasetsmocks.DatasetsVersionsClientMock{}
}
