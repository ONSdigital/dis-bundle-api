package store

import (
	"context"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out datastoretest/mongo.go -pkg storetest . MongoDB
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

// Datastore provides a datastore.Storer interface used to store, retrieve, remove or update bundles
type Datastore struct {
	Backend Storer
}

type dataMongoDB interface {
	ListBundles(ctx context.Context, offset, limit int) (bundles []*models.Bundle, totalCount int, err error)
	GetBundleById(ctx context.Context, bundleId string) (bundle *models.Bundle, err error)
	CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Close(ctx context.Context) error
}

// MongoDB represents all the required methods from mongo DB
type MongoDB interface {
	dataMongoDB
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
}

// Storer represents basic data access via Get, Remove and Upsert methods, abstracting it from mongoDB or graphDB
type Storer interface {
	dataMongoDB
}

func (ds *Datastore) ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
	return ds.Backend.ListBundles(ctx, offset, limit)
}

func (ds *Datastore) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	return ds.Backend.CheckAllBundleContentsAreApproved(ctx, bundleID)
}

func (ds *Datastore) GetBundleById(ctx context.Context, bundleId string) (*models.Bundle, error) {
	return ds.Backend.GetBundleById(ctx, bundleId)
}
