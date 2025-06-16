package datasets

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"

	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	mockDatasetID1        = "dataset1"
	mockDatasetEditionID1 = "edition1"
	mockDatasetVersion1   = 1
	mockDatasetState1     = "draft"
)

var mockVersion1 = datasetAPIModels.Version{
	ID:        "dataset-1-version-1",
	Version:   mockDatasetVersion1,
	DatasetID: mockDatasetID1,
	Edition:   mockDatasetEditionID1,
	State:     mockDatasetState1,
}

var mockVersion2 = datasetAPIModels.Version{

	ID:        "dataset-1-version-2",
	Version:   2,
	DatasetID: "dataset1",
	Edition:   "edition1",
}

var mockVersion3 = datasetAPIModels.Version{
	ID:        "dataset-2-version-1",
	Version:   1,
	DatasetID: "dataset2",
	Edition:   "edition2",
}

func createVersions() []*datasetAPIModels.Version {
	return []*datasetAPIModels.Version{
		&mockVersion1,
		&mockVersion2,
		&mockVersion3,
	}
}

const (
	errNotFound = "not found"
)

func createMockDatasetAPIClient(versions []*datasetAPIModels.Version) datasetAPISDK.Clienter {
	getVersion := func(datasetID, editionID, versionID string) (*datasetAPIModels.Version, error) {
		for _, version := range versions {
			versionIDint, err := strconv.Atoi(versionID)
			if err != nil {
				return nil, errors.New("could not parse versionid to int")
			}

			if version.DatasetID == datasetID && version.Edition == editionID && version.Version == versionIDint {
				return version, nil
			}
		}

		return nil, errors.New(errNotFound)
	}

	getVersionFunc := func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
		version, err := getVersion(datasetID, editionID, versionID)
		if err != nil {
			return datasetAPIModels.Version{}, err
		}
		return *version, nil
	}

	return &datasetAPISDKMock.ClienterMock{
		GetVersionFunc: getVersionFunc,
		PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
			version, err := getVersion(datasetID, editionID, versionID)

			if err != nil {
				return err
			}

			version.State = state
			return nil
		},
	}
}

func TestGetForContentItem(t *testing.T) {
	versions := createVersions()
	mockClient := createMockDatasetAPIClient(versions)

	versionsClient := createVersionsClient(mockClient)

	ctx := context.Background()
	httpRequest := http.Request{
		Header: http.Header{},
	}

	Convey("When GetForContentItem is called with an existing version", t, func() {
		contentItem := models.ContentItem{Metadata: models.Metadata{
			DatasetID: mockDatasetID1,
			VersionID: mockDatasetVersion1,
			EditionID: mockDatasetEditionID1,
		}}
		result, err := versionsClient.GetForContentItem(ctx, &httpRequest, &contentItem)

		Convey("Should return success", func() {
			So(err, ShouldBeNil)

			So(*result, ShouldEqual, mockVersion1)
		})
	})

	Convey("When GetForContentItem is called with a missing version", t, func() {
		contentItem := models.ContentItem{Metadata: models.Metadata{
			DatasetID: "doesnt-exist",
			VersionID: 999,
			EditionID: "doesnt-exist",
		}}

		result, err := versionsClient.GetForContentItem(ctx, &httpRequest, &contentItem)

		Convey("Should return error", func() {
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errNotFound)
		})
	})
}

func TestUpdateStateForContentItem(t *testing.T) {
	versions := createVersions()
	mockClient := createMockDatasetAPIClient(versions)

	versionsClient := createVersionsClient(mockClient)

	ctx := context.Background()
	httpRequest := http.Request{
		Header: http.Header{},
	}

	Convey("When UpdateStateForContentItem is called with an existing version", t, func() {
		contentItem := models.ContentItem{Metadata: models.Metadata{
			DatasetID: mockDatasetID1,
			VersionID: mockDatasetVersion1,
			EditionID: mockDatasetEditionID1,
		}}

		state := models.BundleStateApproved
		err := versionsClient.UpdateStateForContentItem(ctx, &httpRequest, &contentItem, state)

		Convey("Should return success", func() {
			So(err, ShouldBeNil)

			So(mockVersion1.State, ShouldEqual, strings.ToLower(state.String()))
		})
	})

	Convey("When UpdateStateForContentItem is called with a not found version", t, func() {
		contentItem := models.ContentItem{Metadata: models.Metadata{
			DatasetID: "doesnt-exist",
			VersionID: 999,
			EditionID: "doesnt-exist",
		}}

		state := models.BundleStateApproved
		err := versionsClient.UpdateStateForContentItem(ctx, &httpRequest, &contentItem, state)

		Convey("Should return err", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errNotFound)
		})
	})
}
