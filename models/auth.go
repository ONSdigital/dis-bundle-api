package models

import "github.com/ONSdigital/dp-permissions-api/sdk"

type AuthEntityData struct {
	EntityData   *sdk.EntityData
	ServiceToken string
}

func CreateAuthEntityData(entityData *sdk.EntityData, serviceToken string) *AuthEntityData {
	return &AuthEntityData{
		EntityData:   entityData,
		ServiceToken: serviceToken,
	}
}

func (a *AuthEntityData) GetUserID() string {
	if a.EntityData != nil {
		return a.EntityData.UserID
	}

	return ""
}

func (a *AuthEntityData) GetUserEmail() string {
	if a.EntityData != nil {
		return a.EntityData.UserID
	}

	return ""
}
