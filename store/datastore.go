package store

import (
	"context"
	"time"

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
	ListBundleEvents(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error)
	GetBundle(ctx context.Context, bundleID string) (*models.Bundle, error)
	UpdateBundleETag(ctx context.Context, bundleID, email string) (*models.Bundle, error)
	CheckBundleExists(ctx context.Context, bundleID string) (bool, error)
	CreateContentItem(ctx context.Context, contentItem *models.ContentItem) error
	CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error)
	CheckContentItemExistsByDatasetEditionVersion(ctx context.Context, datasetID, editionID string, versionID int) (bool, error)
	CreateBundleEvent(ctx context.Context, event *models.Event) error
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

func (ds *Datastore) ListBundleEvents(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
	return ds.Backend.ListBundleEvents(ctx, offset, limit, bundleID, after, before)
}
func (ds *Datastore) GetBundle(ctx context.Context, bundleID string) (*models.Bundle, error) {
	return ds.Backend.GetBundle(ctx, bundleID)
}

func (ds *Datastore) UpdateBundleETag(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
	return ds.Backend.UpdateBundleETag(ctx, bundleID, email)
}

func (ds *Datastore) CheckBundleExists(ctx context.Context, bundleID string) (bool, error) {
	return ds.Backend.CheckBundleExists(ctx, bundleID)
}

func (ds *Datastore) CreateContentItem(ctx context.Context, contentItem *models.ContentItem) error {
	return ds.Backend.CreateContentItem(ctx, contentItem)
}

func (ds *Datastore) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	return ds.Backend.CheckAllBundleContentsAreApproved(ctx, bundleID)
}

func (ds *Datastore) CheckContentItemExistsByDatasetEditionVersion(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
	return ds.Backend.CheckContentItemExistsByDatasetEditionVersion(ctx, datasetID, editionID, versionID)
}

func (ds *Datastore) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	return ds.Backend.CreateBundleEvent(ctx, event)
}
