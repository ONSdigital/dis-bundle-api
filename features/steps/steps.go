package steps

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/cucumber/godog"
	"gopkg.in/mgo.v2/bson"
)

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	ctx.Step(`^there are no bundles$`, c.thereAreNoBundles)
	ctx.Step(`^I have these bundles:$`, c.iHaveTheseBundles)
	ctx.Step(`^I am an admin user$`, c.adminJWTToken)
}

func (c *BundleComponent) thereAreNoBundles() error {
	return c.MongoClient.Connection.DropDatabase(context.Background())
}

func (c *BundleComponent) adminJWTToken() error {
	err := c.apiFeature.ISetTheHeaderTo("Authorization", authorisationtest.AdminJWTToken)
	return err
}

func (c *BundleComponent) iHaveTheseBundles(bundlesJSON *godog.DocString) error {
	ctx := context.Background()
	bundles := []models.Bundle{}

	err := json.Unmarshal([]byte(bundlesJSON.Content), &bundles)
	if err != nil {
		return err
	}

	for bundle := range bundles {
		bundlesCollection := c.MongoClient.ActualCollectionName("BundlesCollection")
		if err := c.putBundleInDatabase(ctx, bundlesCollection, bundles[bundle]); err != nil {
			return err
		}
	}

	return nil
}

func (c *BundleComponent) putBundleInDatabase(ctx context.Context, collectionName string, bundle models.Bundle) error {
	update := bson.M{
		"$set": bundle,
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}

	_, err := c.MongoClient.Connection.Collection(collectionName).UpsertById(ctx, bundle.ID, update)
	if err != nil {
		return err
	}
	return nil
}
