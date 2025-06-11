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
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	permsdk "github.com/ONSdigital/dp-permissions-api/sdk"
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
		State:       ptrBundleState(models.BundleStateDraft),
		Title:       "Scheduled Bundle 1",
		ManagedBy:   models.ManagedByWagtail,
	}

	// missing closing quote for the "id" field
	invalidBundlesPayload = `{
	  "id": "bundle1,
	  "bundle_type": "SCHEDULED",
	  "created_by": {
		"email": "example@example.com"
	  },
	  "last_updated_by": {
		"email": "example@example.com"
	  },
	  "preview_teams": [
		{
		  "id": "team1"
		},
		{
		  "id": "team2"
		}
	  ],
	  "scheduled_at": "2125-06-05T07:00:00.000Z",
	  "state": "DRAFT",
	  "title": "Scheduled Bundle 1",
	  "managed_by": "WAGTAIL"
	}`

	// scheduled_at is invalid
	invalidTimeInBundlesPayload = `{
	  "id": "bundle1",
	  "bundle_type": "SCHEDULED",
	  "created_by": {
		"email": "example@example.com"
	  },
	  "last_updated_by": {
		"email": "example@example.com"
	  },
	  "preview_teams": [
		{
		  "id": "team1"
		},
		{
		  "id": "team2"
		}
	  ],
	  "scheduled_at": "2125-06-05T07:00:00.000",
	  "state": "DRAFT",
	  "title": "Scheduled Bundle 1",
	  "managed_by": "WAGTAIL"
	}`

	// payload with invalid state for creating a bundle
	payloadWithInvalidState = `{
	  "id": "bundle1",
	  "bundle_type": "SCHEDULED",
	  "created_by": {
		"email": "example@example.com"
	  },
	  "last_updated_by": {
		"email": "example@example.com"
	  },
	  "preview_teams": [
		{
		  "id": "team1"
		},
		{
		  "id": "team2"
		}
	  ],
	  "scheduled_at": "2125-06-05T07:00:00.000Z",
	  "state": "IN_REVIEW",
	  "title": "Scheduled Bundle 1",
	  "managed_by": "WAGTAIL"
	}`
)

func ptrBundleState(state models.BundleState) *models.BundleState {
	return &state
}

func newAuthMiddlwareMock() *authorisation.MiddlewareMock {
	return &authorisation.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
		ParseFunc: func(token string) (*permsdk.EntityData, error) {
			if token == "some.valid.token" {
				return &permsdk.EntityData{
					UserID: "some-user-id",
					Groups: []string{"some-group"},
				}, nil
			} else {
				return nil, errors.New("invalid token")
			}
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

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
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
				CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
					return nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

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
				So(createdBundle.ETag, ShouldEqual, "some-etag")
			})
			Convey("And the correct headers should be set", func() {
				So(w.Header().Get("Location"), ShouldEqual, "/bundles/bundle1")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
				So(w.Header().Get("ETag"), ShouldEqual, "some-etag")
			})
		})
	})
}

func TestCreateBundle_Failure_FailedToParseBody(t *testing.T) {
	Convey("Given an invalid payload", t, func() {
		b := invalidBundlesPayload

		Convey("When a POST request is made to /bundles endpoint with the invalid payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"invalid_parameters"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"`+apierrors.ErrDescription+`"`)
			})
		})
	})
}

func TestCreateBundle_Failure_InvalidScheduledAt(t *testing.T) {
	Convey("Given a payload with invalid scheduled_at format", t, func() {
		b := invalidTimeInBundlesPayload

		Convey("When a POST request is made to /bundles endpoint with the invalid payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"invalid_parameters"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"Invalid time format in request body"`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"scheduled_at"}`)
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

		bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

		bundleAPI.Router.ServeHTTP(w, r)

		Convey("Then the response should be 500 Internal Server Error with an error message", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.String(), ShouldContainSubstring, `"code":"internal_server_error"`)
			So(w.Body.String(), ShouldContainSubstring, `"description":"`+apierrors.ErrInternalErrorDescription+`"`)
		})
	})
}

