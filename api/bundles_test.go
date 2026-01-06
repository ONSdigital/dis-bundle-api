package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	dphttp "github.com/ONSdigital/dp-net/v3/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/utils"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	scheduledTime = time.Date(2125, 6, 5, 7, 0, 0, 0, time.UTC)
	validBundle   = &models.Bundle{
		BundleType: models.BundleTypeScheduled,
		PreviewTeams: &[]models.PreviewTeam{
			{ID: "team1"},
			{ID: "team2"},
		},
		ScheduledAt: &scheduledTime,
		State:       models.BundleStateDraft,
		Title:       "Scheduled Bundle 1",
		ManagedBy:   models.ManagedByWagtail,
	}
	validBundleNoPreviewTeam = &models.Bundle{
		BundleType:  models.BundleTypeScheduled,
		ScheduledAt: &scheduledTime,
		State:       models.BundleStateDraft,
		Title:       "Scheduled Bundle 2",
		ManagedBy:   models.ManagedByWagtail,
	}
)

const MockAuthTokenValue = "test-auth-token"

const (
	someetag = "some-etag"
)

var MockAuthBearerHeaderValue = fmt.Sprintf("Bearer %s", MockAuthTokenValue)

type ParseFuncHandler = func(token string) (*permissionsAPISDK.EntityData, error)

func newAuthMiddlwareMock(validateAuth bool, parseFunc ParseFuncHandler) *authorisationMock.MiddlewareMock {
	if parseFunc == nil {
		parseFunc = func(token string) (*permissionsAPISDK.EntityData, error) {
			if token == MockAuthTokenValue {
				return &permissionsAPISDK.EntityData{
					UserID: "User123",
				}, nil
			}
			return nil, errors.New("authorisation header not found")
		}
	}

	return &authorisationMock.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			if !validateAuth {
				return handlerFunc
			}

			return func(w http.ResponseWriter, r *http.Request) {
				token := r.Header.Get("Authorization")
				isAuthorised := token != "" && token == MockAuthBearerHeaderValue

				if !isAuthorised {
					http.Error(w, `{"errors":[{"code":"Unauthorised","description":"Access denied."}]}`, http.StatusUnauthorized)
				} else if handlerFunc != nil {
					handlerFunc(w, r)
				}
			}
		},
		ParseFunc: parseFunc,
	}
}

func GetBundleAPIWithMocks(datastore store.Datastore, datasetAPIClient datasetAPISDK.Clienter, permissionsAPIClient permissionsAPISDK.Clienter, validateAuth bool) *BundleAPI {
	authMiddleware := newAuthMiddlwareMock(validateAuth, nil)
	return GetBundleAPIWithMocksWithAuthMiddleware(datastore, datasetAPIClient, permissionsAPIClient, authMiddleware, false)
}

func GetBundleAPIWithMocksWhereAuthFails(datastore store.Datastore, datasetAPIClient datasetAPISDK.Clienter, permissionsAPIClient permissionsAPISDK.Clienter, validateAuth bool) *BundleAPI {
	authMiddleware := newAuthMiddlwareMock(validateAuth, nil)
	return GetBundleAPIWithMocksWithAuthMiddleware(datastore, datasetAPIClient, permissionsAPIClient, authMiddleware, true)
}

// valid identity response for testing

var testIdentity = "myIdentity"
var testIdentityResponse = &dprequest.IdentityResponse{
	Identifier: testIdentity,
}

// utility function to generate Clienter mocks
func createHTTPClientMock(retCode int, retBody interface{}) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{}
		},
		SetPathsWithNoRetriesFunc: func([]string) {
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := json.Marshal(retBody)
			return &http.Response{
				StatusCode: retCode,
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
}

func GetBundleAPIWithMocksWithAuthMiddleware(datastore store.Datastore, datasetAPIClient datasetAPISDK.Clienter, permissionsAPIClient permissionsAPISDK.Clienter, authMiddleware *authorisationMock.MiddlewareMock, serviceAuthFails bool) *BundleAPI {
	ctx := context.Background()
	cfg := &config.Config{
		DefaultMaxLimit: 100,
	}
	r := mux.NewRouter()

	mockStates := []application.State{
		application.Draft,
		application.InReview,
		application.Approved,
		application.Published,
	}

	mockTransitions := []application.Transition{
		{
			Label:               "DRAFT",
			TargetState:         application.Draft,
			AllowedSourceStates: []string{"IN_REVIEW", "APPROVED"},
		},
		{
			Label:               "IN_REVIEW",
			TargetState:         application.InReview,
			AllowedSourceStates: []string{"DRAFT", "APPROVED"},
		},
		{
			Label:               "APPROVED",
			TargetState:         application.Approved,
			AllowedSourceStates: []string{"IN_REVIEW"},
		},
		{
			Label:               "PUBLISHED",
			TargetState:         application.Published,
			AllowedSourceStates: []string{"APPROVED"},
		},
	}
	stateMachine := application.NewStateMachine(ctx, mockStates, mockTransitions, datastore)

	stateMachineBundleAPI := &application.StateMachineBundleAPI{
		Datastore:            datastore,
		StateMachine:         stateMachine,
		DatasetAPIClient:     datasetAPIClient,
		PermissionsAPIClient: permissionsAPIClient,
		PermissionsAPIURL:    "http://localhost:25400",
	}

	var cliMock *dphttp.ClienterMock
	if !serviceAuthFails {
		cliMock = createHTTPClientMock(http.StatusOK, testIdentityResponse)
	} else {
		cliMock = createHTTPClientMock(500, nil)
	}

	return Setup(ctx, cfg, r, &datastore, stateMachineBundleAPI, authMiddleware, cliMock)
}

func createRequestWithAuth(method, target string, body io.Reader) *http.Request {
	r := httptest.NewRequest(method, target, body)
	ctx := r.Context()
	ctx = dprequest.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)
	return r
}

