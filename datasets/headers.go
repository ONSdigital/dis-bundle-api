package datasets

import (
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

func CreateAuthHeaders(r *http.Request) datasetAPISDK.Headers {
	var authHeaders datasetAPISDK.Headers
	if r.Header.Get(utils.HeaderFlorenceToken) != "" {
		authHeaders.ServiceToken = r.Header.Get(utils.HeaderFlorenceToken)
	} else {
		authHeaders.ServiceToken = r.Header.Get(utils.HeaderAuthorization)
	}

	return authHeaders
}
