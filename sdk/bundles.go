package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	apiError "github.com/ONSdigital/dis-bundle-api/sdk/errors"
)

// BundlesList represents an object containing a list of paginated bundles. This struct is based
// on the `pagination.page` struct which is returned when we call the `api.getBundles` endpoint
type BundlesList struct {
	Items      []models.Bundle `json:"items"`
	Count      int             `json:"count"`
	Offset     int             `json:"offset"`
	Limit      int             `json:"limit"`
	TotalCount int             `json:"total_count"`
}

// QueryParams represents the possible query parameters that a caller can provide
type QueryParams struct {
	Limit  int
	Offset int
}

// Validate validates tht no negative values are provided for limit or offset, and that the length of
// IDs is lower than the maximum
func (q *QueryParams) Validate() error {
	if q.Limit < 0 || q.Offset < 0 {
		return errors.New("negative offsets or limits are not allowed")
	}
	return nil
}

// GetBundles gets a list of bundles
func (cli *Client) GetBundles(ctx context.Context, headers Headers, scheduledAt *time.Time, queryParams *QueryParams) (*BundlesList, apiError.Error) {
	var bundlesList BundlesList
	path := fmt.Sprintf("%s/bundles", cli.hcCli.URL)
	if !scheduledAt.IsZero() {
		path += "?publish_date=" + scheduledAt.Format(time.RFC3339)
	}

	// Add query parameters to request if valid
	if queryParams != nil {
		if err := queryParams.Validate(); err != nil {
			return nil, apiError.StatusError{
				Code: 400,
				Err:  fmt.Errorf("failed to validate parameters, error is: %v", err),
			}
		}

		// Add query parameters
		query := url.Values{}
		query.Add("limit", strconv.Itoa(queryParams.Limit))
		query.Add("offset", strconv.Itoa(queryParams.Offset))

		path += query.Encode()
	}

	respInfo, apiErr := cli.callBundleAPI(ctx, path, http.MethodGet, headers, nil)
	if apiErr != nil {
		return &bundlesList, apiErr
	}

	if err := json.Unmarshal(respInfo.Body, &bundlesList); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal bundlesList response - error is: %v", err),
		}
	}

	return &bundlesList, nil
}

func (cli *Client) GetBundle(ctx context.Context, headers Headers, id string) (*ResponseInfo, apiError.Error) {
	path := fmt.Sprintf("%s/bundles/%s", cli.hcCli.URL, id)

	respInfo, apiErr := cli.callBundleAPI(ctx, path, http.MethodGet, headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var bundle models.Bundle
	if err := json.Unmarshal(respInfo.Body, &bundle); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal bundleResponse - error is: %v", err),
		}
	}

	return respInfo, nil
}

func (cli *Client) PutBundleState(ctx context.Context, headers Headers, id string, state models.BundleState) (*models.Bundle, apiError.Error) {
	path := fmt.Sprintf("%s/bundles/%s/state", cli.hcCli.URL, id)

	stateRequest := models.UpdateStateRequest{
		State: state,
	}

	b, _ := json.Marshal(stateRequest)

	respInfo, apiErr := cli.callBundleAPI(ctx, path, http.MethodPut, headers, b)
	if apiErr != nil {
		return &models.Bundle{}, apiErr
	}

	var bundle models.Bundle
	if err := json.Unmarshal(respInfo.Body, &bundle); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal bundle - error is: %v", err),
		}
	}

	return &bundle, nil
}