func TestGetBundles_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a GET request to /bundles", t, func() {
		now := time.Now().UTC()
		oneDayLater := now.Add(24 * time.Hour)
		twoDaysLater := now.Add(48 * time.Hour)
		defaultBundles := []*models.Bundle{
			{
				ID:            "bundle1",
				BundleType:    models.BundleTypeScheduled,
				CreatedBy:     &models.User{Email: "creator@example.com"},
				CreatedAt:     &now,
				LastUpdatedBy: &models.User{Email: "updater@example.com"},
				PreviewTeams:  &[]models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
				ScheduledAt:   &oneDayLater,
				State:         models.BundleStatePublished,
				Title:         "Scheduled Bundle 1",
				UpdatedAt:     &now,
				ManagedBy:     models.ManagedByDataAdmin,
			},
			{
				ID:            "bundle2",
				BundleType:    models.BundleTypeManual,
				CreatedBy:     &models.User{Email: "creator2@example.com"},
				CreatedAt:     &now,
				LastUpdatedBy: &models.User{Email: "updater2@example.com"},
				PreviewTeams:  &[]models.PreviewTeam{{ID: "team3"}},
				ScheduledAt:   &twoDaysLater,
				State:         models.BundleStateDraft,
				Title:         "Manual Bundle 2",
				UpdatedAt:     &now,
				ManagedBy:     models.ManagedByWagtail,
			},
		}

		bundleFilterFunc := func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
			if filters == nil || filters.PublishDate == nil {
				return defaultBundles, len(defaultBundles), nil
			}

			var filteredBundles []*models.Bundle

			timeTolerance := time.Second * 2
			for _, bundle := range defaultBundles {
				if bundle.ScheduledAt.Sub(*filters.PublishDate) < timeTolerance {
					filteredBundles = append(filteredBundles, bundle)
				}
			}

			return filteredBundles, len(filteredBundles), nil
		}

		Convey("When offset and limit values are default", func() {
			Convey("And publish_date filter is not supplied, then default values should be returned with no error", func() {
				r := httptest.NewRequest("GET", "http://localhost:29800/bundles", http.NoBody)
				w := httptest.NewRecorder()

				mockedDatastore := &storetest.StorerMock{
					ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
						return defaultBundles, len(defaultBundles), nil
					},
				}
				mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
				mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
				bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

				successResp, errResp := bundleAPI.getBundles(w, r, 10, 0)

				So(errResp, ShouldBeNil)
				So(successResp.Result.Items, ShouldResemble, defaultBundles)
				So(successResp.Result.TotalCount, ShouldEqual, len(defaultBundles))
			})

			Convey("And publish_date filter is supplied, then default values should be returned with no error", func() {
				paramValue := oneDayLater.UTC().Format(time.RFC3339)

				r := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:29800/bundles?%s=%s", filters.PublishDate, paramValue), http.NoBody)
				w := httptest.NewRecorder()

				mockedDatastore := &storetest.StorerMock{
					ListBundlesFunc: bundleFilterFunc,
				}

				expectedBundles := []*models.Bundle{
					defaultBundles[0],
				}
				bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

				successResp, errResp := bundleAPI.getBundles(w, r, 10, 0)

				So(errResp, ShouldBeNil)
				So(successResp.Result.Items, ShouldResemble, expectedBundles)
				So(successResp.Result.TotalCount, ShouldEqual, len(expectedBundles))
			})
		})

		Convey("When offset and limit values are custom", func() {
			r := httptest.NewRequest("GET", "http://localhost:29800/bundles?offset=1&limit=1", http.NoBody)
			w := httptest.NewRecorder()
			customBundles := defaultBundles[1:]

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
					So(offset, ShouldEqual, 1)
					So(limit, ShouldEqual, 1)
					return customBundles, len(customBundles), nil
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			successResp, err := bundleAPI.getBundles(w, r, 1, 1)
			Convey("Then custom paginated values should be returned with no error", func() {
				So(err, ShouldBeNil)
				So(successResp.Result.Items, ShouldResemble, customBundles)
				So(successResp.Result.TotalCount, ShouldEqual, len(customBundles))
			})
		})

		Convey("When no matching bundles are found for the publish date", func() {
			paramValue := time.Now().UTC().Format(time.RFC3339)

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: bundleFilterFunc,
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			Convey("Then it returns a 404 error", func() {
				url := fmt.Sprintf("http://localhost:29800/bundles?%s=%s", filters.PublishDate, paramValue)

				r := httptest.NewRequest("GET", url, http.NoBody)
				w := httptest.NewRecorder()

				successResp, err := bundleAPI.getBundles(w, r, 10, 0)
				So(successResp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.HTTPStatusCode, ShouldEqual, 404)
			})
		})
	})
}

func TestGetBundles_Failure(t *testing.T) {
	t.Parallel()
	Convey("Given a GET request to /bundles", t, func() {
		Convey("When the datastore returns an internal error", func() {
			r := httptest.NewRequest("GET", "http://localhost:29800/bundles", http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
					return nil, 0, errors.New("database failure")
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			successResp, errResp := bundleAPI.getBundles(w, r, 10, 0)
			Convey("Then the status code should be 500", func() {
				So(successResp, ShouldBeNil)
				So(errResp, ShouldNotBeNil)
				So(errResp.HTTPStatusCode, ShouldEqual, 500)
				So(errResp.Error.Description, ShouldEqual, apierrors.ErrorDescriptionInternalError)
			})
		})

		Convey("When response returns an random internal error", func() {
			r := httptest.NewRequest(http.MethodGet, "/bundles", http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
					return nil, 0, errors.New("something broke inside")
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)
			Convey("Then the status code should be 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldEqual, `{"errors":[{"code":"InternalError","description":"Failed to process the request due to an internal error."}]}`+"\n")
			})
		})

		Convey("When an invalid publish_date is supplied", func() {
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bundles?%s=%s", filters.PublishDate, "notactuallyadate"), http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
					return nil, 0, nil
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)
			Convey("Then the status code should be 400", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				expectedErrorCode := models.CodeInvalidParameters
				expectedErrorSource := models.Source{
					Parameter: "publish_date",
				}

				expectedError := &models.Error{
					Code:        &expectedErrorCode,
					Description: apierrors.ErrorDescriptionMalformedRequest,
					Source:      &expectedErrorSource,
				}

				var errList models.ErrorList
				errList.Errors = append(errList.Errors, expectedError)
				errBytes, err := json.Marshal(errList)
				if err != nil {
					fmt.Println(err)
					return
				}
				expectedErrorString := fmt.Sprintf("%s\n", string(errBytes))
				So(w.Body.String(), ShouldEqual, expectedErrorString)
			})
		})
	})
}

