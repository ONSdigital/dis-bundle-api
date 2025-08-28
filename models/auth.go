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
	// Need to set one value for either service token or florence token
	// Setting them both results in the florence token being checked first which doesn't work for service auth
	var headers datasetAPISDK.Headers
	if isServiceAuth {
		headers = datasetAPISDK.Headers{
			ServiceToken: serviceToken,
		}
	} else {
		headers = datasetAPISDK.Headers{
			UserAccessToken: serviceToken,
		}
	}

	return &AuthEntityData{
		EntityData:    entityData,
		IsServiceAuth: isServiceAuth,
		Headers:       headers,
	}
}

func (a *AuthEntityData) GetUserID() string {
	return a.EntityData.UserID
}

func (a *AuthEntityData) GetUserEmail() string {
	return a.EntityData.UserID
}
