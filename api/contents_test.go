package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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

func TestPostBundleContents_MalformedJSON_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a POST request to /bundles/{bundle-id}/contents with a malformed JSON body", t, func() {
		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: &storetest.StorerMock{}}, &datasetAPISDKMock.ClienterMock{}, false)

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
}

func TestPostBundleContents_InvalidBody_Failure(t *testing.T) {
	t.Parallel()

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
				Preview: "/preview",
			},
		}
		invalidContentItemJSON, err := json.Marshal(invalidContentItem)
		So(err, ShouldBeNil)

		r := httptest.NewRequest("POST", "/bundles/bundle-1/contents", bytes.NewReader(invalidContentItemJSON))
		w := httptest.NewRecorder()

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: &storetest.StorerMock{}}, &datasetAPISDKMock.ClienterMock{}, false)

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
						{
							Code:        &codeMissingParameters,
							Description: apierrors.ErrorDescriptionMissingParameters,
							Source:      &models.Source{Field: "/links/edit"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPostBundleContents_NonExistentBundle_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

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
}

func TestPostBundleContents_NonExistentDatasetEditionOrVersion_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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
}

func TestPostBundleContents_ExistingDatasetEditionAndVersion_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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
}

func TestPostBundleContents_CreateContentItem_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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
}

func TestPostBundleContents_ParseJWT_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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
}

func TestPostBundleContents_BundleEventCreation_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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
}

func TestPostBundleContents_UpdateBundleETag_Failure(t *testing.T) {
	t.Parallel()

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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = &mockDatasetAPIClient

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

func TestDeleteContentItem_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE request to /bundles/{bundle-id}/contents/{content-id}", t, func() {
		r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-1", http.NoBody)
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			GetContentItemByBundleIDAndContentItemIDFunc: func(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
				if bundleID == "bundle-1" && contentItemID == "content-1" {
					return &models.ContentItem{
						ID:       "content-1",
						BundleID: "bundle-1",
					}, nil
				}
				return nil, errors.New("content item not found")
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				if contentItemID == "content-1" {
					return nil
				}
				return errors.New("failed to delete content item")
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				if event.ContentItem.BundleID == "bundle-1" && event.ContentItem.ID == "content-1" {
					return nil
				}
				return errors.New("failed to create bundle event")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When deleteContentItem is called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should return a 204 No Content status", func() {
				So(w.Code, ShouldEqual, 204)
			})

			Convey("And the response body should be empty", func() {
				So(w.Body.Len(), ShouldEqual, 0)
			})
		})
	})
}

