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

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)

	ctx.Step(`^I have these bundles:$`, c.iHaveTheseBundles)
	ctx.Step(`^I have these content items:$`, c.iHaveTheseContentItems)
	ctx.Step(`^I have these bundle events:$`, c.iHaveTheseBundleEvents)
	ctx.Step(`^I have these dataset versions:$`, c.iHaveTheseDatasetVersions)

	ctx.Step(`^the response should contain:$`, c.theResponseShouldContain)
	ctx.Step(`^the response should contain the following JSON response with a dynamic timestamp:$`, c.theResponseShouldContainTheFollowingJSONResponseWithADynamicTimestamp)
	ctx.Step(`^the response body should be empty$`, c.theResponseBodyShouldBeEmpty)

	ctx.Step(`^the response header "([^"]*)" should equal "([^"]*)"$`, c.theResponseHeaderShouldBe)
	ctx.Step(`^the response header "([^"]*)" should not be empty$`, c.theResponseHeaderShouldNotBeEmpty)
	ctx.Step(`^the response header "([^"]*)" should contain "([^"]*)"$`, c.theResponseHeaderShouldContain)
	ctx.Step(`^the response header "([^"]*)" should be present$`, c.theResponseHeaderShouldBePresent)

	ctx.Step(`^I set the header "([^"]*)" to "([^"]*)"$`, c.iSetTheHeaderTo)

	ctx.Step(`^the record with id "([^"]*)" should not exist in the "([^"]*)" collection$`, c.theRecordWithIDShouldNotExistInTheCollection)

	ctx.Step(`^bundle "([^"]*)" should have state "([^"]*)"`, c.bundleShouldHaveState)
	ctx.Step(`^bundle "([^"]*)" should have this etag "([^"]*)"$`, c.bundleETagShouldMatch)
	ctx.Step(`^bundle "([^"]*)" should not have this etag "([^"]*)"$`, c.bundleETagShouldNotMatch)

	ctx.Step(`^I should receive the following ContentItem JSON response:$`, c.iShouldReceiveTheFollowingContentItemJSONResponse)

	ctx.Step(`^the release date for the dataset version with id "([^"]*)" should be "([^"]*)"$`, c.theReleaseDateForDatasetVersionShouldBe)

	ctx.Step(`the total number of events should be (\d+)`, c.theTotalNumberOfEventsShouldBe)
	ctx.Step(`the number of events with action "([^"]*)" and datatype "([^"]*)" should be (\d+)`, c.theNumberOfEventsWithActionAndDatatypeShouldBe)

	ctx.Step(`^these content item states should match:$`, c.contentItemsShouldMatchState)
	ctx.Step(`^these dataset versions states should match:$`, c.theseVersionsShouldHaveTheseStates)

	// Policy assertions (permissions API)
	ctx.Step(`I have these policies:$`, c.iHaveThesePolicies)
	ctx.Step(`the following policies should exist:$`, c.theFollowingPoliciesShouldExist)
	ctx.Step(`the policy "([^"]*)" should have these condition values:$`, c.thePolicyShouldHaveTheseConditionValues)
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