func TestGetBundle_Success(t *testing.T) {
	t.Parallel()

	scheduledAt := time.Date(2025, 4, 25, 9, 0, 0, 0, time.UTC)
	createdAt := time.Date(2025, 3, 10, 11, 20, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 3, 25, 14, 30, 0, 0, time.UTC)

	validBundle := &models.Bundle{
		ID:          "bundle-2",
		Title:       "bundle-2",
		ETag:        "12345-etag",
		ManagedBy:   models.ManagedByWagtail,
		BundleType:  models.BundleTypeScheduled,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
		ScheduledAt: &scheduledAt,
		State:       models.BundleStateDraft,
		CreatedBy: &models.User{
			Email: "publisher@ons.gov.uk",
		},
		LastUpdatedBy: &models.User{
			Email: "publisher@ons.gov.uk",
		},
		PreviewTeams: &[]models.PreviewTeam{
			{
				ID: "c78d457e-98de-11ec-b909-0242ac120002",
			},
		},
	}

	validBundleWithoutETag := &models.Bundle{
		ID:          "bundle-3",
		Title:       "bundle-3",
		ManagedBy:   models.ManagedByWagtail,
		BundleType:  models.BundleTypeScheduled,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
		ScheduledAt: &scheduledAt,
		State:       models.BundleStateDraft,
		CreatedBy: &models.User{
			Email: "publisher@ons.gov.uk",
		},
		LastUpdatedBy: &models.User{
			Email: "publisher@ons.gov.uk",
		},
		PreviewTeams: &[]models.PreviewTeam{
			{
				ID: "c78d457e-98de-11ec-b909-0242ac120003",
			},
		},
	}

	Convey("Given a GET /bundles/{bundle-id} request", t, func() {
		Convey("When the bundle-id is valid", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/valid-id", http.NoBody)
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return validBundle, nil
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should have status 200, ETag and Cache-Control headers", func() {
				So(rec.Code, ShouldEqual, http.StatusOK)
				So(rec.Header().Get("ETag"), ShouldEqual, validBundle.ETag)
				So(rec.Header().Get("Cache-Control"), ShouldEqual, "no-store")

				var response models.Bundle
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				So(response.ID, ShouldEqual, validBundle.ID)
				So(response.Title, ShouldEqual, validBundle.Title)
				So(response.BundleType, ShouldEqual, validBundle.BundleType)
				So(response.ManagedBy, ShouldEqual, validBundle.ManagedBy)

				So(response.CreatedAt.Unix(), ShouldEqual, validBundle.CreatedAt.Unix())
				So(response.UpdatedAt.Unix(), ShouldEqual, validBundle.UpdatedAt.Unix())
				So(response.ScheduledAt.Unix(), ShouldEqual, validBundle.ScheduledAt.Unix())

				So(response.CreatedBy, ShouldNotBeNil)
				So(response.CreatedBy.Email, ShouldEqual, validBundle.CreatedBy.Email)

				So(response.LastUpdatedBy, ShouldNotBeNil)
				So(response.LastUpdatedBy.Email, ShouldEqual, validBundle.LastUpdatedBy.Email)

				So(response.State, ShouldNotBeNil)
				So(response.State, ShouldEqual, validBundle.State)

				So(response.PreviewTeams, ShouldNotBeNil)
				So(len(*response.PreviewTeams), ShouldEqual, 1)
				So((*response.PreviewTeams)[0].ID, ShouldEqual, "c78d457e-98de-11ec-b909-0242ac120002")
			})
		})
		Convey("When the bundle-id is valid but without ETag", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/valid-id", http.NoBody)
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return validBundleWithoutETag, nil
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should have status 200, ETag and Cache-Control headers", func() {
				So(rec.Code, ShouldEqual, http.StatusOK)
				So(rec.Header().Get("ETag"), ShouldNotBeEmpty)
				So(rec.Header().Get("Cache-Control"), ShouldEqual, "no-store")
			})
		})
	})
}

