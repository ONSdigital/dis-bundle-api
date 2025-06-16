package auth

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	authorisationmock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-permissions-api/sdk"

	. "github.com/smartystreets/goconvey/convey"
)

func createTestHTTPRequest(authKey *string, isBearer bool) http.Request {
	httpRequest := http.Request{
		Header: http.Header{},
	}

	if authKey != nil {
		authHeaderValue := *authKey
		if isBearer {
			authHeaderValue = fmt.Sprintf("%s%s", TokenBearerPrefix, *authKey)
		}
		httpRequest.Header.Set(utils.HeaderAuthorization, authHeaderValue)
	}
	return httpRequest
}

func TestGetBearerTokenValue(t *testing.T) {
	authKey := "testing-auth-key-1234"

	Convey("When the authorisation header exists", t, func() {
		httpRequest := createTestHTTPRequest(&authKey, true)
		result := getBearerTokenValue(&httpRequest)

		Convey("The auth header value should be returned without the bearer prefix", func() {
			So(result, ShouldEqual, authKey)
		})
	})

	Convey("When the authorisation header does not exist", t, func() {
		httpRequest := createTestHTTPRequest(nil, false)
		result := getBearerTokenValue(&httpRequest)

		Convey("An empty value should be returned", func() {
			So(result, ShouldEqual, "")
		})
	})

	Convey("When the authorisation header does not start with Bearer", t, func() {
		httpRequest := createTestHTTPRequest(&authKey, false)
		result := getBearerTokenValue(&httpRequest)

		Convey("It should return an empty string", func() {
			So(result, ShouldEqual, "")
		})
	})
}

func TestGetJWTEntityData(t *testing.T) {
	testAuthKey := "some-valid-key"
	Convey("When GetJWTEntityData is called with a Bearer token Authorisation header", t, func() {
		middleware := AuthMiddleware{
			Middleware: &authorisationmock.MiddlewareMock{
				ParseFunc: func(token string) (*sdk.EntityData, error) {
					return &sdk.EntityData{
						UserID: token,
					}, nil
				},
			},
		}

		request := createTestHTTPRequest(&testAuthKey, true)

		result, err := middleware.GetJWTEntityData(&request)

		Convey("Should return an instance of sdk.EntityData with UserID", func() {
			So(result.UserID, ShouldEqual, testAuthKey)
			So(err, ShouldBeNil)
		})
	})

	Convey("When GetJWTEntityData is called with an invalid Authorisation header", t, func() {
		middleware := AuthMiddleware{
			Middleware: &authorisationmock.MiddlewareMock{
				ParseFunc: func(token string) (*sdk.EntityData, error) {
					return &sdk.EntityData{
						UserID: token,
					}, nil
				},
			},
		}

		request := createTestHTTPRequest(&testAuthKey, false)

		result, err := middleware.GetJWTEntityData(&request)

		Convey("Should return an error message", func() {
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionNoTokenFound)
		})
	})

	Convey("When GetJWTEntityData is called and the Parse method returns an error", t, func() {
		middleware := AuthMiddleware{
			Middleware: &authorisationmock.MiddlewareMock{
				ParseFunc: func(token string) (*sdk.EntityData, error) {
					return nil, errors.New("some error message")
				},
			},
		}

		request := createTestHTTPRequest(&testAuthKey, true)

		result, err := middleware.GetJWTEntityData(&request)

		Convey("It should return an error message", func() {
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeInternalServerError)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionUserIdentityParseFailed)
		})
	})
}
