package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var testEvent = &models.Event{
	Action:   "CREATE",
	Resource: "/bundles/test-bundle",
}

func TestGetBundleEvents_Success(t *testing.T) {
	Convey("Given a successful request with no query parameters", t, func() {
		mockDatastore := &storetest.StorerMock{
			ListBundleEventsFunc: func(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
				return []*models.Event{testEvent}, 1, nil
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachine := &application.StateMachine{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			events, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should return events successfully", func() {
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 1)
				So(events, ShouldResemble, []*models.Event{testEvent})
			})
		})
	})
}

func TestGetBundleEvents_WithBundleFilter(t *testing.T) {
	Convey("Given a request with bundle ID filter", t, func() {
		mockDatastore := &storetest.StorerMock{
			ListBundleEventsFunc: func(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
				So(bundleID, ShouldEqual, "test-bundle")
				return []*models.Event{testEvent}, 1, nil
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachine := &application.StateMachine{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events?bundle=test-bundle", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			events, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should filter by bundle ID", func() {
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 1)
				So(events, ShouldResemble, []*models.Event{testEvent})
			})
		})
	})
}

func TestGetBundleEvents_WithDateFilter(t *testing.T) {
	Convey("Given a request with valid date filters", t, func() {
		mockDatastore := &storetest.StorerMock{
			ListBundleEventsFunc: func(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
				So(after, ShouldNotBeNil)
				So(before, ShouldNotBeNil)
				So(after.Year(), ShouldEqual, 2025)
				So(after.Month(), ShouldEqual, 1)
				So(after.Day(), ShouldEqual, 1)
				So(before.Year(), ShouldEqual, 2025)
				So(before.Month(), ShouldEqual, 12)
				So(before.Day(), ShouldEqual, 31)
				return []*models.Event{testEvent}, 1, nil
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachine := &application.StateMachine{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events?after=2025-01-01T00:00:00Z&before=2025-12-31T23:59:59Z", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			events, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should filter by date range", func() {
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 1)
				So(events, ShouldResemble, []*models.Event{testEvent})
			})
		})
	})
}

func TestGetBundleEvents_InvalidDateFormat(t *testing.T) {
	Convey("Given a request with invalid date format", t, func() {
		mockDatastore := &storetest.StorerMock{}
		stateMachine := &application.StateMachine{}
		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events?after=invalid-date", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			_, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should handle the error gracefully", func() {
				So(err, ShouldNotBeNil)
				So(totalCount, ShouldEqual, 0)
				So(w.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestGetBundleEvents_UnknownParameter(t *testing.T) {
	Convey("Given a request with unknown query parameter", t, func() {
		mockDatastore := &storetest.StorerMock{}
		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachine := &application.StateMachine{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events?invalid=test", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			_, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should return a 400 error", func() {
				So(err, ShouldNotBeNil)
				So(totalCount, ShouldEqual, 0)
				So(w.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestGetBundleEvents_InternalError(t *testing.T) {
	Convey("Given a request that causes an internal error", t, func() {
		mockDatastore := &storetest.StorerMock{
			ListBundleEventsFunc: func(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachine := &application.StateMachine{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			events, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should handle the internal error", func() {
				So(err, ShouldBeNil)
				So(totalCount, ShouldEqual, 0)
				So(events, ShouldBeNil)
				So(w.Code, ShouldEqual, 500)
			})
		})
	})
}

func TestGetBundleEvents_NoResults(t *testing.T) {
	Convey("Given a request with no results and no filters", t, func() {
		mockDatastore := &storetest.StorerMock{
			ListBundleEventsFunc: func(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
				return []*models.Event{}, 0, nil
			},
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
		stateMachine := &application.StateMachine{}
		stateMachineBundleAPI := application.Setup(store.Datastore{Backend: mockDatastore}, stateMachine, mockDatasetAPIClient)

		api := &BundleAPI{
			stateMachineBundleAPI: stateMachineBundleAPI,
		}

		req := httptest.NewRequest("GET", "/bundle-events", http.NoBody)
		w := httptest.NewRecorder()

		Convey("When getBundleEvents is called", func() {
			_, totalCount, err := api.getBundleEvents(w, req, 20, 0)

			Convey("Then it should return a 404 error", func() {
				So(err, ShouldNotBeNil)
				So(totalCount, ShouldEqual, 0)
				So(w.Code, ShouldEqual, 404)
			})
		})
	})
}
