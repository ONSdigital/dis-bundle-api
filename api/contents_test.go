package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPostBundleContents_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a POST request to /bundles/{bundle-id}/contents with a valid JSON body", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				if bundleID == "bundle-1" {
					return true, nil
				}
				return false, nil
			},
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				if datasetID == "dataset-1" && editionID == "edition-1" && versionID == 1 {
					return false, nil
				}
				return true, nil
			},
			CreateContentItemFunc: func(ctx context.Context, contentItem *models.ContentItem) error {
				if contentItem.BundleID == "bundle-1" &&
					contentItem.Metadata.DatasetID == "dataset-1" &&
					contentItem.Metadata.EditionID == "edition-1" &&
					contentItem.Metadata.VersionID == 1 {
					return nil
				}
				return errors.New("failed to create content item")
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				if event.ContentItem.BundleID == "bundle-1" {
					return nil
				}
				return errors.New("failed to create bundle event")
			},
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				if bundleID == "bundle-1" {
					return &models.Bundle{
						ID:   bundleID,
						ETag: "new-etag",
					}, nil
				}
				return nil, errors.New("failed to update bundle ETag")
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				if datasetID == "dataset-1" && editionID == "edition-1" && versionID == "1" {
					return datasetAPIModels.Version{
						DatasetID: datasetID,
						Edition:   editionID,
						Version:   1,
					}, nil
				}
				return datasetAPIModels.Version{}, errors.New("version not found")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		Convey("When postBundleContents is called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			var createdContentItem models.ContentItem
			err := json.NewDecoder(w.Body).Decode(&createdContentItem)
			So(err, ShouldBeNil)

			Convey("Then it should return a 201 Created status and the created content item", func() {
				So(w.Code, ShouldEqual, 201)

				So(createdContentItem.ID, ShouldNotBeEmpty)
				So(createdContentItem.BundleID, ShouldEqual, "bundle-1")
				So(createdContentItem.ContentType.String(), ShouldEqual, "DATASET")
				So(createdContentItem.Metadata.DatasetID, ShouldEqual, "dataset-1")
				So(createdContentItem.Metadata.EditionID, ShouldEqual, "edition-1")
				So(createdContentItem.Metadata.VersionID, ShouldEqual, 1)
				So(createdContentItem.Links.Edit, ShouldEqual, "/edit")
				So(createdContentItem.Links.Preview, ShouldEqual, "/preview")
			})

			Convey("And the correct headers should be set", func() {
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Header().Get("ETag"), ShouldEqual, "new-etag")
				So(w.Header().Get("Location"), ShouldEqual, "/bundles/bundle-1/contents/"+createdContentItem.ID)
			})
		})
	})
}

