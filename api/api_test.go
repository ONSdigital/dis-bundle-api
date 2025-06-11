package api

import (
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		store := &store.Datastore{}
		api := GetBundleAPIWithMocks(*store)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/bundles", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/bundles", "POST"), ShouldBeTrue)
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
