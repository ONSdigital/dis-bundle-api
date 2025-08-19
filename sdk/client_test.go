package sdk

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"

	. "github.com/smartystreets/goconvey/convey"
)

const testHost = "http://localhost:23900"

var (
	initialTestState = healthcheck.CreateCheckState(service)
)

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(_ []string) {
			// This gets called by the mock, just don't do anything.
		},
		DoFunc: func(_ context.Context, _ *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newBundleAPIClient(_ *testing.T, httpClient *dphttp.ClienterMock) *Client {
	healthClient := healthcheck.NewClientWithClienter(service, testHost, httpClient)
	return NewWithHealthClient(healthClient)
}

func newBundleAPIClientWithoutClienter(_ *testing.T) *Client {
	return New(testHost)
}

func TestClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	timePriorHealthCheck := time.Now().UTC()

	Convey("Given clienter.Do returns success", t, func() {
		bundleAPIClient := newBundleAPIClientWithoutClienter(t)
		check := initialTestState

		Convey("When bundle API client Checker is called", func() {
			err := bundleAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusOK)
				So(check.StatusCode(), ShouldEqual, 200)
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
			})
		})

		Convey("When bundle API client URL is checked", func() {
			strurl := bundleAPIClient.URL()
			So(strurl, ShouldEqual, testHost)
		})

		Convey("When the health is checked", func() {
			healthResponse := bundleAPIClient.Health()
			So(healthResponse, ShouldNotBeNil)
		})
	})
}

func TestHealthCheckerClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	timePriorHealthCheck := time.Now().UTC()
	path := "/health"

	Convey("Given clienter.Do returns an error", t, func() {
		clientError := errors.New("unexpected error")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		bundleAPIClient := newBundleAPIClient(t, httpClient)
		check := initialTestState

		Convey("When bundle API client Checker is called", func() {
			err := bundleAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("Given a 500 response for health check", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		bundleAPIClient := newBundleAPIClient(t, httpClient)
		check := initialTestState

		Convey("When bundle API client Checker is called", func() {
			err := bundleAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+healthcheck.StatusMessage[health.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestCallBundleAPIErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given bundle API returns an error due to a malformed URL", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)

		responseInfo, err := bundleAPIClient.callBundleAPI(ctx, "git@[2001:db8::1]:repository.git", http.MethodGet, Headers{}, nil)
		So(err, ShouldNotBeNil)
		So(responseInfo, ShouldBeNil)
	})

	Convey("Given bundle API returns an error due to an incorect method", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)

		bundleAPIClient := newBundleAPIClient(t, httpClient)

		responseInfo, err := bundleAPIClient.callBundleAPI(ctx, "/bundles", "!@£$$££$£", Headers{}, nil)
		So(err, ShouldNotBeNil)
		So(responseInfo, ShouldBeNil)

	})

}