func TestGetBundle_Failure(t *testing.T) {
	t.Parallel()
	Convey("Given a GET /bundles/{bundle-id} request", t, func() {
		ctx := context.Background()

		Convey("When the bundle-id is invalid", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/invalid-id", http.NoBody)
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 404 with structured NotFound error", func() {
				So(rec.Code, ShouldEqual, http.StatusNotFound)

				var errResp models.ErrorList
				err := json.NewDecoder(rec.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeNotFound
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionNotFound,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When the request causes an internal error", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/valid-id", http.NoBody)
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return nil, errors.New("unexpected failure")
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 500 with structured InternalError", func() {
				So(rec.Code, ShouldEqual, http.StatusInternalServerError)

				var errResp models.ErrorList
				err := json.NewDecoder(rec.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When no valid authentication token is provided", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/protected-id", http.NoBody)
			rec := httptest.NewRecorder()

			authMiddleware := &authorisationMock.MiddlewareMock{
				RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						http.Error(w, `{"errors":[{"code":"Unauthorised","description":"Access denied."}]}`, http.StatusUnauthorized)
					}
				},
			}

			cliMock := createHTTPClientMock(500, nil)

			bundleAPI := Setup(ctx, &config.Config{}, mux.NewRouter(), nil, nil, authMiddleware, cliMock)
			bundleAPI.Router.HandleFunc("/bundles/{bundle-id}", func(w http.ResponseWriter, r *http.Request) {}).Methods(http.MethodGet)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 401 with structured Unauthorised error", func() {
				So(rec.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}

func newTestBundle(id string, state models.BundleState) *models.Bundle {
	now := time.Now()
	return &models.Bundle{
		ID:          id,
		Title:       fmt.Sprintf("Test Bundle %s", id),
		ETag:        fmt.Sprintf("%s-etag", id),
		ManagedBy:   models.ManagedByWagtail,
		BundleType:  models.BundleTypeScheduled,
		CreatedAt:   &now,
		UpdatedAt:   &now,
		ScheduledAt: &now,
		State:       state,
		CreatedBy: &models.User{
			Email: "test@example.com",
		},
		LastUpdatedBy: &models.User{
			Email: "original-email@ons.com",
		},
	}
}

func newTestContentItem(id, bundleID, datasetID, editionID string, versionID int, state *models.State) *models.ContentItem {
	return &models.ContentItem{
		ID:       id,
		BundleID: bundleID,
		Metadata: models.Metadata{
			DatasetID: datasetID,
			EditionID: editionID,
			VersionID: versionID,
		},
		State: state,
	}
}

func newTestVersion(id, datasetID, editionID string, versionID int, state string) *datasetAPIModels.Version {
	return &datasetAPIModels.Version{
		ID:        id,
		DatasetID: datasetID,
		Edition:   editionID,
		Version:   versionID,
		State:     state,
	}
}

type testData struct {
	bundle       *models.Bundle
	contentItems []*models.ContentItem
	versions     []*datasetAPIModels.Version
	events       *[]*models.Event
}

func setupTestData(initialState models.BundleState, contentItemState *models.State, versionState string) *testData {
	bundle := newTestBundle("test-bundle-id", initialState)

	contentItems := []*models.ContentItem{
		// Content items that match bundle state + we expect to be updated
		newTestContentItem("matching-item-1", bundle.ID, "dataset-1", "edition-1", 1, contentItemState),
		newTestContentItem("matching-item-2", bundle.ID, "dataset-2", "edition-2", 2, contentItemState),
	}

	versions := []*datasetAPIModels.Version{
		// Versions that are in the correct state and should be updated
		newTestVersion("version-1", "dataset-1", "edition-1", 1, versionState),
		newTestVersion("version-2", "dataset-2", "edition-2", 2, versionState),
	}
	events := make([]*models.Event, 0)
	return &testData{
		bundle:       bundle,
		contentItems: contentItems,
		versions:     versions,
		events:       &events,
	}
}

func createMockStore(data *testData) *storetest.StorerMock {
	return &storetest.StorerMock{
		GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
			if data.bundle.ID == id {
				return data.bundle, nil
			}
			return nil, apierrors.ErrBundleNotFound
		},
		GetBundleContentsForBundleFunc: func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
			var items []models.ContentItem
			for _, item := range data.contentItems {
				if item.BundleID == bundleID {
					items = append(items, *item)
				}
			}
			return &items, nil
		},
		UpdateContentItemStateFunc: func(ctx context.Context, contentItemID, state string) error {
			for _, item := range data.contentItems {
				if item.ID == contentItemID {
					mappedState := models.State(state)
					item.State = &mappedState
					return nil
				}
			}
			return apierrors.ErrContentItemNotFound
		},
		UpdateBundleFunc: func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
			if data.bundle.ID == id {
				data.bundle.State = update.State
				return data.bundle, nil
			}
			return nil, apierrors.ErrBundleNotFound
		},
		CreateEventFunc: func(ctx context.Context, event *models.Event) error {
			*data.events = append(*data.events, event)
			return nil
		},
	}
}

func createMockDatasetAPIClient(data *testData) *datasetAPISDKMock.ClienterMock {
	return &datasetAPISDKMock.ClienterMock{
		GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
			for _, version := range data.versions {
				versionString := strconv.Itoa(version.Version)
				if versionString == versionID && version.DatasetID == datasetID && version.Edition == editionID {
					return *version, nil
				}
			}
			return datasetAPIModels.Version{}, errors.New("version not found")
		},
		PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
			for _, version := range data.versions {
				versionString := strconv.Itoa(version.Version)
				if versionString == versionID && version.DatasetID == datasetID && version.Edition == editionID {
					version.State = state
					return nil
				}
			}
			return errors.New("version not found")
		},
	}
}

// Helper functions
func shouldItemBeUpdated(itemID string) bool {
	return strings.Contains(itemID, "matching")
}

// createUpdateStateRequest creates a http.Request for updating bundle state.
func createUpdateStateRequest(bundleID, etag string, state models.BundleState, requestBody interface{}, isAuthorised bool) *http.Request {
	if requestBody == nil {
		requestBody = models.UpdateStateRequest{State: state}
	}

	b, err := json.Marshal(requestBody)
	if err != nil {
		panic("failed to marshal update bundle state request")
	}

	reader := bytes.NewReader(b)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/bundles/%s/state", bundleID), reader)

	if isAuthorised {
		req.Header.Set("Authorization", MockAuthBearerHeaderValue)
	}

	if etag != "" {
		req.Header.Add(utils.HeaderIfMatch, etag)
	}
	return req
}

var validTransitionTestCases = []struct {
	name             string
	fromState        models.BundleState
	toState          models.BundleState
	invalidState     models.BundleState
	contentItemState *models.State
	versionState     string
}{
	{"DRAFT to IN_REVIEW", models.BundleStateDraft, models.BundleStateInReview, models.BundleStatePublished, nil, "associated"},
	{"IN_REVIEW to APPROVED", models.BundleStateInReview, models.BundleStateApproved, models.BundleStateDraft, nil, "associated"},
	{"IN_REVIEW to APPROVED with version already approved", models.BundleStateInReview, models.BundleStateApproved, models.BundleStateDraft, nil, "approved"},
	{"APPROVED to PUBLISHED", models.BundleStateApproved, models.BundleStatePublished, models.BundleStateDraft, utils.PtrContentItemState(models.StateApproved), "approved"},
	{"APPROVED to PUBLISHED with version already published", models.BundleStateApproved, models.BundleStatePublished, models.BundleStateDraft, utils.PtrContentItemState(models.StateApproved), "published"},
}

const (
	PutBundleStateRoute = "/bundles/{bundle-id}/state"
	ValidDataset1       = "dataset-1"
	ValidDataset2       = "dataset-2"
)

