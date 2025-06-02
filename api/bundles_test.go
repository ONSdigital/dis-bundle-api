package api

import (
	"context"
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

		bundleFilterFunc := func(ctx context.Context, offset, limit int, filters *filters.Bundlefilters) ([]*models.Bundle, int, error) {
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
					ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.Bundlefilters) ([]*models.Bundle, int, error) {
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
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.Bundlefilters) ([]*models.Bundle, int, error) {
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
			Convey("It should return a 404 error", func() {
				paramValue := time.Now().Format(time.RFC3339)

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
	})
}

func TestGetBundles_Failure(t *testing.T) {
	t.Parallel()
	Convey("Given a GET request to /bundles", t, func() {
		Convey("When the datastore returns an internal error", func() {
			r := httptest.NewRequest("GET", "http://localhost:29800/bundles", http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.Bundlefilters) ([]*models.Bundle, int, error) {
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
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.Bundlefilters) ([]*models.Bundle, int, error) {
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

		Convey("When an invalid publish_date is supplied", func() {
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bundles?%s=%s", filters.PublishDate, "notactuallyadate"), http.NoBody)
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.Bundlefilters) ([]*models.Bundle, int, error) {
					return nil, 0, nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)
			Convey("Then the status code should be 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				expectedResult := fmt.Sprintf(`{"code":"internal_server_error","description":"%s"}`+"\n", errs.ErrorDescriptionMalformedRequest)
				So(w.Body.String(), ShouldEqual, expectedResult)
			})
		})
	})
}
