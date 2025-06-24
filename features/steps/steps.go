package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	ctx.Step(`^I am an admin user$`, c.adminJWTToken)
	ctx.Step(`^I am not authenticated$`, c.iAmNotAuthenticated)
	ctx.Step(`^I have these bundles:$`, c.iHaveTheseBundles)
	ctx.Step(`^I have these content items:$`, c.iHaveTheseContentItems)
	ctx.Step(`^I have these bundle events:$`, c.iHaveTheseBundleEvents)
	ctx.Step(`^the content item in the database for id "([^"]*)" should not exist$`, c.theContentItemInTheDatabaseForIDShouldNotExist)
	ctx.Step(`^the response should contain:$`, c.theResponseShouldContain)
	ctx.Step(`^the response body should be empty$`, c.theResponseBodyShouldBeEmpty)
	ctx.Step(`^the response header "([^"]*)" should equal "([^"]*)"$`, c.theResponseHeaderShouldBe)
	ctx.Step(`^the response header "([^"]*)" should not be empty$`, c.theResponseHeaderShouldNotBeEmpty)
	ctx.Step(`^the response header "([^"]*)" should contain "([^"]*)"$`, c.theResponseHeaderShouldContain)
	ctx.Step(`^the response header "([^"]*)" should be present$`, c.theResponseHeaderShouldBePresent)
	ctx.Step(`^I should receive the following ContentItem JSON response:$`, c.iShouldReceiveTheFollowingContentItemJSONResponse)
}

func (c *BundleComponent) adminJWTToken() error {
	err := c.apiFeature.ISetTheHeaderTo("Authorization", authorisationtest.AdminJWTToken)
	return err
}

func (c *BundleComponent) iAmNotAuthenticated() error {
	err := c.apiFeature.ISetTheHeaderTo("Authorization", "")
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
	// Set the etag (json omitted)
	bundle.ETag = "etag-" + bundle.ID
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

func (c *BundleComponent) iHaveTheseContentItems(contentItemsJSON *godog.DocString) error {
	ctx := context.Background()
	contentItems := []models.ContentItem{}

	err := json.Unmarshal([]byte(contentItemsJSON.Content), &contentItems)
	if err != nil {
		return err
	}

	bundleContentsCollection := c.MongoClient.ActualCollectionName("BundleContentsCollection")

	for contentItem := range contentItems {
		if err := c.putContentItemInDatabase(ctx, bundleContentsCollection, contentItems[contentItem]); err != nil {
			return err
		}
	}
	return nil
}

func (c *BundleComponent) putContentItemInDatabase(ctx context.Context, collectionName string, contentItem models.ContentItem) error {
	update := bson.M{
		"$set": contentItem,
	}

	_, err := c.MongoClient.Connection.Collection(collectionName).UpsertById(ctx, contentItem.ID, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *BundleComponent) theResponseBodyShouldBeEmpty() error {
	if c.apiFeature.HTTPResponse == nil || c.apiFeature.HTTPResponse.Body == nil {
		return fmt.Errorf("response or body is nil")
	}

	bodyBytes, err := io.ReadAll(c.apiFeature.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Reset the body so other steps can still use it
	c.apiFeature.HTTPResponse.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bytes.TrimSpace(bodyBytes)) > 0 {
		return fmt.Errorf("expected empty body, but got: %q", string(bodyBytes))
	}
	return nil
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
		parsedTime, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return fmt.Errorf("failed to parse created_at: %w", err)
		}
		event["created_at"] = parsedTime
	}

	_, err := c.MongoClient.Connection.Collection(collectionName).InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

func (c *BundleComponent) theContentItemInTheDatabaseForIDShouldNotExist(id string) error {
	bundleContentsCollection := c.MongoClient.ActualCollectionName("BundleContentsCollection")

	var contentItem models.ContentItem

	err := c.MongoClient.Connection.Collection(bundleContentsCollection).FindOne(context.Background(), bson.M{"id": id}, contentItem)

	if err == nil {
		return fmt.Errorf("expected content item with ID %q to not exist, but it was found in the database: %+v", id, contentItem)
	} else if errors.Is(err, mongodriver.ErrNoDocumentFound) {
		return nil
	} else {
		return fmt.Errorf("error checking for content item with ID %q: %w", id, err)
	}
}

func (c *BundleComponent) theResponseHeaderShouldNotBeEmpty(header string) error {
	value := c.apiFeature.HTTPResponse.Header.Get(header)
	if value == "" {
		return fmt.Errorf("expected non-empty header %q but got empty", header)
	}
	return nil
}

func (c *BundleComponent) theResponseHeaderShouldBe(header, expected string) error {
	actual := c.apiFeature.HTTPResponse.Header.Get(header)
	if actual != expected {
		return fmt.Errorf("expected header %q to be %q, got %q", header, expected, actual)
	}
	return nil
}

func (c *BundleComponent) theResponseHeaderShouldContain(headerName, expectedValue string) error {
	value := c.apiFeature.HTTPResponse.Header.Get(headerName)
	if !strings.Contains(value, expectedValue) {
		return fmt.Errorf("expected header %q to contain %q, but got %q", headerName, expectedValue, value)
	}
	return nil
}

func (c *BundleComponent) iShouldReceiveTheFollowingContentItemJSONResponse(expectedJSON *godog.DocString) error {
	var expectedContentItem models.ContentItem
	if err := json.Unmarshal([]byte(expectedJSON.Content), &expectedContentItem); err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	bodyBytes, err := io.ReadAll(c.apiFeature.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var actualContentItem models.ContentItem
	if err = json.Unmarshal(bodyBytes, &actualContentItem); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if actualContentItem.ID == "" ||
		actualContentItem.BundleID != expectedContentItem.BundleID ||
		actualContentItem.ContentType != expectedContentItem.ContentType ||
		actualContentItem.Metadata.DatasetID != expectedContentItem.Metadata.DatasetID ||
		actualContentItem.Metadata.EditionID != expectedContentItem.Metadata.EditionID ||
		actualContentItem.Metadata.VersionID != expectedContentItem.Metadata.VersionID ||
		actualContentItem.Metadata.Title != expectedContentItem.Metadata.Title ||
		actualContentItem.State != expectedContentItem.State ||
		actualContentItem.Links.Edit != expectedContentItem.Links.Edit ||
		actualContentItem.Links.Preview != expectedContentItem.Links.Preview {
		return fmt.Errorf("actual content item does not match expected content item:\nExpected: %+v\nActual: %+v", expectedContentItem, actualContentItem)
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
