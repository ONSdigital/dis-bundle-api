package store

import (
	"context"
	"time"

	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

// Datastore provides a datastore.Storer interface used to store, retrieve, remove or update bundles
//
//go:generate moq -out datastoretest/mongo.go -pkg storetest . MongoDB
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

type Datastore struct {
	Backend Storer
}

type dataMongoDB interface {
	ListBundles(ctx context.Context, offset, limit int, filters *filters.BundleFilters) (bundles []*models.Bundle, totalCount int, err error)
	ListBundleEvents(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error)
	GetBundle(ctx context.Context, bundleID string) (*models.Bundle, error)
	CreateBundle(ctx context.Context, bundle *models.Bundle) error
	DeleteBundle(ctx context.Context, id string) (err error)
	CheckBundleExistsByTitle(ctx context.Context, title string) (bool, error)
	UpdateBundleETag(ctx context.Context, bundleID, email string) (*models.Bundle, error)
	CheckBundleExists(ctx context.Context, bundleID string) (bool, error)
	GetContentItemByBundleIDAndContentItemID(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error)
	CreateContentItem(ctx context.Context, contentItem *models.ContentItem) error
	CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error)
	CheckContentItemExistsByDatasetEditionVersion(ctx context.Context, datasetID, editionID string, versionID int) (bool, error)
	DeleteContentItem(ctx context.Context, contentItemID string) error
	ListBundleContents(ctx context.Context, bundleID string, offset, limit int) ([]*models.ContentItem, int, error)
	ListBundleContentsWithoutLimit(ctx context.Context, bundleID string) (contents []*models.ContentItem, err error)
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

func (ds *Datastore) ListBundles(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
	return ds.Backend.ListBundles(ctx, offset, limit, filters)
}

func (ds *Datastore) ListBundleEvents(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
	return ds.Backend.ListBundleEvents(ctx, offset, limit, bundleID, after, before)
}
func (ds *Datastore) GetBundle(ctx context.Context, bundleID string) (*models.Bundle, error) {
	return ds.Backend.GetBundle(ctx, bundleID)
}

func (ds *Datastore) CreateBundle(ctx context.Context, bundle *models.Bundle) error {
	return ds.Backend.CreateBundle(ctx, bundle)
}

func (ds *Datastore) DeleteBundle(ctx context.Context, id string) (err error) {
	return ds.Backend.DeleteBundle(ctx, id)
}

func (ds *Datastore) UpdateBundleETag(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
	return ds.Backend.UpdateBundleETag(ctx, bundleID, email)
}

func (ds *Datastore) CheckBundleExists(ctx context.Context, bundleID string) (bool, error) {
	return ds.Backend.CheckBundleExists(ctx, bundleID)
}

func (ds *Datastore) GetContentItemByBundleIDAndContentItemID(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
	return ds.Backend.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)
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

func (ds *Datastore) DeleteContentItem(ctx context.Context, contentItemID string) error {
	return ds.Backend.DeleteContentItem(ctx, contentItemID)
}

func (ds *Datastore) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	return ds.Backend.CreateBundleEvent(ctx, event)
}

func (ds *Datastore) CheckBundleExistsByTitle(ctx context.Context, title string) (bool, error) {
	return ds.Backend.CheckBundleExistsByTitle(ctx, title)
}

func (ds *Datastore) ListBundleContents(ctx context.Context, bundleID string, offset, limit int) ([]*models.ContentItem, int, error) {
	return ds.Backend.ListBundleContents(ctx, bundleID, offset, limit)
}

func (ds *Datastore) ListBundleContentsWithoutLimit(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
	return ds.Backend.ListBundleContentsWithoutLimit(ctx, bundleID)
}