func TestCreateBundle_Failure_ValidationError(t *testing.T) {
	Convey("Given a payload with missing mandatory fields", t, func() {
		bundle := `{
			"id": "",
			"bundle_type": "",
			"preview_teams": [],
			"title": "",
			"managed_by": ""
		}`

		Convey("When a POST request is made to /bundles endpoint with the invalid payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(bundle))
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with validation errors", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"invalid_parameters"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"`+apierrors.ErrDescription+`"`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/bundle_type"}`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/preview_teams"}`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/title"}`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/managed_by"}`)
			})
		})
	})
}

func TestCreateBundle_Failure_FailedToTransitionBundleState(t *testing.T) {
	Convey("Given a payload with invalid state for creating a bundle ", t, func() {
		b := payloadWithInvalidState

		Convey("When a POST request is made to /bundles endpoint with the payload", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"bad_request"`)
				So(w.Body.String(), ShouldContainSubstring, "Failed to transition bundle state")
			})
		})
	})
}

func TestCreateBundle_Failure_AuthTokenIsMissing(t *testing.T) {
	Convey("Given a payload for creating a bundle ", t, func() {
		b := payloadWithInvalidState

		Convey("When a POST request is made to /bundles endpoint with the payload and the auth token is missing", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			r.Header.Set("Authorization", "")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 401 Unauthorized with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, `"code":"unauthorized"`)
				So(w.Body.String(), ShouldContainSubstring, "Authorization token is required")
			})
		})
	})
}

func TestCreateBundle_Failure_AuthTokenIsInvalid(t *testing.T) {
	Convey("Given a payload for creating a bundle ", t, func() {
		b := payloadWithInvalidState

		Convey("When a POST request is made to /bundles endpoint with the payload and the auth token is invalid", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewBufferString(b))
			r.Header.Set("Authorization", "some.invalid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, `"code":"internal_server_error"`)
				So(w.Body.String(), ShouldContainSubstring, apierrors.ErrInternalErrorDescription)
			})
		})
	})
}

func TestCreateBundle_Failure_GetBundleByTitleFails(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and GetBundleByTitle fails", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, errors.New("failed to get bundle by title")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, `"code":"internal_server_error"`)
				So(w.Body.String(), ShouldContainSubstring, apierrors.ErrInternalErrorDescription)
			})
		})
	})
}

func TestCreateBundle_Failure_BundleWithSameTitleAlreadyExists(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and there is a bundle with the same title", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return &inputBundle, nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 409 conflict with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusConflict)
				So(w.Body.String(), ShouldContainSubstring, `"code":"conflict"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"A bundle with the same title already exists"`)
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
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return errors.New("failed to create bundle")
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, `"code":"internal_server_error"`)
				So(w.Body.String(), ShouldContainSubstring, apierrors.ErrInternalErrorDescription)
			})
		})
	})
}

func TestCreateBundle_Failure_CreateBundleEventReturnsAnError(t *testing.T) {
	Convey("Given a valid payload", t, func() {
		inputBundle := *validBundle
		inputBundleJSON, err := json.Marshal(inputBundle)
		So(err, ShouldBeNil)

		Convey("When a POST request is made to /bundles endpoint with the payload and CreateBundleEvent returns an error", func() {
			r := createRequestWithAuth(http.MethodPost, "/bundles", bytes.NewReader(inputBundleJSON))
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return nil
				},
				CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
					return errors.New("failed to create bundle event")
				},
				GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
					return &models.Bundle{}, nil
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 500 internal server error with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, `"code":"internal_server_error"`)
				So(w.Body.String(), ShouldContainSubstring, apierrors.ErrInternalErrorDescription)
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
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return apierrors.ErrScheduledAtRequired
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"bad_request"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"scheduled_at is required for scheduled bundles"`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/scheduled_at"}`)
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
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return apierrors.ErrScheduledAtSet
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"bad_request"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"scheduled_at should not be set for manual bundles"`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/scheduled_at"}`)
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
			r.Header.Set("Authorization", "Bearer some.valid.token")
			w := httptest.NewRecorder()

			mockedDatastore := &storetest.StorerMock{
				GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
					return nil, apierrors.ErrBundleNotFound
				},
				CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
					return apierrors.ErrScheduledAtInPast
				},
			}

			bundleAPI := GetBundleAPIWithMocks(store.Datastore{Backend: mockedDatastore})

			bundleAPI.Router.ServeHTTP(w, r)

			Convey("Then the response should be 400 Bad Request with an error message", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, `"code":"bad_request"`)
				So(w.Body.String(), ShouldContainSubstring, `"description":"scheduled_at cannot be in the past"`)
				So(w.Body.String(), ShouldContainSubstring, `"source":{"field":"/scheduled_at"}`)
			})
		})
	})
}
