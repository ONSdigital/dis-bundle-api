package events

import (
	"net/http"
	"testing"

	authmocks "github.com/ONSdigital/dis-bundle-api/auth/mocks"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	"github.com/ONSdigital/dp-permissions-api/sdk"
	. "github.com/smartystreets/goconvey/convey"
)

func CreateTestBundleEvents(mockEntityData *sdk.EntityData) *EventsManager {
	mockMongo := storetest.StorerMock{}
	mockDatastore := store.Datastore{
		Backend: &mockMongo,
	}

	mockAuthMiddleware := authmocks.AuthorisationMiddlewareMock{
		GetJWTEntityDataFunc: func(r *http.Request) (*sdk.EntityData, *models.Error) {
			if mockEntityData == nil {
				return nil, models.CreateModelError(models.CodeBadRequest, "error")
			}

			return mockEntityData, nil
		},
	}

	return &EventsManager{
		datastore:      mockDatastore,
		authMiddleware: &mockAuthMiddleware,
	}
}

func TestCreateBundleEventsManager(t *testing.T) {
	Convey("When valid ", t, func() {

	})
}
