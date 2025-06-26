package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/store"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	"github.com/ONSdigital/dp-permissions-api/sdk"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	EmptyString      = ""
	WhitespaceString = " "
	NotABearerToken  = "notabearertoken"
	BearerTokenValue = "abearertoken"
	BearerToken      = fmt.Sprintf("Bearer %s", BearerTokenValue)
)

func TestGetAuthEntityData_Success(t *testing.T) {
	mockUserID := "mock-user-id"
	mockEntityData := &sdk.EntityData{
		UserID: mockUserID,
	}
	mockAuthMiddleware := authorisationMock.MiddlewareMock{
		ParseFunc: func(token string) (*sdk.EntityData, error) {
			return mockEntityData, nil
		},
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
	}

	r := http.Request{
		Header: http.Header{},
	}

	r.Header.Set("Authorization", BearerToken)
	store := &store.Datastore{}

	api := GetBundleAPIWithMocksWithAuthMiddleware(*store, &datasetAPISDKMock.ClienterMock{}, &mockAuthMiddleware)

	Convey("When GetAuthEntityData is called with a valid authorization header", t, func() {
		entityData, err := api.GetAuthEntityData(&r)
		Convey("Then it should return an instance of AuthEntityData", func() {
			So(entityData, ShouldNotBeNil)

			Convey("With entity data from the auth middleware", func() {
				So(entityData.EntityData, ShouldEqual, mockEntityData)
			})

			Convey("And the bearer token should be returned", func() {
				So(entityData.ServiceToken, ShouldEqual, BearerTokenValue)
			})
		})

		Convey("And no error should be returned", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestGetAuthEntityData_Failure(t *testing.T) {
	Convey("When GetAuthEntityData is called", t, func() {
		mockAuthMiddleware := authorisationMock.MiddlewareMock{
			ParseFunc: func(token string) (*sdk.EntityData, error) {
				return nil, apierrors.ErrUnauthorised
			},
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
		}

		Convey("And the middleware parse func cannot find a valid auth token", func() {
			r := http.Request{
				Header: http.Header{},
			}

			r.Header.Set("Authorization", BearerToken)
			store := &store.Datastore{}

			api := GetBundleAPIWithMocksWithAuthMiddleware(*store, &datasetAPISDKMock.ClienterMock{}, &mockAuthMiddleware)

			entityData, err := api.GetAuthEntityData(&r)

			Convey("Then AuthEntityData should be nil", func() {
				So(entityData, ShouldBeNil)
			})

			Convey("And an error should be returned", func() {
				So(err, ShouldEqual, apierrors.ErrUnauthorised)
			})
		})

		Convey("And no valid auth token is supplied", func() {
			r := http.Request{
				Header: http.Header{},
			}

			store := &store.Datastore{}

			api := GetBundleAPIWithMocksWithAuthMiddleware(*store, &datasetAPISDKMock.ClienterMock{}, &mockAuthMiddleware)

			entityData, err := api.GetAuthEntityData(&r)

			Convey("Then AuthEntityData should be nil", func() {
				So(entityData, ShouldBeNil)
			})

			Convey("And an error should be returned", func() {
				So(err, ShouldEqual, apierrors.ErrUnauthorised)
			})
		})
	})
}

func TestGetBearerTokenValue(t *testing.T) {
	testCases := []struct {
		name           string
		headerValue    *string
		expectedResult *string
	}{
		{"no value", nil, &EmptyString},
		{"an empty value", &EmptyString, &EmptyString},
		{"a whitespace value", &WhitespaceString, nil},
		{"a value that isn't a bearer token", &NotABearerToken, nil},
		{"a value that is a bearer token", &BearerToken, &BearerTokenValue},
	}

	for index := range testCases {
		tc := testCases[index]
		testCaseName := fmt.Sprintf("When getBearerTokenValue is called with/%s", tc.name)

		t.Run(testCaseName, func(t *testing.T) {
			t.Parallel()
			r := http.Request{
				Header: http.Header{},
			}

			if tc.headerValue != nil {
				r.Header.Set("Authorization", *tc.headerValue)
			}
			token := getBearerTokenValue(&r)
			Convey("Then it should return the expected value", t, func() {
				So(token, ShouldEqual, tc.expectedResult)
			})
		})
	}
}