//nolint:gocognit // cognitive complexity is high but acceptable for this test
func TestPutBundleState_ValidTransitions(t *testing.T) {
	t.Parallel()

	for index := range validTransitionTestCases {
		tc := validTransitionTestCases[index]

		var additionalEventCalls int
		if tc.toState == models.BundleStateApproved || tc.toState == models.BundleStatePublished {
			// manually set to 2 since we have 2 content items that will be updated within the test data
			additionalEventCalls = 2
		}

		testCaseName := generateTestCaseName(PutBundleStateRoute, "PUT", fmt.Sprintf("With a valid request and a valid state transition to update state from %s", tc.name))
		t.Run(testCaseName, func(t *testing.T) {
			t.Parallel()

			Convey(tc.name, t, func() {
				data := setupTestData(tc.fromState, tc.contentItemState, tc.versionState)
				mockStore := createMockStore(data)
				mockDatasetAPIClient := createMockDatasetAPIClient(data)
				mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
				bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, true)

				req := createUpdateStateRequest(data.bundle.ID, data.bundle.ETag, tc.toState, nil, true)
				rec := httptest.NewRecorder()

				bundleAPI.Router.ServeHTTP(rec, req)

				Convey(fmt.Sprintf("Then the bundle state should be updated to %s", tc.toState.String()), func() {
					So(data.bundle.State.String(), ShouldEqual, tc.toState.String())
				})

				Convey("And the response should be HTTP 200 OK", func() {
					So(rec.Code, ShouldEqual, http.StatusOK)
				})

				Convey("And the response body should contain the updated bundle with the new state", func() {
					var responseBundle models.Bundle
					err := json.NewDecoder(rec.Body).Decode(&responseBundle)
					So(err, ShouldBeNil)

					So(responseBundle.ID, ShouldEqual, data.bundle.ID)
					So(responseBundle.State.String(), ShouldEqual, tc.toState.String())
				})

				Convey("And an event should be created", func() {
					So(len(mockStore.CreateEventCalls()), ShouldEqual, 1+additionalEventCalls)
				})

				Convey("And only matching content items should be updated", func() {
					calls := mockStore.UpdateContentItemStateCalls()

					So(len(calls), ShouldEqual, 0+additionalEventCalls)

					if len(calls) > 0 {
						for _, item := range data.contentItems {
							if shouldItemBeUpdated(item.ID) {
								So(item.State.String(), ShouldEqual, tc.toState.String())
							} else {
								So(item.State.String(), ShouldNotEqual, tc.toState.String())
							}
						}
					}
				})

				expectedVersionState := strings.ToLower(tc.toState.String())

				Convey("And only matching versions should be updated", func() {
					calls := mockDatasetAPIClient.PutVersionStateCalls()

					// TODO: remove if condition and keep else block once we know if approved or published versions can be added to bundles
					if strings.EqualFold(tc.toState.String(), tc.versionState) {
						So(len(calls), ShouldEqual, 0)
					} else {
						So(len(calls), ShouldEqual, 0+additionalEventCalls)
					}

					if len(calls) > 0 {
						for _, call := range calls {
							So(call.State, ShouldEqual, expectedVersionState)
							isMatchingDataset := call.DatasetID == ValidDataset1 || call.DatasetID == ValidDataset2
							So(isMatchingDataset, ShouldBeTrue)
						}
					}
				})

				Convey("Then version states should be updated", func() {
					for _, version := range data.versions {
						isMatchingDataset := version.DatasetID == ValidDataset1 || version.DatasetID == ValidDataset2
						if isMatchingDataset && (tc.toState == models.BundleStateApproved || tc.toState == models.BundleStatePublished) {
							So(version.State, ShouldEqual, expectedVersionState)
						} else {
							So(version.State, ShouldNotEqual, expectedVersionState)
						}
					}
				})

				Convey("And events should have been created", func() {
					So(len(*data.events), ShouldEqual, 1+additionalEventCalls)

					for index := range *data.events {
						event := (*data.events)[index]

						So(event.Action, ShouldEqual, models.ActionUpdate)
					}
				})
			})
		})
	}
}

var invalidTransitionTestCases = []struct {
	name             string
	fromState        models.BundleState
	toState          models.BundleState
	invalidState     models.BundleState
	contentItemState *models.State
	versionState     string
}{
	{"DRAFT to PUBLISHED", models.BundleStateDraft, models.BundleStatePublished, models.BundleStateInReview, nil, "associated"},
	{"IN_REVIEW to PUBLISHED", models.BundleStateInReview, models.BundleStatePublished, models.BundleStateApproved, nil, "associated"},
	{"PUBLISHED to DRAFT", models.BundleStatePublished, models.BundleStateDraft, models.BundleStateApproved, utils.PtrContentItemState(models.StatePublished), "published"},
}

func TestPutBundleState_InvalidStateTransitions(t *testing.T) {
	t.Parallel()

	for index := range invalidTransitionTestCases {
		tc := invalidTransitionTestCases[index]
		testCaseName := generateTestCaseName(PutBundleStateRoute, "PUT", fmt.Sprintf("With a valid request to and an invalid state request to update state from %s", tc.name))

		t.Run(testCaseName, func(t *testing.T) {
			t.Parallel()

			Convey(tc.name, t, func() {
				data := setupTestData(tc.fromState, tc.contentItemState, tc.versionState)
				mockStore := createMockStore(data)
				mockDatasetAPIClient := createMockDatasetAPIClient(data)
				mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
				bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, true)

				req := createUpdateStateRequest(data.bundle.ID, data.bundle.ETag, tc.toState, nil, true)
				rec := httptest.NewRecorder()

				bundleAPI.Router.ServeHTTP(rec, req)

				Convey("Then a bad request status should be returned", func() {
					So(rec.Code, ShouldEqual, http.StatusBadRequest)
				})

				Convey("And the bundle state should not be updated", func() {
					So(data.bundle.State.String(), ShouldEqual, tc.fromState.String())
				})

				Convey("And the relevant error response body should be returned", func() {
					var errResp models.ErrorList
					err := json.NewDecoder(rec.Body).Decode(&errResp)
					So(err, ShouldBeNil)

					expectedErrResp := models.ErrorList{
						Errors: []*models.Error{
							models.ErrorToModelErrorMap[apierrors.ErrInvalidBundleState],
						},
					}
					So(errResp, ShouldResemble, expectedErrResp)
				})

				Convey("And no events should have been created", func() {
					So(len(*data.events), ShouldEqual, 0)
				})
			})
		})
	}
}

