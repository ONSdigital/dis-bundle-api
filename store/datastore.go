package store

import (
	"context"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out datastoretest/mongo.go -pkg storetest . MongoDB
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

// DataStore provides a datastore.Storer interface used to store, retrieve, remove or update datasets
type DataStore struct {
	Backend Storer
}

type dataMongoDB interface {
	ListBundles(ctx context.Context, offset, limit int) (bundles []*models.Bundle, totalCount int, err error)
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

func (ds *DataStore) ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
	return ds.Backend.ListBundles(ctx, offset, limit)
}
