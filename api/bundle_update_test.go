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
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	bundle1   = "bundle-1"
	title1    = "title1"
	newEtag   = "new-etag"
	urlString = "/bundles/bundle-1"
	team1     = "team-1"
	team2     = "team-2"
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
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
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
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				if event.Action == models.ActionUpdate && event.Resource == urlString {
					return nil
				}
				return errors.New("failed to create event")
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				if id == team1 {
					return &permissionsAPIModels.Policy{}, nil
				}
				return nil, errors.New("404 policy not found")
			},
			PostPolicyWithIDFunc: func(ctx context.Context, id string, policy permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return nil, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, mockPermissionsClient, false)

		Convey("When putBundle is called with existing preview teams", func() {
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

			Convey("And no policies should have been created", func() {
				So(len(mockPermissionsClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsClient.PostPolicyWithIDCalls()), ShouldEqual, 0)
			})
		})

		Convey("When putBundle is called with new preview teams and an existing preview team", func() {
			bundleRequest := updateRequest
			bundleRequest.PreviewTeams = &[]models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}, {ID: "team-3"}}
			bundleRequestJSON, err := json.Marshal(bundleRequest)
			So(err, ShouldBeNil)

			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(bundleRequestJSON))
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

			Convey("And new policies should have been created for the new preview teams", func() {
				So(len(mockPermissionsClient.GetPolicyCalls()), ShouldEqual, 3)
				So(len(mockPermissionsClient.PostPolicyWithIDCalls()), ShouldEqual, 2)
			})
		})
	})
}

func TestPutBundleNoPreviewTeam_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a valid PUT request to /bundles/{bundle-id}", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:         bundle1,
			Title:      "Original Title",
			BundleType: models.BundleTypeManual,
			ETag:       "original-etag",
			State:      models.BundleStateDraft,
			CreatedAt:  &now,
			CreatedBy:  &models.User{Email: "creator@example.com"},
			UpdatedAt:  &now,
			ManagedBy:  models.ManagedByDataAdmin,
		}

		updateRequest := &models.Bundle{
			Title:      title1,
			BundleType: models.BundleTypeManual,
			State:      models.BundleStateDraft,
			ManagedBy:  models.ManagedByDataAdmin,
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
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				if event.Action == models.ActionUpdate && event.Resource == urlString {
					return nil
				}
				return errors.New("failed to create event")
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

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
				So(response.PreviewTeams, ShouldBeNil)
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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: &storetest.StorerMock{}}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeMissingParameters := models.CodeMissingParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeMissingParameters,
							Description: apierrors.ErrorDescriptionMissingIfMatchHeader,
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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

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
							Description: apierrors.ErrorDescriptionConflict,
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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

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

				codeInvalidParameters := models.CodeInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &codeInvalidParameters,
							Description: apierrors.ErrorDescriptionMalformedRequest,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestPutBundle_BundleNotFound_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request for non-existent bundle", t, func() {
		updateRequest := &models.Bundle{
			Title:        "Updated Title",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return nil, apierrors.ErrBundleNotFound
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-missing", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": "bundle-missing"})
			r.Header.Set("If-Match", "some-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)

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
	})
}

func TestPutBundle_AuthenticationFailure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with invalid authentication", t, func() {
		updateRequest := &models.Bundle{
			Title:        "Updated Title",
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		bundleAPI := GetBundleAPIWithMocksWhereAuthFails(store.Datastore{Backend: &storetest.StorerMock{}}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

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

				codeInternalError := models.CodeInternalError
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

func TestPutBundle_MultipleValidationErrors_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request with multiple validation errors", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:        bundle1,
			Title:     "Original Title",
			ETag:      "original-etag",
			State:     models.BundleStateDraft,
			CreatedAt: &now,
			CreatedBy: &models.User{Email: "creator@example.com"},
		}

		updateRequest := map[string]interface{}{
			"state": "INVALID_STATE",
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
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 400 Bad Request with multiple errors", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				So(len(errResp.Errors), ShouldBeGreaterThan, 1)

				errorFields := make([]string, len(errResp.Errors))
				for i, err := range errResp.Errors {
					if err.Source != nil {
						errorFields[i] = err.Source.Field
					}
				}

				So(errorFields, ShouldContain, "/state")
				So(errorFields, ShouldContain, "/bundle_type")
				So(errorFields, ShouldContain, "/title")
				So(errorFields, ShouldContain, "/managed_by")
			})
		})
	})
}

