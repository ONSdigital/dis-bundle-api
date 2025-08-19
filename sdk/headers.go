package sdk

import (
	"net/http"

	dprequest "github.com/ONSdigital/dp-net/v3/request"
)

type Headers struct {
	ServiceAuthToken string
	UserAccessToken  string
	IfMatch          string
}

func (h *Headers) Add(req *http.Request) {
	if h.ServiceAuthToken != "" {
		dprequest.AddServiceTokenHeader(req, h.ServiceAuthToken)
	}

	if h.UserAccessToken != "" {
		dprequest.AddFlorenceHeader(req, h.UserAccessToken)
	}
}
