package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/cucumber/godog"
	"gopkg.in/mgo.v2/bson"
)

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	ctx.Step(`^I have these bundles:$`, c.iHaveTheseBundles)
	ctx.Step(`^I have these content items:$`, c.iHaveTheseContentItems)
	ctx.Step(`^I am an admin user$`, c.adminJWTToken)
	ctx.Step(`^I am not authenticated$`, c.iAmNotAuthenticated)
	ctx.Step(`^the response body should be empty$`, c.theResponseBodyShouldBeEmpty)
	ctx.Step(`^the response header "([^"]*)" should equal "([^"]*)"$`, c.theResponseHeaderShouldBe)
	ctx.Step(`^the response header "([^"]*)" should not be empty$`, c.theResponseHeaderShouldNotBeEmpty)
	ctx.Step(`^the response header "([^"]*)" should contain "([^"]*)"$`, c.theResponseHeaderShouldContain)
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

	for contentItem := range contentItems {
		bundleContentsCollection := c.MongoClient.ActualCollectionName("BundleContentsCollection")
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
