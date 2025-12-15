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

	dataStore := &store.Datastore{}

	api := GetBundleAPIWithMocksWithAuthMiddleware(*dataStore, &datasetAPISDKMock.ClienterMock{}, &mockAuthMiddleware, false)

	Convey("When GetAuthEntityData is called with a valid Authorization header", t, func() {
		authEntityData, err := api.GetAuthEntityData(&r)

		Convey("Then authEntityData should not be nil", func() {
			So(authEntityData, ShouldNotBeNil)

			Convey("And it should contain EntityData from the auth middleware", func() {
				So(authEntityData.EntityData, ShouldEqual, mockEntityData)
			})

			Convey("And it should contain the correct Headers", func() {
				So(authEntityData.Headers.AccessToken, ShouldEqual, BearerTokenValue)
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
			dataStore := &store.Datastore{}

			api := GetBundleAPIWithMocksWithAuthMiddleware(*dataStore, &datasetAPISDKMock.ClienterMock{}, &mockAuthMiddleware, true)

			entityData, err := api.GetAuthEntityData(&r)

			Convey("Then AuthEntityData should be nil", func() {
				So(entityData, ShouldBeNil)
			})

			Convey("And an error should be returned", func() {
				So(err.Error(), ShouldContainSubstring, "unable to determine the user or service making the request")
			})
		})

		Convey("And no valid auth token is supplied", func() {
			r := http.Request{
				Header: http.Header{},
			}

			dataStore := &store.Datastore{}

			api := GetBundleAPIWithMocksWithAuthMiddleware(*dataStore, &datasetAPISDKMock.ClienterMock{}, &mockAuthMiddleware, true)

			entityData, err := api.GetAuthEntityData(&r)

			Convey("Then AuthEntityData should be nil", func() {
				So(entityData, ShouldBeNil)
			})

			Convey("And an error should be returned", func() {
				So(err.Error(), ShouldContainSubstring, "Empty token provided")
			})
		})
	})
}
