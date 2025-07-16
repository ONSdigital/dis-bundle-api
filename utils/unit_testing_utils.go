package utils

import "github.com/ONSdigital/dis-bundle-api/models"

// PtrContentItemState returns a pointer for content item state
func PtrContentItemState(s models.State) *models.State {
	return &s
}
