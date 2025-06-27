package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	bundle1  = "bundle-1"
	dataset1 = "dataset-1"
	title1   = "title1"
	newEtag  = "new-etag"
	content1 = "content1"
)

func TestPutBundle_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a valid PUT request to /bundles/{bundle-id}", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:           bundle1,
			Title:        "Original Title",
			BundleType:   models.BundleTypeManual,
			ETag:         "original-etag",
			State:        models.BundleStateDraft,
			CreatedAt:    &now,
			CreatedBy:    &models.User{Email: "creator@example.com"},
			UpdatedAt:    &now,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				if id == bundle1 {
					return existingBundle, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				if title == title1 && excludeID == bundle1 {
					return false, nil
				}
				return false, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				if bundleID == bundle1 {
					bundle.ID = bundleID
					return bundle, nil
				}
				return nil, errors.New("failed to update bundle")
			},
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				if bundleID == bundle1 {
					updatedBundle := *existingBundle
					updatedBundle.Title = title1
					updatedBundle.ETag = newEtag
					return &updatedBundle, nil
				}
				return nil, errors.New("failed to update ETag")
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				if event.Action == models.ActionUpdate && event.Resource == "/bundles/bundle-1" {
					return nil
				}
				return errors.New("failed to create event")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with valid data", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK with updated bundle", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
				So(w.Header().Get("ETag"), ShouldEqual, newEtag)

				var response models.Bundle
				err := json.NewDecoder(w.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(response.ID, ShouldEqual, bundle1)
				So(response.Title, ShouldEqual, title1)
			})
		})
	})
}

func TestPutBundle_StateTransition_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with valid state transition", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:           bundle1,
			Title:        "Test Bundle",
			BundleType:   models.BundleTypeManual,
			ETag:         "original-etag",
			State:        models.BundleStateDraft,
			CreatedAt:    &now,
			CreatedBy:    &models.User{Email: "creator@example.com"},
			UpdatedAt:    &now,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        "Test Bundle",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateInReview,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				return false, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				updatedBundle := *existingBundle
				updatedBundle.State = models.BundleStateInReview
				updatedBundle.ETag = newEtag
				return &updatedBundle, nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with valid state transition", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK with updated state", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				var response models.Bundle
				err := json.NewDecoder(w.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(response.State, ShouldEqual, models.BundleStateInReview)
			})
		})
	})
}

func TestPutBundle_StateTransitionToPublished_UpdatesContentItems(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request transitioning to PUBLISHED state", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:           bundle1,
			Title:        "Test Bundle",
			BundleType:   models.BundleTypeManual,
			ETag:         "original-etag",
			State:        models.BundleStateApproved,
			CreatedAt:    &now,
			CreatedBy:    &models.User{Email: "creator@example.com"},
			UpdatedAt:    &now,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        "Test Bundle",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStatePublished,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		contentItems := []*models.ContentItem{
			{
				ID:       content1,
				BundleID: bundle1,
				Metadata: models.Metadata{
					DatasetID: dataset1,
					Title:     "Old Title",
				},
				State: ptrState(models.StateApproved),
			},
		}

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				return false, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				updatedBundle := *existingBundle
				updatedBundle.State = models.BundleStatePublished
				updatedBundle.ETag = newEtag
				return &updatedBundle, nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				if bundleID == bundle1 {
					return contentItems, nil
				}
				return nil, errors.New("bundle not found")
			},
			UpdateContentItemDatasetInfoFunc: func(ctx context.Context, contentItemID, title, state string) error {
				if contentItemID == content1 && title == "Dataset Title" && state == "published" {
					return nil
				}
				return errors.New("failed to update content item")
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
				if datasetID == dataset1 {
					return datasetAPIModels.Dataset{
						ID:    datasetID,
						Title: "Dataset Title",
						State: "published",
					}, nil
				}
				return datasetAPIModels.Dataset{}, errors.New("dataset not found")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)
		bundleAPI.stateMachineBundleAPI.DatasetAPIClient = mockDatasetAPIClient

		Convey("When putBundle is called with transition to PUBLISHED", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK and update content items", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				var response models.Bundle
				err := json.NewDecoder(w.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(response.State, ShouldEqual, models.BundleStatePublished)

				So(len(mockedDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
				So(mockedDatastore.GetContentItemsByBundleIDCalls()[0].BundleID, ShouldEqual, bundle1)

				So(len(mockedDatastore.UpdateContentItemDatasetInfoCalls()), ShouldEqual, 1)
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].ContentItemID, ShouldEqual, content1)
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].Title, ShouldEqual, "Dataset Title")
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].State, ShouldEqual, "published")
			})
		})
	})
}