func TestPostBundleContents_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a POST request to /bundles/{bundle-id}/contents with a malformed JSON body", t, func() {
		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: &storetest.StorerMock{}})

		Convey("When postBundleContents is called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 400 Bad Request status code", func() {
				So(w.Code, ShouldEqual, 400)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeBadRequest
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionMalformedRequest,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents with an invalid body", t, func() {
		invalidContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		invalidContentItemJSON, err := json.Marshal(invalidContentItem)
		So(err, ShouldBeNil)

		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(invalidContentItemJSON))
		w := httptest.NewRecorder()

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: &storetest.StorerMock{}})

		Convey("When postBundleContents is called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 400 Bad Request status code", func() {
				So(w.Code, ShouldEqual, 400)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeMissingParameters := models.CodeMissingParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeMissingParameters,
							Description: apierrors.ErrorDescriptionMissingParameters,
							Source:      &models.Source{Field: "/metadata/edition_id"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents with a non-existent bundle", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				if bundleID == "non-existent-bundle" {
					return false, nil
				}
				return false, errors.New("unexpected error")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

		Convey("When postBundleContents is called and CheckBundleExists fails", func() {
			r := httptest.NewRequest("POST", "/bundles/database-failure-bundle/contents", bytes.NewReader(newContentItemJSON))
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 500 Internal Server Error status code", func() {
				So(w.Code, ShouldEqual, 500)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalError := models.CodeInternalServerError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInternalError,
							Description: "Failed to check if bundle exists",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When postBundleContents is called with a non-existent bundle", func() {
			r := httptest.NewRequest("POST", "/bundles/non-existent-bundle/contents", bytes.NewReader(newContentItemJSON))
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 404 Not Found status code", func() {
				So(w.Code, ShouldEqual, 404)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeNotFound := models.CodeNotFound
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeNotFound,
							Description: "Bundle not found",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents with a dataset, edition or version that does not exist", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				return true, nil
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				if datasetID == "dataset-1" {
					return datasetAPIModels.Version{}, errors.New("dataset not found")
				}
				if editionID == "edition-1" {
					return datasetAPIModels.Version{}, errors.New("edition not found")
				}
				if versionID == "1" {
					return datasetAPIModels.Version{}, errors.New("version not found")
				}
				return datasetAPIModels.Version{}, errors.New("unexpected error")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		Convey("When postBundleContents is called with a non-existent dataset", func() {
			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("X-Florence-Token", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 404 Not Found status code", func() {
				So(w.Code, ShouldEqual, 404)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeNotFound := models.CodeNotFound
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeNotFound,
							Description: apierrors.ErrorDescriptionNotFound,
							Source:      &models.Source{Field: "/metadata/dataset_id"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When postBundleContents is called with a non-existent edition", func() {
			newContentItem.Metadata.DatasetID = "dataset-2"
			newContentItemJSON, err := json.Marshal(newContentItem)
			So(err, ShouldBeNil)

			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("X-Florence-Token", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 404 Not Found status code", func() {
				So(w.Code, ShouldEqual, 404)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeNotFound := models.CodeNotFound
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeNotFound,
							Description: apierrors.ErrorDescriptionNotFound,
							Source:      &models.Source{Field: "/metadata/edition_id"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When postBundleContents is called with a non-existent version", func() {
			newContentItem.Metadata.DatasetID = "dataset-2"
			newContentItem.Metadata.EditionID = "edition-2"
			newContentItemJSON, err := json.Marshal(newContentItem)
			So(err, ShouldBeNil)

			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("X-Florence-Token", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 404 Not Found status code", func() {
				So(w.Code, ShouldEqual, 404)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeNotFound := models.CodeNotFound
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeNotFound,
							Description: apierrors.ErrorDescriptionNotFound,
							Source:      &models.Source{Field: "/metadata/version_id"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When postBundleContents is called and getVersion fails", func() {
			newContentItem.Metadata.DatasetID = "dataset-2"
			newContentItem.Metadata.EditionID = "edition-2"
			newContentItem.Metadata.VersionID = 2
			newContentItemJSON, err := json.Marshal(newContentItem)
			So(err, ShouldBeNil)

			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("X-Florence-Token", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, 500)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalServerError := models.CodeInternalServerError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInternalServerError,
							Description: "Failed to get version from dataset API",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents with a content item that already contains the same dataset, edition, and version", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				return true, nil
			},
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				if datasetID == "dataset-1" && editionID == "edition-1" && versionID == 1 {
					return true, nil
				}
				return false, errors.New("unexpected error")
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				return datasetAPIModels.Version{}, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		Convey("When postBundleContents is called with an existing content item", func() {
			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("Authorization", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 409 Conflict status code", func() {
				So(w.Code, ShouldEqual, 409)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeConflict := models.CodeConflict
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeConflict,
							Description: "Content item already exists for the given dataset, edition and version",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When postBundleContents is called and CheckContentItemExistsByDatasetEditionVersion fails", func() {
			newContentItem.Metadata.VersionID = 2
			newContentItemJSON, err := json.Marshal(newContentItem)
			So(err, ShouldBeNil)

			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("Authorization", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 500 Internal Server Error status code", func() {
				So(w.Code, ShouldEqual, 500)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalServerError := models.CodeInternalServerError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInternalServerError,
							Description: "Failed to check if content item exists",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents and the datastore fails to create the content item", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				return true, nil
			},
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				return false, nil
			},
			CreateContentItemFunc: func(ctx context.Context, contentItem *models.ContentItem) error {
				return errors.New("failed to create content item")
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				return datasetAPIModels.Version{}, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		Convey("When postBundleContents is called", func() {
			r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
			r.Header.Set("Authorization", "test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 500 Internal Server Error status code", func() {
				So(w.Code, ShouldEqual, 500)
			})

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalServerError := models.CodeInternalServerError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInternalServerError,
							Description: "Failed to create content item in the datastore",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents and parsing the JWT fails", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				return true, nil
			},
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				return false, nil
			},
			CreateContentItemFunc: func(ctx context.Context, contentItem *models.ContentItem) error {
				return nil
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				return datasetAPIModels.Version{}, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
		r.Header.Set("Authorization", "invalid-auth-token")
		w := httptest.NewRecorder()

		Convey("When postBundleContents is called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			codeInternalServerError := models.CodeInternalServerError
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &codeInternalServerError,
						Description: "Failed to get user identity from JWT",
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents and bundle event creation fails", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				return true, nil
			},
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				return false, nil
			},
			CreateContentItemFunc: func(ctx context.Context, contentItem *models.ContentItem) error {
				return nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create bundle event")
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				return datasetAPIModels.Version{}, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		bundleAPI.Router.ServeHTTP(w, r)

		Convey("Then it should return a 500 Internal Server Error status code", func() {
			So(w.Code, ShouldEqual, 500)
		})

		Convey("And the response body should contain an error message", func() {
			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			codeInternalServerError := models.CodeInternalServerError
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &codeInternalServerError,
						Description: "Failed to create event",
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})
	})

	Convey("Given a POST request to /bundles/{bundle-id}/contents and updating the bundle ETag fails", t, func() {
		newContentItem := &models.ContentItem{
			BundleID:    "bundle-1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset-1",
				EditionID: "edition-1",
				Title:     "Example Content Item",
				VersionID: 1,
			},
			Links: models.Links{
				Edit:    "/edit",
				Preview: "/preview",
			},
		}
		newContentItemJSON, err := json.Marshal(newContentItem)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				return true, nil
			},
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				return false, nil
			},
			CreateContentItemFunc: func(ctx context.Context, contentItem *models.ContentItem) error {
				return nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				return nil, errors.New("failed to update bundle ETag")
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
				return datasetAPIModels.Version{}, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
		bundleAPI.datasetAPIClient = &mockDatasetAPIClient

		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(newContentItemJSON))
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		bundleAPI.Router.ServeHTTP(w, r)

		Convey("Then it should return a 500 Internal Server Error status code", func() {
			So(w.Code, ShouldEqual, 500)
		})

		Convey("And the response body should contain an error message", func() {
			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			codeInternalServerError := models.CodeInternalServerError
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &codeInternalServerError,
						Description: "Failed to update bundle ETag",
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})
	})
}
