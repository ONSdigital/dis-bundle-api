package pagination

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// PaginatedHandler is a func type for an endpoint that returns a list of values that we want to paginate
type PaginatedHandler func(w http.ResponseWriter, r *http.Request, limit int, offset int) (list interface{}, totalCount int, errBundles *models.Error)

type PaginatedResponse struct {
	Items                   interface{} `json:"items"`
	models.PaginationFields             // embedded, flattening fields into JSON
}

type Paginator struct {
	DefaultLimit    int
	DefaultOffset   int
	DefaultMaxLimit int
}

func NewPaginator(defaultLimit, defaultOffset, defaultMaxLimit int) *Paginator {
	return &Paginator{
		DefaultLimit:    defaultLimit,
		DefaultOffset:   defaultOffset,
		DefaultMaxLimit: defaultMaxLimit,
	}
}

func (p *Paginator) getPaginationParameters(r *http.Request) (offset, limit int, err error) {
	logData := log.Data{}
	offsetParameter := r.URL.Query().Get("offset")
	limitParameter := r.URL.Query().Get("limit")

	offset = p.DefaultOffset
	limit = p.DefaultLimit

	if offsetParameter != "" {
		logData["offset"] = offsetParameter
		offset, err = strconv.Atoi(offsetParameter)
		if err != nil || offset < 0 {
			err = errors.New("invalid query parameter")
			log.Error(r.Context(), "invalid query parameter: offset", err, logData)
			return 0, 0, err
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			err = errors.New("invalid query parameter")
			log.Error(r.Context(), "invalid query parameter: limit", err, logData)
			return 0, 0, err
		}
	}

	if limit > p.DefaultMaxLimit {
		logData["max_limit"] = p.DefaultMaxLimit
		err = errors.New("invalid query parameter")
		log.Error(r.Context(), "limit is greater than the maximum allowed", err, logData)
		return 0, 0, err
	}
	return offset, limit, err
}

func renderPage(list interface{}, offset, limit, totalCount int) PaginatedResponse {
	return PaginatedResponse{
		Items: list,
		PaginationFields: models.PaginationFields{
			Count:      listLength(list),
			Offset:     offset,
			Limit:      limit,
			TotalCount: totalCount,
		},
	}
}

func listLength(list interface{}) int {
	l := reflect.ValueOf(list)
	return l.Len()
}

// Paginate wraps a http endpoint to return a paginated list from the list returned by the provided function
func (p *Paginator) Paginate(paginatedHandler PaginatedHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		offset, limit, err := p.getPaginationParameters(r)
		if err != nil {
			log.Error(r.Context(), "pagination parameters incorrect", err)
			code := models.CodeInternalServerError
			handleErr(w, code, "Unable to process request due to a malformed or invalid request body or query parameter", http.StatusBadRequest)
			return
		}
		list, totalCount, errBundle := paginatedHandler(w, r, limit, offset)
		if errBundle != nil {
			fmt.Println(err)
			log.Error(r.Context(), "something went wrong", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			code := models.CodeInternalServerError
			newError := models.Error{
				Code:        &code,
				Description: "Error decoding json",
			}
			errBytes, err := json.Marshal(newError)
			if err != nil {
				fmt.Println("smething went wrong decoding errbundle")
				fmt.Println(err)
			}
			http.Error(w, string(errBytes), http.StatusBadGateway)
			return
		}

		renderedPage := renderPage(list, offset, limit, totalCount)

		returnPaginatedResults(w, r, renderedPage)
	}
}

func returnPaginatedResults(w http.ResponseWriter, r *http.Request, list PaginatedResponse) {
	logData := log.Data{"path": r.URL.Path, "method": r.Method}
	b, err := json.Marshal(list)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Error(r.Context(), "api endpoint failed to marshal resource into bytes", err, logData)
		code := models.CodeInternalServerError
		handleErr(w, code, "internal error", http.StatusBadGateway)
		return
	}

	etag := dpresponse.GenerateETag(b, false)
	dpresponse.SetETag(w, etag)

	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		log.Error(r.Context(), "api endpoint error writing response body", err, logData)
		code := models.CodeInternalServerError
		handleErr(w, code, "internal error", http.StatusBadGateway)
		return
	}

	log.Info(r.Context(), "api endpoint request successful", logData)
}

func handleErr(w http.ResponseWriter, code models.Code, description string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	err := models.Error{Code: &code, Description: description}
	errBytes, errCheck := json.Marshal(err)
	if errCheck != nil {
		fmt.Println("api endpoint error writing response body")
		fmt.Println(err)
	}
	http.Error(w, string(errBytes), httpStatusCode)
}
