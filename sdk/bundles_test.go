package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	now         = time.Now().UTC()
	oneDayLater = now.Add(24 * time.Hour)

	testBundle = models.Bundle{
		ID:            "bundle1",
		BundleType:    models.BundleTypeScheduled,
		CreatedBy:     &models.User{Email: "creator@example.com"},
		CreatedAt:     &now,
		LastUpdatedBy: &models.User{Email: "updater@example.com"},
		PreviewTeams:  []models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
		ScheduledAt:   &oneDayLater,
		State:         models.BundleStatePublished,
		Title:         "Scheduled Bundle 1",
		UpdatedAt:     &now,
		ManagedBy:     models.ManagedByDataAdmin,
	}

	testBundleItems = []models.Bundle{testBundle}

	testBundles = BundlesList{
		Count:  1,
		Items:  testBundleItems,
		Offset: 0,
		Limit:  20,
	}
)

func TestGetBundles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given bundle API returns successfully", t, func() {
		body, err := json.Marshal(testBundles)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)

		Convey("When GetBundles is called", func() {
			bundlesResponse, err := bundleAPIClient.GetBundles(ctx, Headers{}, oneDayLater)

			Convey("Then the expected bundles are returned", func() {
				So(*bundlesResponse, ShouldResemble, testBundles)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles")
					})
				})
			})
		})
	})

	Convey("When GetBundles is called with no response body returned", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			},
			nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)
		bundlesResponse, err := bundleAPIClient.GetBundles(ctx, Headers{}, oneDayLater)

		Convey("Then no bundles are returned", func() {
			So(bundlesResponse, ShouldBeNil)

			Convey("And no error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And client.Do should be called once with the expected parameters", func() {
					doCalls := httpClient.DoCalls()
					So(doCalls, ShouldHaveLength, 1)
					So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles")
				})
			})
		})
	})

	Convey("Given a 500 response from bundle api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		bundleAPIClient := newBundleAPIClient(t, httpClient)

		Convey("When GetBundles is called", func() {
			bundlesResponse, err := bundleAPIClient.GetBundles(ctx, Headers{}, oneDayLater)

			Convey("Then an error should be returned ", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusInternalServerError)

				Convey("And the expected bundle list should be empty", func() {
					So(bundlesResponse.Items, ShouldBeEmpty)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles")
					})
				})
			})
		})
	})
}

func TestGetBundle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given bundle API returns successfully", t, func() {
		body, err := json.Marshal(testBundle)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)

		Convey("When GetBundle is called with a response body populated", func() {
			responseInfo := ResponseInfo{Body: body, Status: 200, Headers: nil}
			bundleResponse, err := bundleAPIClient.GetBundle(ctx, Headers{}, "bundle1")

			Convey("Then the expected bundle is returned", func() {
				So(*bundleResponse, ShouldResemble, responseInfo)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles/bundle1")
					})
				})
			})
		})
	})

	Convey("When GetBundle is called with no response body returned", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			},
			nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)
		bundleResponse, err := bundleAPIClient.GetBundle(ctx, Headers{}, "bundle1")

		Convey("Then no bundle is returned", func() {
			So(bundleResponse, ShouldBeNil)

			Convey("And no error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And client.Do should be called once with the expected parameters", func() {
					doCalls := httpClient.DoCalls()
					So(doCalls, ShouldHaveLength, 1)
					So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles/bundle1")
				})
			})
		})
	})

	Convey("Given a 500 response from bundle api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		bundleAPIClient := newBundleAPIClient(t, httpClient)

		Convey("When GetBundle is called", func() {
			bundleResponse, err := bundleAPIClient.GetBundle(ctx, Headers{}, "bundle1")

			Convey("Then an error should be returned ", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusInternalServerError)

				Convey("And the expected bundle should be empty", func() {
					So(bundleResponse, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles/bundle1")
					})
				})
			})
		})
	})
}

func TestPutBundleState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given bundle API returns successfully", t, func() {
		body, err := json.Marshal(testBundle)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)

		headers := Headers{IfMatch: "12345566"}

		Convey("When PutBundleState is called", func() {
			bundleResponse, err := bundleAPIClient.PutBundleState(ctx, headers, "bundle1", models.BundleStatePublished)

			Convey("Then the expected bundle is returned", func() {
				So(*bundleResponse, ShouldResemble, testBundle)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles/bundle1/state")
					})
				})
			})
		})
	})

	Convey("When PutBundleState is called and responds with no response body", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			},
			nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)
		bundleResponse, err := bundleAPIClient.PutBundleState(ctx, Headers{}, "bundle1", "bob")

		Convey("Then no bundle is returned", func() {
			So(bundleResponse, ShouldBeNil)

			Convey("And no error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And client.Do should be called once with the expected parameters", func() {
					doCalls := httpClient.DoCalls()
					So(doCalls, ShouldHaveLength, 1)
					So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles/bundle1/state")
				})
			})
		})
	})

	Convey("When PutBundle is called with an invalid state", t, func() {
		statusError := errors.New("failed to create request for call to bundle api, error is")
		httpClient := newMockHTTPClient(&http.Response{}, statusError)

		bundleAPIClient := newBundleAPIClient(t, httpClient)
		bundleResponse, err := bundleAPIClient.PutBundleState(ctx, Headers{}, "bundle1", "invalidstate")

		Convey("Then no bundle is returned", func() {
			So(bundleResponse.ID, ShouldBeEmpty)

			Convey("And an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And client.Do should be called once with the expected parameters", func() {
					doCalls := httpClient.DoCalls()
					So(doCalls, ShouldHaveLength, 1)
					So(doCalls[0].Req.URL.Path, ShouldEqual, "/bundles/bundle1/state")
				})
			})
		})
	})
}
