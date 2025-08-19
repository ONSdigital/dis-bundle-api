package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"

	apiError "github.com/ONSdigital/dis-bundle-api/sdk/errors"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
)

const (
	service = "dis-bundle-api"
)

type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given bundle api url
func New(bundleAPIURL string) *Client {
	return &Client{
		hcCli: healthcheck.NewClient(service, bundleAPIURL),
	}
}

// NewWithHealthClient creates a new instance of bundle API Client,
// reusing the URL and Clienter from the provided healthcheck client
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		hcCli: healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (cli *Client) URL() string {
	return cli.hcCli.URL
}

// Health returns the underlying Healthcheck Client for this bundle API client
func (cli *Client) Health() *healthcheck.Client {
	return cli.hcCli
}

// Checker calls identity api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

type ResponseInfo struct {
	Body    []byte
	Headers http.Header
	Status  int
}

// callBundleAPI calls the Bundle API endpoint given by path for the provided REST method, request headers, and body payload.
// It returns the response body and any error that occurred.
func (cli *Client) callBundleAPI(ctx context.Context, path, method string, headers Headers, payload []byte) (*ResponseInfo, apiError.Error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to parse path: \"%v\" error is: %v", path, err),
		}
	}

	path = URL.String()

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload))
	} else {
		req, err = http.NewRequest(method, path, http.NoBody)
	}

	if err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to create request for call to bundle api, error is: %v", err),
		}
	}

	if payload != nil {
		req.Header.Add("Content-type", "application/json")
	}

	if headers.IfMatch != "" {
		req.Header.Add("If-Match", headers.IfMatch)
	}
	headers.Add(req)

	resp, err := cli.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to call bundle api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}
	defer func() {
		err = closeResponseBody(resp)
	}()

	respInfo := &ResponseInfo{
		Headers: resp.Header.Clone(),
		Status:  resp.StatusCode,
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 400 {
		return respInfo, apiError.StatusError{
			Err:  fmt.Errorf("failed as unexpected code from bundle api: %v", resp.StatusCode),
			Code: resp.StatusCode,
		}
	}

	if resp.Body == nil {
		return respInfo, nil
	}

	respInfo.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return respInfo, apiError.StatusError{
			Err:  fmt.Errorf("failed to read response body from call to bundle api, error is: %v", err),
			Code: resp.StatusCode,
		}
	}

	return respInfo, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(resp *http.Response) apiError.Error {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			return apiError.StatusError{
				Err:  fmt.Errorf("error closing http response body from call to bundle api, error is: %v", err),
				Code: http.StatusInternalServerError,
			}
		}
	}

	return nil
}
