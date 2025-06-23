package pagination

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	"github.com/ONSdigital/log.go/v2/log"
)

// TPaginatedHandler is a func type for an endpoint that returns a list of values that we want to paginate
type TPaginatedHandler[TItem any] func(w http.ResponseWriter, r *http.Request, limit int, offset int) (successResult *models.PaginationSuccessResult[TItem], errorResult *models.ErrorResult[models.Error])
type PaginatedHandler func(w http.ResponseWriter, r *http.Request, limit int, offset int) (items any, totalCount int, eventErrors *models.Error)

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
			err = errors.New("invalid query parameter: offset")
			log.Error(r.Context(), "invalid query parameter: offset", err, logData)
			return 0, 0, err
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			err = errors.New("invalid query parameter: limit")
			log.Error(r.Context(), "invalid query parameter: limit", err, logData)
			return 0, 0, err
		}
	}

	if limit > p.DefaultMaxLimit {
		logData["max_limit"] = p.DefaultMaxLimit
		err = errors.New("invalid query parameter: max_limit")
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
func Paginate[TItem any](p *Paginator, paginatedHandler TPaginatedHandler[TItem]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		offset, limit, err := p.getPaginationParameters(r)
		if err != nil {
			p.handlePaginationError(w, r, err)
			return
		}

		successResult, errorResult := paginatedHandler(w, r, limit, offset)
		if errorResult != nil {
			utils.HandleBundleAPIErr(w, r, errorResult.HTTPStatusCode, errorResult.Error)
			return
		}

		renderedPage := renderPage(successResult.Result.Items, offset, limit, successResult.Result.TotalCount)

		returnPaginatedResults(w, r, renderedPage)
	}
}

func (p *Paginator) Paginate(paginatedHandler PaginatedHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		offset, limit, err := p.getPaginationParameters(r)
		if err != nil {
			p.handlePaginationError(w, r, err)
			return
		}

		items, totalCount, requestError := paginatedHandler(w, r, limit, offset)
		if requestError != nil {
			if w.Header().Get("Content-Type") == "" {
				utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, requestError)
			}
			return
		}

		renderedPage := renderPage(items, offset, limit, totalCount)
		returnPaginatedResults(w, r, renderedPage)
	}
}

func (p *Paginator) handlePaginationError(w http.ResponseWriter, r *http.Request, err error) {
	log.Error(r.Context(), "pagination parameters incorrect", err)

	errArray := strings.Split(err.Error(), ":")
	param := errArray[len(errArray)-1]
	code := models.CodeBadRequest

	errInfo := &models.Error{
		Code:        &code,
		Description: "Unable to process request due to a malformed or invalid request body or query parameter",
		Source:      &models.Source{Parameter: param},
	}

	utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
}

func returnPaginatedResults(w http.ResponseWriter, r *http.Request, list PaginatedResponse) {
	logData := log.Data{"path": r.URL.Path, "method": r.Method}
	b, err := json.Marshal(list)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Error(r.Context(), "api endpoint failed to marshal resource into bytes", err, logData)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to process the request due to an internal error",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadGateway, errInfo)
		return
	}

	etag := dpresponse.GenerateETag(b, false)
	dpresponse.SetETag(w, etag)

	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		log.Error(r.Context(), "api endpoint error writing response body", err, logData)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to process the request due to an internal error",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadGateway, errInfo)
		return
	}

	log.Info(r.Context(), "api endpoint request successful", logData)
}
