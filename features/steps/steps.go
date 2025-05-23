package steps

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
)

func WellKnownTestTime() time.Time {
	testTime, _ := time.Parse("2006-01-02T15:04:05Z", "2021-01-01T00:00:00Z")
	return testTime
}

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^there are no bundles$`, c.thereAreNoBundles)
	ctx.Step(`^I have these bundles:$`, c.iHaveTheseBundles)
}

func (c *BundleComponent) thereAreNoBundles() error {
	return c.MongoClient.Connection.DropDatabase(context.Background())
}

func (c *BundleComponent) iHaveTheseBundles(bundlesJSON *godog.DocString) error {
	var bundles []models.Bundle
	if err := json.Unmarshal([]byte(bundlesJSON.Content), &bundles); err != nil {
		return err
	}

	for timeOffset := range bundles {
		bundle := &bundles[timeOffset]
		bundleId := bundle.ID
		// Set the etag (json omitted)
		bundle.ETag = "etag-" + bundle.ID

		bundleCollection := c.MongoClient.ActualCollectionName(config.BundlesCollection)
		if err := c.putDocumentInDatabase(bundle, bundleId, bundleCollection, timeOffset); err != nil {
			return err
		}
	}
	return nil
}

func (c *BundleComponent) putDocumentInDatabase(document interface{}, id, collectionName string, timeOffset int) error {
	update := bson.M{
		"$set": document,
		"$setOnInsert": bson.M{
			"last_updated": WellKnownTestTime().Add(time.Second * time.Duration(timeOffset)),
		},
	}

	_, err := c.MongoClient.Connection.Collection(collectionName).UpsertById(context.Background(), id, update)

	if err != nil {
		return err
	}
	return nil
}
