package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	ctx.Step(`^I am not authenticated$`, c.iAmNotAuthenticated)
	ctx.Step(`^the response body should be empty$`, c.theResponseBodyShouldBeEmpty)
	ctx.Step(`^the response header "([^"]*)" should equal "([^"]*)"$`, c.theResponseHeaderShouldBe)
	ctx.Step(`^the response header "([^"]*)" should not be empty$`, c.theResponseHeaderShouldNotBeEmpty)
	ctx.Step(`^I should receive a JSON response with (\d+) item$`, c.iShouldReceiveAJSONResponseWithItems)
	ctx.Step(`^the first bundle in the response should have title "([^"]*)"$`, c.theJSONResponseShouldContain)
	// ctx.Step(`^an internal server error is returned$`, c.anInternalServerErrorIsReturned)
}

func (c *BundleComponent) thereAreNoBundles() error {
	return c.MongoClient.Connection.DropDatabase(context.Background())
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

func (c *BundleComponent) iShouldReceiveAJSONResponseWithItems(expectedCount int) error {
	var body struct {
		Items []map[string]interface{} `json:"items"`
	}
	bodyBytes, err := io.ReadAll(c.apiFeature.HTTPResponse.Body)

	if err = json.Unmarshal(bodyBytes, &body); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}
	if len(body.Items) != expectedCount {
		return fmt.Errorf("expected %d items, got %d", expectedCount, len(body.Items))
	}
	return nil
}

func (c *BundleComponent) theJSONResponseShouldContain(expectedTitle string) error {
	var body struct {
		Items []struct {
			Title string `json:"title"`
		} `json:"items"`
	}

	bodyBytes, err := io.ReadAll(c.apiFeature.HTTPResponse.Body)

	if err = json.Unmarshal(bodyBytes, &body); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}
	if len(body.Items) == 0 {
		return fmt.Errorf("response items list is empty")
	}
	if body.Items[0].Title != expectedTitle {
		return fmt.Errorf("expected first bundle title %q, got %q", expectedTitle, body.Items[0].Title)
	}
	return nil
}

// func (c *BundleComponent) anInternalServerErrorIsReturned() error {
// 	return nil
// }