func TestPutBundle_CreateBundlePolicies_Failure(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request where the permissions API call fails", t, func() {
		now := time.Now().UTC()
		existingBundle := &models.Bundle{
			ID:         bundle1,
			Title:      "Original Title",
			BundleType: models.BundleTypeManual,
			ETag:       "original-etag",
			State:      models.BundleStateDraft,
			CreatedAt:  &now,
			CreatedBy:  &models.User{Email: "creator@example.com"},
			UpdatedAt:  &now,
			ManagedBy:  models.ManagedByDataAdmin,
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
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
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return nil, errors.New("permissions API error")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, mockPermissionsClient, false)
		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				codeInternalError := models.CodeInternalError
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

func TestPutBundle_AddPolicyConditionsForAddedPreviewTeams_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request that adds a new preview team to a bundle with existing content", t, func() {
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
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		contentItems := []*models.ContentItem{
			{
				ID:       "content-1",
				BundleID: bundle1,
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
					VersionID: 1,
				},
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
				updatedBundle.ETag = newEtag
				updatedBundle.PreviewTeams = updateRequest.PreviewTeams
				return &updatedBundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return contentItems, nil
			},
		}

		existingPolicy := &permissionsAPIModels.Policy{
			Condition: permissionsAPIModels.Condition{
				Attribute: "dataset_edition",
				Operator:  "StringEquals",
				Values:    []string{},
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				if id == team1 {
					return &permissionsAPIModels.Policy{}, nil
				}
				if id == team2 {
					return existingPolicy, nil
				}
				return nil, errors.New("404 policy not found")
			},
			PostPolicyWithIDFunc: func(ctx context.Context, id string, policy permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return nil, nil
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, mockPermissionsClient, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And policy conditions should be added for the new team", func() {
				So(len(mockPermissionsClient.GetPolicyCalls()), ShouldEqual, 3)
				So(len(mockPermissionsClient.PutPolicyCalls()), ShouldEqual, 1)

				putCall := mockPermissionsClient.PutPolicyCalls()[0]
				So(putCall.ID, ShouldEqual, "team-2")
				So(putCall.Policy.Condition.Values, ShouldContain, "dataset-1")
				So(putCall.Policy.Condition.Values, ShouldContain, "dataset-1/edition-1")
			})
		})
	})
}

func TestPutBundle_RemovePolicyConditionsForRemovedPreviewTeams_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request that removes a preview team from a bundle", t, func() {
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
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}},
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		contentItems := []*models.ContentItem{
			{
				ID:       "content-1",
				BundleID: bundle1,
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
					VersionID: 1,
				},
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
				updatedBundle.ETag = newEtag
				updatedBundle.PreviewTeams = updateRequest.PreviewTeams
				return &updatedBundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return contentItems, nil
			},
			GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
				return []*models.Bundle{}, nil
			},
		}

		existingPolicy := &permissionsAPIModels.Policy{
			Condition: permissionsAPIModels.Condition{
				Attribute: "dataset_edition",
				Operator:  "StringEquals",
				Values:    []string{"dataset-1", "dataset-1/edition-1"},
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				if id == team1 {
					return &permissionsAPIModels.Policy{}, nil
				}
				if id == team2 {
					return existingPolicy, nil
				}
				return nil, errors.New("404 policy not found")
			},
			PostPolicyWithIDFunc: func(ctx context.Context, id string, policy permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return nil, nil
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, mockPermissionsClient, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And policy conditions should be removed for the removed team", func() {
				So(len(mockPermissionsClient.PutPolicyCalls()), ShouldEqual, 1)

				putCall := mockPermissionsClient.PutPolicyCalls()[0]
				So(putCall.ID, ShouldEqual, "team-2")
				So(putCall.Policy.Condition.Values, ShouldBeNil)
			})
		})
	})
}