func TestPutBundleState_InternalErrors(t *testing.T) {
	t.Parallel()

	fromState := models.BundleStateApproved
	expectedErrorResponse := models.ErrorList{
		Errors: []*models.Error{
			models.ErrorToModelErrorMap[apierrors.ErrInternalServer],
		},
	}

	Convey(generateTestCaseName(PutBundleStateRoute, "PUT", "When the auth middleware's Parse method has an error"), t, func() {
		data := setupTestData(fromState, nil, "")
		mockStore := createMockStore(data)
		mockDatasetAPIClient := createMockDatasetAPIClient(data)
		mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
		parseFunc := func(token string) (*permissionsAPISDK.EntityData, error) {
			return nil, errors.New("Parse func error")
		}

		mockAuthMiddleware := newAuthMiddlwareMock(true, parseFunc)
		bundleAPI := GetBundleAPIWithMocksWithAuthMiddleware(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, mockAuthMiddleware, true)

		req := createUpdateStateRequest(data.bundle.ID, data.bundle.ETag, models.BundleStatePublished, nil, true)
		rec := httptest.NewRecorder()

		bundleAPI.Router.ServeHTTP(rec, req)

		Convey("Then an internal server error status should be returned", func() {
			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("And the bundle state should not be updated", func() {
			So(data.bundle.State.String(), ShouldEqual, fromState.String())
		})

		Convey("And the relevant error response body should be returned", func() {
			var errResp models.ErrorList
			err := json.NewDecoder(rec.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			So(errResp, ShouldResemble, expectedErrorResponse)
		})

		Convey("And no events should have been created", func() {
			So(len(*data.events), ShouldEqual, 0)
		})
	})
}

func TestPutBundleState_BadRequests(t *testing.T) {
	t.Parallel()
	data := setupTestData(models.BundleStateApproved, nil, "")
	mockStore := createMockStore(data)
	mockDatasetAPIClient := createMockDatasetAPIClient(data)
	mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
	bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore}, mockDatasetAPIClient, mockPermissionsAPIClient, true)

	UpdateStateRequestBodyValid := models.UpdateStateRequest{State: models.BundleStateApproved}
	UpdateStateRequestBodyMissingStateField := models.UpdateStateRequest{}
	UpdateStateRequestBodyInvalidStateField := models.UpdateStateRequest{State: models.BundleState("not-an-actual-state")}
	UpdateStateRequestBodyInvalidType := struct {
		name       string
		someNumber int
	}{
		name:       "some name",
		someNumber: 12345,
	}

	var errorTestCases = []struct {
		name               string
		bundleID           string
		etag               string
		expectedStatusCode int
		expectedError      *models.Error
		authorised         bool        // Whether to make the request authorised or not
		requestBody        interface{} // Request body to send.
	}{
		{"With a non-existent bundle ID", "missing-bundle-id", "etag", 404, models.ErrorToModelErrorMap[apierrors.ErrBundleNotFound], true, UpdateStateRequestBodyValid},
		{"Where the supplied ETag doesn't match the bundle", data.bundle.ID, "etag", 409, models.ErrorToModelErrorMap[apierrors.ErrInvalidIfMatchHeader], true, UpdateStateRequestBodyValid},
		{"Where no ETag value supplied", data.bundle.ID, "", 400, models.ErrorToModelErrorMap[apierrors.ErrMissingIfMatchHeader], true, UpdateStateRequestBodyValid},
		{"With an unauthenticated request", data.bundle.ID, data.bundle.ETag, 401, models.ErrorToModelErrorMap[apierrors.ErrUnauthorised], false, UpdateStateRequestBodyValid},
		{"With a request body missing the state field", data.bundle.ID, data.bundle.ETag, 400, models.ErrorToModelErrorMap[apierrors.ErrInvalidBody], true, UpdateStateRequestBodyMissingStateField},
		{"With a request body that has an invalid state field", data.bundle.ID, data.bundle.ETag, 400, models.ErrorToModelErrorMap[apierrors.ErrInvalidBody], true, UpdateStateRequestBodyInvalidStateField},
		{"With a request body that is not the correct type", data.bundle.ID, data.bundle.ETag, 400, models.ErrorToModelErrorMap[apierrors.ErrInvalidBody], true, UpdateStateRequestBodyInvalidType},
	}

	for index := range errorTestCases {
		tc := errorTestCases[index]

		// Using unicode divison slash here to prevent the test name being split into an unnecessary hierarchy
		testName := generateTestCaseName(PutBundleStateRoute, "PUT", tc.name)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			Convey(tc.name, t, func() {
				req := createUpdateStateRequest(tc.bundleID, tc.etag, models.BundleStatePublished, tc.requestBody, tc.authorised)
				rec := httptest.NewRecorder()

				bundleAPI.Router.ServeHTTP(rec, req)

				Convey("Then it should return the appropriate error code", func() {
					So(rec.Code, ShouldEqual, tc.expectedStatusCode)
				})

				Convey("And it should return the appropriate error response body", func() {
					var errResp models.ErrorList
					err := json.NewDecoder(rec.Body).Decode(&errResp)
					So(err, ShouldBeNil)

					expectedErrResp := models.ErrorList{
						Errors: []*models.Error{
							tc.expectedError,
						},
					}
					So(errResp, ShouldResemble, expectedErrResp)
				})
			})
		})
	}
}

