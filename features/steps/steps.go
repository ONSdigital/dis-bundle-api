package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	ctx.Step(`^there are no bundles$`, c.thereAreNoBundles)
	ctx.Step(`^I have these bundles:$`, c.iHaveTheseBundles)
	ctx.Step(`^I am an admin user$`, c.adminJWTToken)
	ctx.Step(`^there are no bundle events$`, c.thereAreNoBundleEvents)
	ctx.Step(`^I have these bundle events:$`, c.iHaveTheseBundleEvents)
	ctx.Step(`^the response header "([^"]*)" should be present$`, c.theResponseHeaderShouldBePresent)
	ctx.Step(`^the response should contain:$`, c.theResponseShouldContain)
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

func (c *BundleComponent) thereAreNoBundleEvents() error {
	ctx := context.Background()
	bundleEventsCollection := c.MongoClient.ActualCollectionName("BundleEventsCollection")

	_, err := c.MongoClient.Connection.Collection(bundleEventsCollection).DeleteMany(ctx, bson.M{})
	return err
}

func (c *BundleComponent) iHaveTheseBundleEvents(eventsJSON *godog.DocString) error {
	ctx := context.Background()

	var mapEvents []map[string]interface{}
	err := json.Unmarshal([]byte(eventsJSON.Content), &mapEvents)
	if err != nil {
		return fmt.Errorf("failed to unmarshal events JSON: %w", err)
	}

	bundleEventsCollection := c.MongoClient.ActualCollectionName("BundleEventsCollection")

	for _, event := range mapEvents {
		if err := c.putBundleEventInDatabase(ctx, bundleEventsCollection, event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return nil
}

func (c *BundleComponent) putBundleEventInDatabase(ctx context.Context, collectionName string, event map[string]interface{}) error {
	if event["_id"] == nil {
		event["_id"] = primitive.NewObjectID()
	}

	if createdAtStr, ok := event["created_at"].(string); ok {
		if _, err := time.Parse(time.RFC3339, createdAtStr); err != nil {
			return fmt.Errorf("failed to parse created_at: %w", err)
		}
		event["created_at"] = createdAtStr
	}

	_, err := c.MongoClient.Connection.Collection(collectionName).InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

func (c *BundleComponent) theResponseHeaderShouldBePresent(headerName string) error {
	if c.apiFeature.HTTPResponse == nil {
		return fmt.Errorf("no HTTP response available")
	}

	headerValue := c.apiFeature.HTTPResponse.Header.Get(headerName)
	if headerValue == "" {
		return fmt.Errorf("expected header '%s' to be present, but it was not found", headerName)
	}

	return nil
}

func (c *BundleComponent) theResponseShouldContain(expectedJSON *godog.DocString) error {
	return c.apiFeature.IShouldReceiveTheFollowingJSONResponse(expectedJSON)
}
