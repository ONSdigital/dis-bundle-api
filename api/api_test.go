package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// func TestSetup(t *testing.T) {
// 	Convey("Given an API instance", t, func() {
// 		r := mux.NewRouter()
// 		ctx := context.Background()
// 		store := store.Datastore{}
// 		mockStateMachine := &application.StateMachineBundleAPI{
// 			Datastore:    store,
// 			StateMachine: &application.StateMachine{},
// 		}
// 		api := Setup(ctx, r, &store, mockStateMachine)

// 		// TODO: remove hello world example handler route test case
// 		Convey("When created the following routes should have been added", func() {
// 			// Replace the check below with any newly added api endpoints
// 			So(hasRoute(api.Router, "/hello", "GET"), ShouldBeTrue)
// 		})
// 	})
// }

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
