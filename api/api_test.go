package api

import (
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/store"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		dataStore := &store.Datastore{}
		api := GetBundleAPIWithMocks(*dataStore, &datasetAPISDKMock.ClienterMock{}, &permissionsAPISDKMock.ClienterMock{}, false)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/bundles", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundles", "POST"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundles/{bundle-id}", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundles/{bundle-id}", "DELETE"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundles/{bundle-id}/contents", "POST"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundles/{bundle-id}/contents/{content-id}", "DELETE"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundle-events", "GET"), ShouldBeTrue)

			So(hasRoute(api.Router, "/bundles/{bundle-id}/state", "PUT"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	var found bool
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		routePath, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		routeMethods, err := route.GetMethods()
		if err != nil {
			routeMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
		}
		if routePath == path {
			for _, m := range routeMethods {
				if m == method {
					found = true
					break
				}
			}
		}
		return nil
	})
	return found
}