func TestPutBundle_RemovePolicyConditionsForRemovedPreviewTeams_DatasetInOtherBundle(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request that removes a preview team from a bundle, but the dataset is in another bundle", t, func() {
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
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		contentItemsBundle1 := []*models.ContentItem{
			{
				ID:       "content-1",
				BundleID: bundle1,
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
					VersionID: 1,
				},
			},
		}

		contentItemsBundle2 := []*models.ContentItem{
			{
				ID:       "content-2",
				BundleID: "bundle-2",
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-2",
					VersionID: 1,
				},
			},
		}

		bundle2 := &models.Bundle{
			ID:           "bundle-2",
			Title:        "Bundle 2",
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
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
				updatedBundle.ETag = newEtag
				updatedBundle.PreviewTeams = updateRequest.PreviewTeams
				return &updatedBundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				if bundleID == bundle1 {
					return contentItemsBundle1, nil
				}
				if bundleID == "bundle-2" {
					return contentItemsBundle2, nil
				}
				return []*models.ContentItem{}, nil
			},
			GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
				if teamID == "team-1" {
					return []*models.Bundle{bundle2, existingBundle}, nil
				}
				return []*models.Bundle{}, nil
			},
		}

		existingPolicy := &permissionsAPIModels.Policy{
			Condition: permissionsAPIModels.Condition{
				Attribute: "dataset_edition",
				Operator:  "StringEquals",
				Values:    []string{"dataset-1", "dataset-1/edition-1", "dataset-1/edition-2"},
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				if id == team1 {
					return existingPolicy, nil
				}
				return nil, errors.New("404 policy not found")
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, mockPermissionsClient, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And only the edition should be removed, not the dataset", func() {
				So(len(mockPermissionsClient.PutPolicyCalls()), ShouldEqual, 1)

				putCall := mockPermissionsClient.PutPolicyCalls()[0]
				So(putCall.ID, ShouldEqual, "team-1")
				So(putCall.Policy.Condition.Values, ShouldContain, "dataset-1")
				So(putCall.Policy.Condition.Values, ShouldContain, "dataset-1/edition-2")
				So(putCall.Policy.Condition.Values, ShouldNotContain, "dataset-1/edition-1")
				So(len(putCall.Policy.Condition.Values), ShouldEqual, 2)
			})
		})
	})
}

func TestPutBundle_RemovePolicyConditionsForRemovedPreviewTeams_NoChanges(t *testing.T) {
	t.Parallel()

	Convey("Given a PUT request that removes a team, but the policy has no matching values", t, func() {
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
			PreviewTeams: &[]models.PreviewTeam{{ID: "team-1"}},
		}

		updateRequest := &models.Bundle{
			Title:        title1,
			BundleType:   models.BundleTypeManual,
			State:        models.BundleStateDraft,
			ManagedBy:    models.ManagedByDataAdmin,
			PreviewTeams: &[]models.PreviewTeam{},
		}

		updateRequestJSON, err := json.Marshal(updateRequest)
		So(err, ShouldBeNil)

		contentItems := []*models.ContentItem{
			{
				ID:       "content-1",
				BundleID: bundle1,
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
					VersionID: 1,
				},
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
				updatedBundle.ETag = newEtag
				updatedBundle.PreviewTeams = updateRequest.PreviewTeams
				return &updatedBundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return contentItems, nil
			},
			GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
				return []*models.Bundle{}, nil
			},
		}

		existingPolicy := &permissionsAPIModels.Policy{
			Condition: permissionsAPIModels.Condition{
				Attribute: "dataset_edition",
				Operator:  "StringEquals",
				Values:    []string{"different-dataset", "different-dataset/edition"},
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				if id == team1 {
					return existingPolicy, nil
				}
				return nil, errors.New("404 policy not found")
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, mockPermissionsClient, false)

		Convey("When putBundle is called", func() {
			r := httptest.NewRequest("PUT", "/bundles/bundle-1", bytes.NewReader(updateRequestJSON))
			r = mux.SetURLVars(r, map[string]string{"bundle-id": bundle1})
			r.Header.Set("If-Match", "original-etag")
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.putBundle(w, r)

			Convey("Then it should return 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And PutPolicy should NOT be called since values didn't change", func() {
				So(len(mockPermissionsClient.PutPolicyCalls()), ShouldEqual, 0)
			})
		})
	})
}
