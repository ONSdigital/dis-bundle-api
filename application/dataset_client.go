package application

import (
	"context"

	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

func (s *StateMachineBundleAPI) GetDataset(ctx context.Context, authHeaders datasetAPISDK.Headers, datasetID string) (datasetAPIModels.Dataset, error) {
	return s.DatasetAPIClient.GetDataset(ctx, authHeaders, "", datasetID)
}

func (s *StateMachineBundleAPI) GetVersion(ctx context.Context, authHeaders datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
	return s.DatasetAPIClient.GetVersion(ctx, authHeaders, datasetID, editionID, versionID)
}

func (s *StateMachineBundleAPI) PutVersion(ctx context.Context, authHeaders datasetAPISDK.Headers, datasetID, editionID, versionID string, version datasetAPIModels.Version) (datasetAPIModels.Version, error) {
	return s.DatasetAPIClient.PutVersion(ctx, authHeaders, datasetID, editionID, versionID, version)
}