// generateTestCaseName generates a formatted test case name, with the route escaped
func generateTestCaseName(route, method, testCase string) string {
	// Using unicode divison slash here to prevent the test name being split into an unnecessary hierarchy
	escapedRoute := strings.ReplaceAll(route, "/", "∕")

	// Forward slash here separates this into hierarchy:
	// 1. "Given a X request to ROUTE"
	// 2. "TEST CASE"
	return fmt.Sprintf("Given a %s request to %s/%s", method, escapedRoute, testCase)
}

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return nil
				},
				GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
					//nolint:goconst //The strings aren't actually the same.
					inputBundle.ID = "bundle1"
					inputBundle.ETag = someetag
					inputBundle.CreatedBy = &models.User{Email: "example@example.com"}
					inputBundle.LastUpdatedBy = &models.User{Email: "example@example.com"}
					now := time.Now()
					inputBundle.UpdatedAt = &now
					return &inputBundle, nil
				},
				CreateEventFunc: func(ctx context.Context, event *models.Event) error {
					return nil
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				PostPolicyWithIDFunc: func(ctx context.Context, headers permissionsAPISDK.Headers, id string, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
					return nil, nil
				},
			}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 201 Created with the created bundle", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
				var createdBundle models.Bundle
				err := json.Unmarshal(w.Body.Bytes(), &createdBundle)
				So(err, ShouldBeNil)
				So(createdBundle.ID, ShouldEqual, "bundle1")
				So(createdBundle.BundleType, ShouldEqual, validBundle.BundleType)
				So(createdBundle.CreatedBy.Email, ShouldEqual, "example@example.com")
				So(createdBundle.LastUpdatedBy.Email, ShouldEqual, "example@example.com")
				So(createdBundle.PreviewTeams, ShouldEqual, validBundle.PreviewTeams)
				So(createdBundle.ScheduledAt, ShouldEqual, validBundle.ScheduledAt)
				So(createdBundle.State, ShouldEqual, validBundle.State)
				So(createdBundle.Title, ShouldEqual, validBundle.Title)
				So(createdBundle.UpdatedAt, ShouldNotBeNil)
				So(createdBundle.ManagedBy, ShouldEqual, validBundle.ManagedBy)
			})
			Convey("And the correct headers should be set", func() {
				So(w.Header().Get("Location"), ShouldEqual, "/bundles/bundle1")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
				So(w.Header().Get("ETag"), ShouldEqual, "some-etag")
			})
			Convey("And a policy should be created for each preview team", func() {
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, len(*validBundle.PreviewTeams))
			})
		})
	})
}

