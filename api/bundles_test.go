package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestGetBundlesReturnsOK(t *testing.T) {
	t.Parallel()

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

	Convey("get bundles with default offset and limit", t, func() {
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
		So(ok, ShouldBeTrue)
		So(errResp, ShouldBeNil)
		So(actualBundles, ShouldResemble, defaultBundles)
		So(count, ShouldEqual, len(defaultBundles))
	})

	Convey("get bundles with custom offset and limit", t, func() {
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
		So(ok, ShouldBeTrue)
		So(err, ShouldBeNil)
		So(actualBundles, ShouldResemble, customBundles)
		So(count, ShouldEqual, len(customBundles))
	})

	Convey("get bundles with internal error from datastore", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:29800/bundles", http.NoBody)
		w := httptest.NewRecorder()

		mockedDatastore := &storetest.StorerMock{
			ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
				return nil, 0, errors.New("database failure")
			},
		}

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

		results, _, errResp := bundleAPI.getBundles(w, r, 10, 0)

		So(results, ShouldBeNil)
		So(errResp, ShouldNotBeNil)
		So(errResp.Description, ShouldEqual, "Failed to process the request due to an internal error")
	})
}