func (c *BundleComponent) theRecordWithIDShouldNotExistInTheCollection(id, collection string) error {
	collectionMap := map[string]string{
		"bundles":         c.MongoClient.ActualCollectionName("BundlesCollection"),
		"bundle_contents": c.MongoClient.ActualCollectionName("BundleContentsCollection"),
		"bundle_events":   c.MongoClient.ActualCollectionName("BundleEventsCollection"),
	}

	collectionName, exists := collectionMap[collection]
	if !exists {
		return fmt.Errorf("unknown collection: %s", collection)
	}

	var result bson.M
	err := c.MongoClient.Connection.Collection(collectionName).FindOne(context.Background(), bson.M{"id": id}, &result)

	if err == nil {
		return fmt.Errorf("expected record with ID %q to not exist in collection %q, but it was found in the database: %+v", id, collectionName, result)
	} else if errors.Is(err, mongodriver.ErrNoDocumentFound) {
		return nil
	} else {
		return fmt.Errorf("error checking for record with ID %q in collection %q: %w", id, collectionName, err)
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

func (c *BundleComponent) theReleaseDateForDatasetVersionShouldBe(versionID, expectedReleaseDate string) error {
	for _, v := range c.DatasetAPIVersions {
		if v.ID == versionID {
			if v.ReleaseDate != expectedReleaseDate {
				return fmt.Errorf("expected release date %q for version %q, got %q", expectedReleaseDate, versionID, v.ReleaseDate)
			}
			return nil
		}
	}
	return fmt.Errorf("version with id %q not found", versionID)
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

func (c *BundleComponent) iSetTheHeaderTo(headerName, headerValue string) error {
	return c.apiFeature.ISetTheHeaderTo(headerName, headerValue)
}

func (c *BundleComponent) theResponseShouldContainTheFollowingJSONResponseWithADynamicTimestamp(expectedJSON *godog.DocString) error {
	b, err := io.ReadAll(c.apiFeature.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}
	c.apiFeature.HTTPResponse.Body = io.NopCloser(bytes.NewReader(b))

	var actual, expected map[string]interface{}
	if err := json.Unmarshal(b, &actual); err != nil {
		return fmt.Errorf("invalid actual JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(expectedJSON.Content), &expected); err != nil {
		return fmt.Errorf("invalid expected JSON: %w", err)
	}

	if expectedTimestamp, ok := expected["updated_at"].(string); ok && expectedTimestamp == "{{DYNAMIC_TIMESTAMP}}" {
		actualTimestampStr, ok := actual["updated_at"].(string)
		if !ok {
			return fmt.Errorf("missing or non-string updated_at in actual")
		}
		parsedTimestamp, err := time.Parse(time.RFC3339, actualTimestampStr)
		if err != nil {
			return fmt.Errorf("updated_at is not a valid RFC3339 timestamp: %w", err)
		}
		timestampAge := time.Since(parsedTimestamp)
		if timestampAge < 0 || timestampAge > 10*time.Second {
			return fmt.Errorf("updated_at %v is not within 10s of now", parsedTimestamp)
		}

		delete(actual, "updated_at")
		delete(expected, "updated_at")
	}

	got, _ := json.Marshal(actual)
	want, _ := json.Marshal(expected)
	if !bytes.Equal(got, want) {
		return fmt.Errorf("response mismatch:\nExpected: %s\nActual:   %s", want, got)
	}
	return nil
}

func (c *BundleComponent) iHaveTheseDatasetVersions(contentItemsJSON *godog.DocString) error {
	versions := []*datasetAPIModels.Version{}

	err := json.Unmarshal([]byte(contentItemsJSON.Content), &versions)
	if err != nil {
		return err
	}

	c.DatasetAPIVersions = append(c.DatasetAPIVersions, versions...)

	return nil
}

func (c *BundleComponent) bundleShouldHaveState(bundleID, expectedState string) error {
	bundle, err := c.getBundleByID(bundleID)
	if err != nil {
		return err
	}

	if bundle.State.String() != expectedState {
		return fmt.Errorf("expected state %s but actual state is %s", expectedState, bundle.State.String())
	}

	return nil
}

func (c *BundleComponent) getBundleByID(bundleID string) (*models.Bundle, error) {
	ctx := context.Background()
	bundlesCollection := c.MongoClient.ActualCollectionName("BundlesCollection")

	var bundle models.Bundle
	err := c.MongoClient.Connection.Collection(bundlesCollection).FindOne(ctx, bson.M{"id": bundleID}, &bundle)

	if err != nil {
		return nil, err
	}

	if bundle.ID != bundleID {
		return nil, errors.New("no error was returned fetching bundle but the bundle does not appear valid")
	}

	return &bundle, nil
}

func (c *BundleComponent) contentItemsShouldMatchState(expectedContentItemsJSON *godog.DocString) error {
	var expectedContentItems []models.ContentItem
	if err := json.Unmarshal([]byte(expectedContentItemsJSON.Content), &expectedContentItems); err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	ctx := context.Background()
	contentsCollection := c.MongoClient.ActualCollectionName(config.BundleContentsCollection)

	var actualContentItems []models.ContentItem
	_, err := c.MongoClient.Connection.Collection(contentsCollection).Find(ctx, bson.M{}, &actualContentItems)

	if err != nil {
		return err
	}

	if len(actualContentItems) == 0 {
		return errors.New("no content items returned")
	}

	idSelector := func(content models.ContentItem) string { return content.ID }
	predicate := func(actual, expected models.ContentItem) string {
		if actual.State.String() != expected.State.String() {
			return fmt.Sprintf("content item %s state does not match. expected %s but is actually %s", expected.ID, expected.State, actual.State)
		}

		return ""
	}

	errs := compareSlices(expectedContentItems, actualContentItems, predicate, idSelector)

	if len(errs) > 0 {
		return fmt.Errorf("content items do not match:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

func (c *BundleComponent) bundleETagShouldMatch(bundleID, etag string) error {
	return c.bundleETagValueMatch(bundleID, etag, true)
}

func (c *BundleComponent) bundleETagShouldNotMatch(bundleID, etag string) error {
	return c.bundleETagValueMatch(bundleID, etag, false)
}

func (c *BundleComponent) bundleETagValueMatch(bundleID, etag string, shouldMatch bool) error {
	bundle, err := c.getBundleByID(bundleID)

	if err != nil {
		return err
	}

	etagMatches := bundle.ETag == etag

	if etagMatches != shouldMatch {
		return fmt.Errorf("bundle %s has unexpected etag %s", bundleID, bundle.ETag)
	}

	return nil
}

func (c *BundleComponent) theseVersionsShouldHaveTheseStates(expectedVersionsJSON *godog.DocString) error {
	var expectedVersions []datasetAPIModels.Version
	if err := json.Unmarshal([]byte(expectedVersionsJSON.Content), &expectedVersions); err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	idSelector := func(version datasetAPIModels.Version) string { return version.ID }
	predicate := func(actual, expected datasetAPIModels.Version) string {
		if actual.State != expected.State {
			return fmt.Sprintf("version %s does not have expected state %s. actual state is %s", actual.ID, expected.ID, actual.State)
		}

		return ""
	}

	versions := make([]datasetAPIModels.Version, len(c.DatasetAPIVersions))
	for index := range versions {
		versions[index] = *c.DatasetAPIVersions[index]
	}

	errs := compareSlices(expectedVersions, versions, predicate, idSelector)

	if len(errs) > 0 {
		return fmt.Errorf("content items do not match:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

func (c *BundleComponent) theTotalNumberOfEventsShouldBe(expectedCount int) error {
	ctx := context.Background()
	collectionName := c.MongoClient.ActualCollectionName("BundleEventsCollection")

	count, err := c.MongoClient.Connection.Collection(collectionName).Count(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count events: %w", err)
	}

	if count != expectedCount {
		return fmt.Errorf("expected %d bundle events, but found %d", expectedCount, count)
	}
	return nil
}

func (c *BundleComponent) theNumberOfEventsWithActionAndDatatypeShouldBe(action, datatype string, expectedCount int) error {
	ctx := context.Background()
	collectionName := c.MongoClient.ActualCollectionName("BundleEventsCollection")

	filter := bson.M{
		"action": action,
		datatype: bson.M{"$exists": true},
	}

	count, err := c.MongoClient.Connection.Collection(collectionName).Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to count events: %w", err)
	}

	if count != expectedCount {
		return fmt.Errorf("expected %d events with action '%s' and type '%s', but found %d", expectedCount, action, datatype, count)
	}
	return nil
}

func (c *BundleComponent) iHaveThesePolicies(policiesJSON *godog.DocString) error {
	var policies []*permissionsAPIModels.Policy
	if err := json.Unmarshal([]byte(policiesJSON.Content), &policies); err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	c.permissionsAPIPolicies = append(c.permissionsAPIPolicies, policies...)

	return nil
}

func (c *BundleComponent) theFollowingPoliciesShouldExist(policiesJSON *godog.DocString) error {
	var expectedPolicies []*permissionsAPIModels.Policy
	if err := json.Unmarshal([]byte(policiesJSON.Content), &expectedPolicies); err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	for _, expectedPolicy := range expectedPolicies {
		policyFound := false
		for _, existingPolicy := range c.permissionsAPIPolicies {
			if existingPolicy.ID == expectedPolicy.ID {
				if diff := cmp.Diff(expectedPolicy, existingPolicy); diff != "" {
					return fmt.Errorf("policy with ID %s does not match expected:\n%s", expectedPolicy.ID, diff)
				}
				policyFound = true
				break
			}
		}
		if !policyFound {
			return fmt.Errorf("expected policy with ID %s to exist, but it was not found", expectedPolicy.ID)
		}
	}

	return nil
}

func (c *BundleComponent) thePolicyShouldHaveTheseConditionValues(policyID string, valuesJSON *godog.DocString) error {
	var expectedValues []string
	if err := json.Unmarshal([]byte(valuesJSON.Content), &expectedValues); err != nil {
		return fmt.Errorf("failed to unmarshal expected values: %w", err)
	}

	var foundPolicy *permissionsAPIModels.Policy
	for _, p := range c.permissionsAPIPolicies {
		if p.ID == policyID {
			foundPolicy = p
			break
		}
	}

	if foundPolicy == nil {
		return fmt.Errorf("policy with id %s not found", policyID)
	}

	actualValues := foundPolicy.Condition.Values

	if len(expectedValues) == 0 && len(actualValues) == 0 {
		return nil
	}

	if diff := cmp.Diff(expectedValues, actualValues); diff != "" {
		return fmt.Errorf("policy condition values do not match:\n%s", diff)
	}

	return nil
}

type Predicate[T any] = func(actual, expected T) string
type IDSelector[T any] = func(T) string

func compareSlices[T any](expectedItems, actualItems []T, matches Predicate[T], idSelector IDSelector[T]) []string {
	var errs []string

	actualItemsMap := make(map[string]*T)
	for i := range actualItems {
		id := idSelector(actualItems[i])
		actualItemsMap[id] = &actualItems[i]
	}

	for _, expected := range expectedItems {
		id := idSelector(expected)
		actual, exists := actualItemsMap[id]
		if !exists {
			errs = append(errs, fmt.Sprintf("failed to find item for id %s", id))
			continue
		}

		err := matches(*actual, expected)

		if err != "" {
			errs = append(errs, err)
		}
	}

	return errs
}
