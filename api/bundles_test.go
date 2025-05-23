package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetBundlesReturnsOK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("get bundles delegates offset and limit to db func and returns results list", t, func() {
		r := httptest.NewRequest("GET", "localhost:29800/bundles", http.NoBody)
		w := httptest.NewRecorder()
		now := time.Now()
		expectedBundles := []*models.Bundle{
			{
				ID:         "bundle1",
				BundleType: "scheduled",
				Contents: []models.BundleContent{
					{DatasetID: "dataset1", EditionID: "edition1", ItemID: "item1", State: "published", Title: "Dataset 1", URLPath: "/dataset1/edition1/item1"},
					{DatasetID: "dataset2", EditionID: "edition2", ItemID: "item2", State: "draft", Title: "Dataset 2", URLPath: "/dataset2/edition2/item2"},
				},
				CreatedDate:     now,
				LastUpdatedBy:   models.User{Email: "ABCD"},
				PreviewTeams:    []models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
				PublishDateTime: now.Add(24 * time.Hour),
				State:           "active",
				Title:           "Scheduled Bundle 1",
				UpdatedDate:     now,
				WagtailManaged:  false,
			},
			{
				ID:         "bundle2",
				BundleType: "manual",
				Contents: []models.BundleContent{
					{DatasetID: "dataset3", EditionID: "edition3", ItemID: "item3", State: "draft", Title: "Dataset 3", URLPath: "/dataset3/edition3/item3"},
				},
				CreatedDate:     now,
				LastUpdatedBy:   models.User{Email: "ABCD"},
				PreviewTeams:    []models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
				PublishDateTime: now.Add(48 * time.Hour),
				State:           "inactive",
				Title:           "Manual Bundle 2",
				UpdatedDate:     now,
				WagtailManaged:  true,
			},
		}

		mockedDatastore := &storetest.StorerMock{
			ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
				return expectedBundles, len(expectedBundles), nil
			},
		}
		permissions := getAuthorisationHandlerMock()
		BundleAPI := Setup(ctx, &config.Config{}, mux.NewRouter(), &store.DataStore{Backend: mockedDatastore}, permissions)

		resultsList, count, err := BundleAPI.getBundles(w, r, 10, 0)

		So(err, ShouldBeNil)
		So(resultsList, ShouldResemble, expectedBundles)
		So(count, ShouldEqual, len(expectedBundles))
		So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
		So(w.Header().Get("ETag"), ShouldNotBeEmpty)
	})
}
