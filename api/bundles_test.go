package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/filters"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsSDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func ptrBundleState(state models.BundleState) *models.BundleState {
	return &state
}

func newAuthMiddlwareMock() *authorisationMock.MiddlewareMock {
	return &authorisationMock.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
		ParseFunc: func(token string) (*permissionsSDK.EntityData, error) {
			if token == "test-auth-token" {
				return &permissionsSDK.EntityData{
					UserID: "User123",
				}, nil
			}
			return nil, errors.New("authorisation header not found")
		},
	}
}

func GetBundleAPIWithMocks(datastore store.Datastore) *BundleAPI {
	ctx := context.Background()
	cfg := &config.Config{}
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
		Datastore:    datastore,
		StateMachine: stateMachine,
	}
	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
	return Setup(ctx, cfg, r, &datastore, stateMachineBundleAPI, mockDatasetAPIClient, newAuthMiddlwareMock())
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
				State:         ptrBundleState(models.BundleStatePublished),
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
				State:         ptrBundleState(models.BundleStateDraft),
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

				bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

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

				bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

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

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

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

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

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

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
			successResp, errResp := bundleAPI.getBundles(w, r, 10, 0)
			Convey("Then the status code should be 500", func() {
				So(successResp, ShouldBeNil)
				So(errResp, ShouldNotBeNil)
				So(errResp.HTTPStatusCode, ShouldEqual, 500)
				So(errResp.Error.Description, ShouldEqual, "Failed to process the request due to an internal error")
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

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)
			Convey("Then the status code should be 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldEqual, `{"errors":[{"code":"internal_server_error","description":"Failed to process the request due to an internal error"}]}`+"\n")
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

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)
			Convey("Then the status code should be 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				expectedErrorCode := models.CodeInternalServerError
				expectedErrorSource := models.Source{
					Parameter: "publish_date",
				}

				expectedError := &models.Error{
					Code:        &expectedErrorCode,
					Description: errs.ErrorDescriptionMalformedRequest,
					Source:      &expectedErrorSource,
				}

				var errList models.ErrorList
				errList.Errors = append(errList.Errors, expectedError)
				bytes, err := json.Marshal(errList)
				if err != nil {
					fmt.Println(err)
					return
				}
				expectedErrorString := fmt.Sprintf("%s\n", string(bytes))
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
	state := models.BundleStateDraft

	validBundle := &models.Bundle{
		ID:          "bundle-2",
		Title:       "bundle-2",
		ETag:        "12345-etag",
		ManagedBy:   models.ManagedByWagtail,
		BundleType:  models.BundleTypeScheduled,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
		ScheduledAt: &scheduledAt,
		State:       &state,
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

	Convey("Given a GET /bundles/{bundle-id} request", t, func() {
		Convey("When the bundle-id is valid", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/valid-id", http.NoBody)
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return validBundle, nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore})
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
				So(*response.State, ShouldEqual, *validBundle.State)

				So(response.PreviewTeams, ShouldNotBeNil)
				So(len(*response.PreviewTeams), ShouldEqual, 1)
				So((*response.PreviewTeams)[0].ID, ShouldEqual, "c78d457e-98de-11ec-b909-0242ac120002")
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
					return nil, errs.ErrBundleNotFound
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore})
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
							Description: errs.ErrorDescriptionNotFound,
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

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore})
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 500 with structured InternalError", func() {
				So(rec.Code, ShouldEqual, http.StatusInternalServerError)

				var errResp models.ErrorList
				err := json.NewDecoder(rec.Body).Decode(&errResp)
				So(err, ShouldBeNil)

				code := models.CodeInternalServerError
				expectedErrResp := models.ErrorList{
					Errors: []*models.Error{
						{
							Code:        &code,
							Description: errs.ErrorDescriptionInternalError,
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

			bundleAPI := Setup(ctx, &config.Config{}, mux.NewRouter(), nil, nil, nil, authMiddleware)
			bundleAPI.Router.HandleFunc("/bundles/{bundle-id}", func(w http.ResponseWriter, r *http.Request) {}).Methods(http.MethodGet)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 401 with structured Unauthorised error", func() {
				So(rec.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}