func TestDeleteContentItem_ContentItemNotFound_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE request to /bundles/{bundle-id}/contents/{content-id} for a non-existent content item", t, func() {
		mockedDatastore := &storetest.StorerMock{
			GetContentItemByBundleIDAndContentItemIDFunc: func(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
				if bundleID == "bundle-1" && contentItemID == "content-1" {
					return nil, apierrors.ErrContentItemNotFound
				}
				return nil, errors.New("unexpected error")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When deleteContentItem is called with a non-existent content item or bundle", func() {
			r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-1", http.NoBody)
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
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When deleteContentItem is called and the GetContentItemByBundleIDAndContentItemID fails", func() {
			r := httptest.NewRequest("DELETE", "/bundles/bundle-0/contents/content-0", http.NoBody)
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
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestDeleteContentItem_ContentItemIsPublished_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE request to /bundles/{bundle-id}/contents/{content-id} for a published content item", t, func() {
		r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-1", http.NoBody)
		w := httptest.NewRecorder()

		publishedState := models.StatePublished

		mockedDatastore := &storetest.StorerMock{
			GetContentItemByBundleIDAndContentItemIDFunc: func(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
				if bundleID == "bundle-1" && contentItemID == "content-1" {
					return &models.ContentItem{
						ID:       "content-1",
						BundleID: "bundle-1",
						State:    &publishedState,
					}, nil
				}
				return nil, apierrors.ErrContentItemNotFound
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When deleteContentItem is called for a published content item", func() {
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
							Description: apierrors.ErrorDescriptionContentItemAlreadyPublished,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestDeleteContentItem_DeleteContentItem_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE request to /bundles/{bundle-id}/contents/{content-id}", t, func() {
		mockedDatastore := &storetest.StorerMock{
			GetContentItemByBundleIDAndContentItemIDFunc: func(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
				if bundleID == "bundle-1" && contentItemID == "content-1" {
					return &models.ContentItem{
						ID:       "content-1",
						BundleID: "bundle-1",
					}, nil
				}
				if bundleID == "bundle-1" && contentItemID == "content-2" {
					return &models.ContentItem{
						ID:       "content-2",
						BundleID: "bundle-1",
					}, nil
				}
				return nil, apierrors.ErrContentItemNotFound
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				if contentItemID == "content-1" {
					return apierrors.ErrContentItemNotFound
				}
				if contentItemID == "content-2" {
					return errors.New("failed to delete content item")
				}
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When deleteContentItem is called and the content item is not found", func() {
			r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-1", http.NoBody)
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
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When deleteContentItem is called and the DeleteContentItem fails", func() {
			r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-2", http.NoBody)
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
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestDeleteContentItem_ParseJWT_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE request to /bundles/{bundle-id}/contents/{content-id} and parsing the JWT fails", t, func() {
		r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-1", http.NoBody)
		r.Header.Set("Authorization", "invalid-auth-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			GetContentItemByBundleIDAndContentItemIDFunc: func(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
				if bundleID == "bundle-1" && contentItemID == "content-1" {
					return &models.ContentItem{
						ID:       "content-1",
						BundleID: "bundle-1",
					}, nil
				}
				return nil, errors.New("content item not found")
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				if contentItemID == "content-1" {
					return nil
				}
				return errors.New("failed to delete content item")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When deleteContentItem is called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			codeInternalServerError := models.CodeInternalServerError
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &codeInternalServerError,
						Description: apierrors.ErrorDescriptionInternalError,
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})
	})
}

func TestDeleteContentItem_CreateBundleEvent_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE request to /bundles/{bundle-id}/contents/{content-id} and bundle event creation fails", t, func() {
		r := httptest.NewRequest("DELETE", "/bundles/bundle-1/contents/content-1", http.NoBody)
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			GetContentItemByBundleIDAndContentItemIDFunc: func(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
				if bundleID == "bundle-1" && contentItemID == "content-1" {
					return &models.ContentItem{
						ID:       "content-1",
						BundleID: "bundle-1",
					}, nil
				}
				return nil, apierrors.ErrContentItemNotFound
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				if contentItemID == "content-1" {
					return nil
				}
				return errors.New("failed to delete content item")
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create bundle event")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When deleteContentItem is called", func() {
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
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestGetBundleContents_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a GET /bundles/{bundle-id}/contents request with valid bundle and default pagination", t, func() {
		bundleID := "bundle-1"
		statePublished := models.State(models.BundleStatePublished.String())

		expectedContents := []*models.ContentItem{
			{
				ID:          "1",
				BundleID:    bundleID,
				ContentType: models.ContentType("dataset"),
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "2023-01",
					Title:     "Dataset One Title",
					VersionID: 1,
				},
				State: &statePublished,
				Links: models.Links{},
			},
			{
				ID:          "2",
				BundleID:    bundleID,
				ContentType: models.ContentType("dataset"),
				Metadata: models.Metadata{
					DatasetID: "dataset-2",
					EditionID: "2023-02",
					Title:     "Dataset Two Title",
					VersionID: 2,
				},
				State: &statePublished,
				Links: models.Links{},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, id string) (bool, error) {
				return id == bundleID, nil
			},
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				state := models.BundleStatePublished
				return &models.Bundle{
					ID:    bundleID,
					ETag:  "dummy-etag",
					State: state,
				}, nil
			},
			ListBundleContentsFunc: func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return expectedContents, len(expectedContents), nil
			},
		}

		mockAPIClient := &datasetAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockAPIClient, false)

		r := httptest.NewRequest("GET", "/bundles/"+bundleID+"/contents", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		Convey("When the handler is called with no pagination params", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should respond 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And response contains expected contents with default pagination", func() {
				var resp struct {
					Items      []*models.ContentItem `json:"items"`
					TotalCount int                   `json:"total_count"`
				}
				err := json.NewDecoder(w.Body).Decode(&resp)
				So(err, ShouldBeNil)
				So(resp.Items, ShouldResemble, expectedContents)
				So(resp.TotalCount, ShouldEqual, len(expectedContents))
			})

			Convey("And response headers include ETag and Cache-Control", func() {
				So(w.Header().Get("ETag"), ShouldNotBeEmpty)
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
			})
		})
	})

	Convey("Given a GET request with custom pagination params", t, func() {
		bundleID := "bundle-1"
		statePublished := models.State(models.BundleStatePublished.String())

		expectedContents := []*models.ContentItem{
			{
				ID:          "1",
				BundleID:    bundleID,
				ContentType: models.ContentType("dataset"),
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "2023-01",
					Title:     "Dataset One Title",
					VersionID: 1,
				},
				State: &statePublished,
				Links: models.Links{},
			},
			{
				ID:          "2",
				BundleID:    bundleID,
				ContentType: models.ContentType("dataset"),
				Metadata: models.Metadata{
					DatasetID: "dataset-2",
					EditionID: "2023-02",
					Title:     "Dataset Two Title",
					VersionID: 2,
				},
				State: &statePublished,
				Links: models.Links{},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, id string) (bool, error) {
				return id == bundleID, nil
			},
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				state := models.BundleStatePublished
				return &models.Bundle{
					ID:    bundleID,
					ETag:  "dummy-etag",
					State: state,
				}, nil
			},
			ListBundleContentsFunc: func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return expectedContents, len(expectedContents), nil
			},
		}

		mockAPIClient := &datasetAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockAPIClient, false)

		r := httptest.NewRequest("GET", "/bundles/"+bundleID+"/contents"+"?offset=1&limit=1", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		Convey("When the handler is called with custom pagination params", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should respond 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And response contains expected contents with default pagination", func() {
				var resp struct {
					Items      []*models.ContentItem `json:"items"`
					TotalCount int                   `json:"total_count"`
				}
				err := json.NewDecoder(w.Body).Decode(&resp)
				So(err, ShouldBeNil)
				So(resp.Items, ShouldResemble, expectedContents)
				So(resp.TotalCount, ShouldEqual, len(expectedContents))
			})

			Convey("And response headers include ETag and Cache-Control", func() {
				So(w.Header().Get("ETag"), ShouldNotBeEmpty)
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
			})
		})
	})

	Convey("Given a bundle with state NOT published, enrichment is applied", t, func() {
		bundleID := "bundle-2"

		originalContents := []*models.ContentItem{
			{
				ID:          "10",
				BundleID:    bundleID,
				ContentType: models.ContentType("dataset"),
				Metadata: models.Metadata{
					DatasetID: "dataset-10",
					EditionID: "2023-10",
					Title:     "",
					VersionID: 1,
				},
				State: nil,
				Links: models.Links{},
			},
		}

		r := httptest.NewRequest("GET", "/bundles/"+bundleID+"/contents", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				state := models.BundleStateDraft
				return &models.Bundle{
					ID:    bundleID,
					ETag:  "dummy-etag",
					State: state,
				}, nil
			},
			ListBundleContentsFunc: func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return originalContents, len(originalContents), nil
			},
			CheckBundleExistsFunc: func(ctx context.Context, id string) (bool, error) {
				return id == bundleID, nil
			},
		}

		mockDatasetAPIClient := datasetAPISDKMock.ClienterMock{
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, collectionID string, DatasetID string) (datasetAPIModels.Dataset, error) {
				dataset := datasetAPIModels.Dataset{
					ID:    "dataset-10",
					Title: "Test Title",
					State: "DRAFT",
				}
				return dataset, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &mockDatasetAPIClient, false)

		Convey("When the handler is called for unpublished bundle", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then it should respond 200 OK and enrich content items", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				var resp struct {
					Items      []*models.ContentItem `json:"items"`
					TotalCount int                   `json:"total_count"`
				}
				err := json.NewDecoder(w.Body).Decode(&resp)
				So(err, ShouldBeNil)
				So(len(resp.Items), ShouldEqual, 1)
				So(resp.Items[0].Metadata.Title, ShouldEqual, "Test Title")
				So(resp.Items[0].State.String(), ShouldEqual, "DRAFT")
			})
		})
	})
}