func TestCreateBundleNoPreviewTeam_Success(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundleNoPreviewTeam
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return nil
				},
				GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
					inputBundle.ID = "bundle1"
					inputBundle.ETag = "some-etag"
					inputBundle.CreatedBy = &models.User{Email: "example@example.com"}
					inputBundle.LastUpdatedBy = &models.User{Email: "example@example.com"}
					now := time.Now()
					inputBundle.UpdatedAt = &now
					return &inputBundle, nil
				},
				CreateEventFunc: func(ctx context.Context, event *models.Event) error {
					return nil
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				PostPolicyWithIDFunc: func(ctx context.Context, headers permissionsAPISDK.Headers, id string, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
					return nil, nil
				},
			}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 201 Created with the created bundle", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
				var createdBundle models.Bundle
				err := json.Unmarshal(w.Body.Bytes(), &createdBundle)
				So(err, ShouldBeNil)
				So(createdBundle.ID, ShouldEqual, "bundle1")
				So(createdBundle.BundleType, ShouldEqual, validBundleNoPreviewTeam.BundleType)
				So(createdBundle.CreatedBy.Email, ShouldEqual, "example@example.com")
				So(createdBundle.LastUpdatedBy.Email, ShouldEqual, "example@example.com")
				So(createdBundle.PreviewTeams, ShouldBeNil)
				So(createdBundle.ScheduledAt, ShouldEqual, validBundleNoPreviewTeam.ScheduledAt)
				So(createdBundle.State, ShouldEqual, validBundleNoPreviewTeam.State)
				So(createdBundle.Title, ShouldEqual, validBundleNoPreviewTeam.Title)
				So(createdBundle.UpdatedAt, ShouldNotBeNil)
				So(createdBundle.ManagedBy, ShouldEqual, validBundleNoPreviewTeam.ManagedBy)
			})
			Convey("And the correct headers should be set", func() {
				So(w.Header().Get("Location"), ShouldEqual, "/bundles/bundle1")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
				So(w.Header().Get("ETag"), ShouldEqual, "some-etag")
			})
			Convey("And no policy should be created since no preview teams were provided", func() {
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestCreateBundle_Failure_FailedToParseBody(t *testing.T) {
	Convey("Given an invalid payload", t, func() {
		b := "{invalid_json"

		Convey("When a POST request is made to /bundles endpoint with the invalid payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			r.Header.Set("Authorization", "test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
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

func TestCreateBundle_Failure_InvalidScheduledAt(t *testing.T) {
	Convey("Given a payload with invalid scheduled_at format", t, func() {
		b := `{
			"scheduled_at": "invalid-date-format"
		}`

		Convey("When a POST request is made to /bundles endpoint with the invalid payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			r.Header.Set("Authorization", "test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInvalidTimeFormat,
							Source: &models.Source{
								Field: "scheduled_at",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock read error")
}
func TestCreateBundle_Failure_ReaderReturnError(t *testing.T) {
	Convey("Given a request with a reader that returns an error", t, func() {
		r := createRequestWithAuth(http.MethodPost, "/bundles", &ErrorReader{})
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

		bundleAPI.Router.ServeHTTP(w, r)

		Convey("Then the response should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("And the response body should contain an error message", func() {
			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			code := models.CodeInternalError
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &code,
						Description: apierrors.ErrorDescriptionInternalError,
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})
	})
}

func TestCreateBundle_Failure_ValidationError(t *testing.T) {
	Convey("Given a payload with missing mandatory fields", t, func() {
		Convey("When a POST request is made to /bundles endpoint with the invalid payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader([]byte("{}")))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with validation errors", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
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
							Source: &models.Source{
								Field: "/bundle_type",
							},
						},
						{
							Code:        &codeMissingParameters,
							Description: apierrors.ErrorDescriptionMissingParameters,
							Source: &models.Source{
								Field: "/state",
							},
						},
						{
							Code:        &codeMissingParameters,
							Description: apierrors.ErrorDescriptionMissingParameters,
							Source: &models.Source{
								Field: "/title",
							},
						},
						{
							Code:        &codeMissingParameters,
							Description: apierrors.ErrorDescriptionMissingParameters,
							Source: &models.Source{
								Field: "/managed_by",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_FailedToTransitionBundleState(t *testing.T) {
	Convey("Given a payload with invalid state for creating a bundle", t, func() {
		inputBundle := *validBundle
		inputBundle.State = models.BundleStateApproved
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
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
							Description: apierrors.ErrorDescriptionStateNotAllowedToTransition,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_AuthTokenIsMissing(t *testing.T) {
	Convey("Given a payload for creating a bundle ", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{}
		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocksWhereAuthFails(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

		Convey("When a POST request is made to /bundles endpoint with the payload and the auth token is missing", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When a POST request is made to /bundles endpoint with the payload and the auth token is invalid", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "test auth token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_CheckBundleExistsByTitleFails(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		inputBundleNonExistentTitle := *validBundle
		inputBundleNonExistentTitle.Title = "unique title"
		inputBundleNonExistentTitleJSON, err := json.Marshal(inputBundleNonExistentTitle)
		So(err, ShouldBeNil)

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				if title == "Scheduled Bundle 1" {
					return true, nil
				}
				return false, errors.New("failed to check bundle existence")
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

		Convey("When a POST request is made to /bundles endpoint with the payload and GetBundleByTitle fails", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleNonExistentTitleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When a POST request is made to /bundles endpoint with the payload and there is a bundle with the same title", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 409 conflict", func() {
				So(w.Code, ShouldEqual, http.StatusConflict)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeConflict
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionBundleTitleAlreadyExist,
							Source: &models.Source{
								Field: "/title",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})

		Convey("When a POST request is made to /bundles but the policies fail to create", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return nil
				},
				GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
					//nolint:goconst //The strings aren't actually the same.
					inputBundle.ID = "bundle1"
					inputBundle.ETag = someetag
					inputBundle.CreatedBy = &models.User{Email: "example@example.com"}
					inputBundle.LastUpdatedBy = &models.User{Email: "example@example.com"}
					now := time.Now()
					inputBundle.UpdatedAt = &now
					return &inputBundle, nil
				},
				CreateEventFunc: func(ctx context.Context, event *models.Event) error {
					return nil
				},
			}
			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				PostPolicyWithIDFunc: func(ctx context.Context, headers permissionsAPISDK.Headers, id string, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("failed to create policy")
				},
			}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_CreateBundleReturnsAnError(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and CreateBundle returns an error", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return errors.New("failed to create bundle")
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_CreateEventReturnsAnError(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		createdBundle := *validBundle
		createdBundle.CreatedBy = &models.User{Email: "example@example.com"}

		Convey("When a POST request is made to /bundles endpoint with the payload and CreateEvent returns an error", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return nil
				},
				CreateEventFunc: func(ctx context.Context, event *models.Event) error {
					return errors.New("failed to create event")
				},
				GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
					return &createdBundle, nil
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_ScheduledAtNotSet(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundle.ScheduledAt = nil // Set ScheduledAt to nil to simulate a scheduled bundle without a scheduled time
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and scheduled at is not set for a scheduled bundle", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return apierrors.ErrScheduledAtRequired
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionScheduledAtIsRequired,
							Source: &models.Source{
								Field: "/scheduled_at",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_ScheduledAtSetForManualBundles(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundle.BundleType = models.BundleTypeManual // Change to Manual Bundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and scheduled at is set for a manual bundle", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return apierrors.ErrScheduledAtSet
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionScheduledAtShouldNotBeSet,
							Source: &models.Source{
								Field: "/scheduled_at",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestCreateBundle_Failure_ScheduledAtIsInThePast(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		pastTime := time.Now().Add(-24 * time.Hour)
		inputBundle.ScheduledAt = &pastTime
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and scheduled at is set in the past", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer test-auth-token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
					return false, nil
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return apierrors.ErrScheduledAtInPast
				},
			}

			mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient, mockPermissionsAPIClient, false)

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
			Convey("And the response body should contain an error message", func() {
				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInvalidParameters
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionScheduledAtIsInPast,
							Source: &models.Source{
								Field: "/scheduled_at",
							},
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestDeleteBundle_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE /bundles/{bundle-id} request", t, func() {
		r := httptest.NewRequest(http.MethodDelete, "/bundles/bundle-1", http.NoBody)
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				if id == bundle1 {
					return &models.Bundle{
						ID:    "bundle-1",
						State: models.BundleStateDraft,
					}, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID: "content-1",
					},
					{
						ID: "content-2",
					},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				if contentItemID == "content-1" || contentItemID == "content-2" {
					return nil
				}
				return errors.New("failed to delete content item")
			},
			DeleteBundleFunc: func(ctx context.Context, id string) error {
				return nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, nil, nil, false)
		bundleAPI.Router.ServeHTTP(w, r)

		Convey("Then the response should be 204 No Content", func() {
			So(w.Code, ShouldEqual, http.StatusNoContent)
		})
	})
}

func TestDeleteBundle_Failure_UnableToParseToken(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE /bundles/{bundle-id} request with an invalid auth token", t, func() {
		r := httptest.NewRequest(http.MethodDelete, "/bundles/bundle-1", http.NoBody)
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{}
		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, nil, nil, false)

		Convey("When the request is made", func() {
			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				var errResp models.ErrorList
				err := json.NewDecoder(w.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: apierrors.ErrorDescriptionInternalError,
						},
					},
				}
				So(errResp, ShouldResemble, expectedErrResp)
			})
		})
	})
}

func TestDeleteBundle_Failure_BundleNonExistent(t *testing.T) {
	t.Parallel()

	Convey("Given a DELETE /bundles/{bundle-id} request", t, func() {
		r := httptest.NewRequest(http.MethodDelete, "/bundles/bundle-2", http.NoBody)
		r.Header.Set("Authorization", "test-auth-token")
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				if id == bundle1 {
					return &models.Bundle{
						ID:    "bundle-1",
						State: models.BundleStateDraft,
					}, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore}, nil, nil, false)
		bundleAPI.Router.ServeHTTP(w, r)

		Convey("Then the response should be 400 Not Found", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)

			var errResp models.ErrorList
			err := json.NewDecoder(w.Body).Decode(&errResp)
			So(err, ShouldBeNil)

			code := models.CodeNotFound
			expectedErrResp := models.ErrorList{
				Errors: []*models.Error{
					{
						Code:        &code,
						Description: apierrors.ErrorDescriptionNotFound,
					},
				},
			}
			So(errResp, ShouldResemble, expectedErrResp)
		})
	})
}
