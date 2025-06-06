package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func ptrBundleState(state models.BundleState) *models.BundleState {
	return &state
}

func newAuthMiddlwareMock() *authorisation.MiddlewareMock {
	return &authorisation.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
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
	return Setup(ctx, cfg, r, &datastore, stateMachineBundleAPI, newAuthMiddlwareMock())
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

		Convey("When offset and limit values are default", func() {
			r := httptest.NewRequest("GET", "http://localhost:29800/bundles", http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
					return defaultBundles, len(defaultBundles), nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			results, count, errResp := bundleAPI.getBundles(w, r, 10, 0)
			actualBundles, ok := results.([]*models.Bundle)
			Convey("Then default values should be returned with no error", func() {
				So(ok, ShouldBeTrue)
				So(errResp, ShouldBeNil)
				So(actualBundles, ShouldResemble, defaultBundles)
				So(count, ShouldEqual, len(defaultBundles))
			})
		})

		Convey("When offset and limit values are custom", func() {
			r := httptest.NewRequest("GET", "http://localhost:29800/bundles?offset=1&limit=1", http.NoBody)
			w := httptest.NewRecorder()
			customBundles := defaultBundles[1:]

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
					So(offset, ShouldEqual, 1)
					So(limit, ShouldEqual, 1)
					return customBundles, len(customBundles), nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			results, count, err := bundleAPI.getBundles(w, r, 1, 1)
			actualBundles, ok := results.([]*models.Bundle)
			Convey("Then custom paginated values should be returned with no error", func() {
				So(ok, ShouldBeTrue)
				So(err, ShouldBeNil)
				So(actualBundles, ShouldResemble, customBundles)
				So(count, ShouldEqual, len(customBundles))
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
				ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
					return nil, 0, errors.New("database failure")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})
			results, errCode, errResp := bundleAPI.getBundles(w, r, 10, 0)
			Convey("Then the status code should be 500", func() {
				So(errCode, ShouldEqual, http.StatusInternalServerError)
				So(results, ShouldBeNil)
				So(results, ShouldBeNil)
				So(errResp, ShouldNotBeNil)
				So(errResp.Description, ShouldEqual, "Failed to process the request due to an internal error")
			})
		})

		Convey("When response returns an random internal error", func() {
			r := httptest.NewRequest(http.MethodGet, "/bundles", http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
					return nil, 0, errors.New("something broke inside")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)
			Convey("Then the status code should be 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldEqual, `{"code":"internal_server_error","description":"Failed to process the request due to an internal error"}`+"\n")
			})
		})
	})
}

func TestGetBundleById_Success(t *testing.T) {
	t.Parallel()

	validBundle := &models.Bundle{
		ID:         "valid-id",
		Title:      "Test Bundle",
		ETag:       "12345-etag",
		ManagedBy:  models.ManagedByDataAdmin,
		BundleType: models.BundleTypeManual,
	}

	Convey("Given a GET /bundles/{bundle-id} request", t, func() {
		Convey("When the bundle-id is valid", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/valid-id", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{"bundle_id": "valid-id"})
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleByIDFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
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
			})
		})
	})
}

func TestGetBundleById_Failure(t *testing.T) {
	t.Parallel()
	Convey("Given a GET /bundles/{bundle-id} request", t, func() {
		ctx := context.Background()
		Convey("When the bundle-id is invalid", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/invalid-id", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{"bundle_id": "invalid-id"})
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleByIDFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore})
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 404 with 'The requested resource does not exist'", func() {
				So(rec.Code, ShouldEqual, http.StatusNotFound)
				expected := `{"code":"not_found","description":"The requested resource does not exist"}` + "\n"
				So(rec.Body.String(), ShouldEqual, expected)
			})
		})

		Convey("When the request causes an internal error", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/valid-id", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{"bundle_id": "valid-id"})
			rec := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetBundleByIDFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
					return nil, errors.New("unexpected failure")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockStore})
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 500 with 'An internal error occurred'", func() {
				So(rec.Code, ShouldEqual, http.StatusInternalServerError)
				expected := `{"code":"internal_server_error","description":"An internal error occurred"}` + "\n"
				So(rec.Body.String(), ShouldEqual, expected)
			})
		})

		Convey("When no valid authentication token is provided", func() {
			req := httptest.NewRequest(http.MethodGet, "/bundles/protected-id", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{"bundle_id": "protected-id"})
			rec := httptest.NewRecorder()

			authMiddleware := &authorisation.MiddlewareMock{
				RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						http.Error(w, `{"code":"Unauthorised","description":"Access denied."}`, http.StatusUnauthorized)
					}
				},
			}

			bundleAPI := Setup(ctx, &config.Config{}, mux.NewRouter(), nil, nil, authMiddleware)
			bundleAPI.Router.HandleFunc("/bundles/{bundle_id}", func(w http.ResponseWriter, r *http.Request) {}).Methods(http.MethodGet)
			bundleAPI.Router.ServeHTTP(rec, req)

			Convey("Then the response should be 401 with 'Access denied.'", func() {
				So(rec.Code, ShouldEqual, http.StatusUnauthorized)
				So(rec.Body.String(), ShouldEqual, `{"code":"Unauthorised","description":"Access denied."}`+"\n")
			})
		})
	})
}
