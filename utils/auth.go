package utils

import (
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dp-net/v3/request"
)

var (
	authHeaderTokens = []string{
		request.FlorenceHeaderKey,
		request.AuthHeaderKey,
	}
)

func GetServiceToken(r *http.Request) (*string, error) {
	for index := range authHeaderTokens {
		headerValue := r.Header.Get(authHeaderTokens[index])

		if headerValue != "" {
			return &headerValue, nil
		}
	}

	return nil, apierrors.ErrUnauthorised
}