func TestGetBundleContents_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given invalid pagination params", t, func() {
		mockAPIClient := &datasetAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{}, mockAPIClient, false)
		r := httptest.NewRequest("GET", "/bundles/bundle-1/contents?offset=-1", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		Convey("When called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeBadRequest := models.CodeBadRequest
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeBadRequest,
							Description: "Unable to process request due to a malformed or invalid request body or query parameter",
							Source: &models.Source{
								Parameter: " offset",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a nonexistent bundle", t, func() {
		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, id string) (bool, error) {
				return false, nil
			},
		}
		mockAPIClient := &datasetAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockAPIClient, false)
		r := httptest.NewRequest("GET", "/bundles/non-existent-bundle-id/contents", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		Convey("When called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusNotFound)

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

	Convey("Given internal error from datastore", t, func() {
		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, id string) (bool, error) {
				return true, nil
			},
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return nil, errors.New("datastore failure")
			},
		}
		mockAPIClient := &datasetAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockAPIClient, false)
		r := httptest.NewRequest("GET", "/bundles/bundle-1/contents", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		Convey("When called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalServerError := models.CodeInternalServerError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInternalServerError,
							Description: "Failed to get dataset from dataset API",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})

	Convey("Given a bundle whose datasetID does not exist in the dataset API", t, func() {
		bundleID := "bundle-1"
		originalContents := []*models.ContentItem{
			{
				ID:          "10",
				BundleID:    bundleID,
				ContentType: models.ContentType("dataset"),
				Metadata: models.Metadata{
					DatasetID: "dataset-10",
					EditionID: "2023-10",
					Title:     "",
					VersionID: 1,
				},
				State: nil,
				Links: models.Links{},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, id string) (bool, error) {
				return true, nil
			},
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				state := models.BundleStateDraft
				return &models.Bundle{
					ID:    bundleID,
					ETag:  "dummy-etag",
					State: state,
				}, nil
			},
			ListBundleContentsFunc: func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return originalContents, len(originalContents), nil
			},
		}

		// Mock dataset API client to return not found error
		mockedDatasetAPI := &datasetAPISDKMock.ClienterMock{
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, collectionID string, DatasetID string) (datasetAPIModels.Dataset, error) {
				dataset := datasetAPIModels.Dataset{}
				return dataset, errors.New("Dataset not found")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockedDatasetAPI, false)
		r := httptest.NewRequest("GET", "/bundles/bundle-1/contents", http.NoBody)
		r.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		Convey("When called", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusNotFound)

			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeNotFound := models.CodeNotFound
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeNotFound,
							Description: "The requested resource does not exist",
							Source: &models.Source{
								Field: "/metadata/dataset_id",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}
