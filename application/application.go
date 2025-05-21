package application

import (
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type StateMachineBundleAPI struct {
	Datastore    store.Datastore
	StateMachine *StateMachine
}

func Setup(datastore store.Datastore, stateMachine *StateMachine) *StateMachineBundleAPI {
	return &StateMachineBundleAPI{
		Datastore:    datastore,
		StateMachine: stateMachine,
	}
}

func checkAllBundleContentsAreApproved(contents []models.BundleContent) bool {
	for _, bundleContent := range contents {
		if bundleContent.State != Approved.String() {
			return false
		}
	}
	return true
}
