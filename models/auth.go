package models

import (
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
)

type AuthEntityData struct {
	EntityData *permissionsAPISDK.EntityData
	Headers    datasetAPISDK.Headers
}

func CreateAuthEntityData(entityData *permissionsAPISDK.EntityData, serviceToken, florenceToken string) *AuthEntityData {
	// florenceToken is only used for local development, when in an environment we use the service token
	if florenceToken == "" {
		florenceToken = serviceToken
	}

	return &AuthEntityData{
		EntityData: entityData,
		Headers: datasetAPISDK.Headers{
			ServiceToken:    serviceToken,
			UserAccessToken: florenceToken,
		},
	}
}

func (a *AuthEntityData) GetUserID() string {
	return a.EntityData.UserID
}

func (a *AuthEntityData) GetUserEmail() string {
	return a.EntityData.UserID
}
