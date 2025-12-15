package models

import (
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
)

type AuthEntityData struct {
	EntityData    *permissionsAPISDK.EntityData
	IsServiceAuth bool
	Headers       datasetAPISDK.Headers
}

func CreateAuthEntityData(entityData *permissionsAPISDK.EntityData, serviceToken string, isServiceAuth bool) *AuthEntityData {
	headers := datasetAPISDK.Headers{
		AccessToken: serviceToken,
	}

	return &AuthEntityData{
		EntityData: entityData,
		//	IsServiceAuth: isServiceAuth,
		Headers: headers,
	}
}

func (a *AuthEntityData) GetUserID() string {
	return a.EntityData.UserID
}

func (a *AuthEntityData) GetUserEmail() string {
	return a.EntityData.UserID
}