func TestPutBundle_MissingIfMatchHeader_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request without If-Match header", t, func() {
		updateRequest := &models.Bundle{
			Title:      "title-2",
			BundleType: models.BundleTypeManual,
			State:      models.BundleStateDraft,
			ManagedBy:  models.ManagedByDataAdmin,
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: &storetest.StorerMock{}}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 409 Conflict", func() {
				So(w.Code, ShouldEqual, http.StatusConflict)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeConflict := models.CodeConflict
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeConflict,
							Description: "Change rejected due to a conflict with the current resource state. A common cause is attempted to change a bundle that is already locked pending publication or has already been published.",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPutBundle_ETagMismatch_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with mismatched ETag", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			Title:     "Original Title",
			ETag:      "correct-etag",
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		updateRequest := &models.Bundle{
			Title:      "title-3",
			BundleType: models.BundleTypeManual,
			State:      models.BundleStateDraft,
			ManagedBy:  models.ManagedByDataAdmin,
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with wrong ETag", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "wrong-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 409 Conflict", func() {
				So(w.Code, ShouldEqual, http.StatusConflict)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeConflict := models.CodeConflict
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeConflict,
							Description: "Change rejected due to a conflict with the current resource state. A common cause is attempted to change a bundle that is already locked pending publication or has already been published.",
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPutBundle_MalformedJSON_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with malformed JSON", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			ETag:      "original-etag",
			State:     models.BundleStateDraft,
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with invalid JSON", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader([]byte("invalid json")))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeBadRequest := models.ErrInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeBadRequest,
							Description: apierrors.ErrorDescriptionMalformedRequest,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPutBundle_DuplicateTitle_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with duplicate title", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			Title:     "Original Title",
			ETag:      "original-etag",
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		updateRequest := &models.Bundle{
			Title:        "Existing Title",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				if title == "Existing Title" {
					return true, nil
				}
				return false, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with duplicate title", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeBadRequest := models.ErrInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeBadRequest,
							Description: apierrors.ErrorDescriptionMalformedRequest,
							Source:      &models.Source{Field: "/title"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPutBundle_InvalidStateTransition_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with invalid state transition", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			Title:     "Test Bundle",
			ETag:      "original-etag",
			State:     models.BundleStateDraft,
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		updateRequest := &models.Bundle{
			Title:        "Test Bundle",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStatePublished,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with invalid state transition", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeBadRequest := models.ErrInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeBadRequest,
							Description: apierrors.ErrorDescriptionMalformedRequest,
							Source:      &models.Source{Field: "/state"},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPutBundle_DatabaseErrors_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request that causes database errors", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			Title:     "Original Title",
			ETag:      "original-etag",
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		updateRequest := &models.Bundle{
			Title:        "Updated Title",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		Convey("When GetBundle fails with internal error", func() {
			mockedDatastore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return nil, errors.New("database connection failed")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			codeInternalError := models.InternalError
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &codeInternalError,
						Description: apierrors.ErrorDescriptionInternalError,
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})

		Convey("When UpdateBundle fails", func() {
			mockedDatastore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return existingBundle, nil
				},
				CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
					return false, nil
				},
				UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
					return nil, errors.New("database update failed")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When UpdateBundleETag fails", func() {
			mockedDatastore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return existingBundle, nil
				},
				CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
					return false, nil
				},
				UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
					return bundle, nil
				},
				UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
					return nil, errors.New("failed to update ETag")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When CreateBundleEvent fails", func() {
			mockedDatastore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return existingBundle, nil
				},
				CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
					return false, nil
				},
				UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
					return bundle, nil
				},
				UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
					updatedBundle := *existingBundle
					updatedBundle.ETag = newEtag
					return &updatedBundle, nil
				},
				CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
					return errors.New("failed to create event")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestPutBundle_AuthenticationFailure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with invalid authentication", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			Title:     "Original Title",
			ETag:      "original-etag",
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		updateRequest := &models.Bundle{
			Title:        "Updated Title",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: []models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return existingBundle, nil
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				return false, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called with invalid JWT", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer invalid-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalError := models.InternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInternalError,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func ptrState(state models.State) *models.State {
	return &state
}
