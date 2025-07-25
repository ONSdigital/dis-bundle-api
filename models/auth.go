package models

import (
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
)

type AuthEntityData struct {
	EntityData *permissionsAPISDK.EntityData
	Headers    datasetAPISDK.Headers
}

func CreateAuthEntityData(entityData *permissionsAPISDK.EntityData, serviceToken string) *AuthEntityData {
	return &AuthEntityData{
		EntityData: entityData,
		Headers: datasetAPISDK.Headers{
			ServiceToken:    serviceToken,
			UserAccessToken: serviceToken,
		},
	}
}

func (a *AuthEntityData) GetUserID() string {
	return a.EntityData.UserID
}

func (a *AuthEntityData) GetUserEmail() string {
	return a.EntityData.UserID
}
