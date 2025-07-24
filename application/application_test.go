package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPIMocks "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsSDK "github.com/ONSdigital/dp-permissions-api/sdk"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	today     = time.Now()
	yesterday = today.Add(-24 * time.Hour)
	tomorrow  = today.Add(24 * time.Hour)
)

var fullyPopulatedBundle = models.Bundle{
	ID:            "123",
	BundleType:    models.BundleTypeManual,
	CreatedBy:     &models.User{Email: "example@example.com"},
	CreatedAt:     &today,
	LastUpdatedBy: &models.User{Email: "example@example.com"},
	PreviewTeams:  []models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
	ScheduledAt:   &tomorrow,
	State:         models.BundleStateDraft,
	Title:         "Fully Populated Bundle",
	UpdatedAt:     &today,
	ManagedBy:     models.ManagedByWagtail,
	ETag:          "f9226b8eb338ac139b1c39d2bb69f5abad8bea09",
}

func TestListBundles(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedBundles := []*models.Bundle{
			{
				ID:         "bundle-123",
				Title:      "Example Bundle",
				CreatedAt:  &yesterday,
				UpdatedAt:  &today,
				BundleType: models.BundleTypeScheduled,
				State:      models.BundleStatePublished,
			},
		}

		mockedDatastore := &storetest.StorerMock{
			ListBundlesFunc: func(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
				return expectedBundles, len(expectedBundles), nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When ListBundles is called", func() {
			now := time.Now()
			filters := filters.BundleFilters{
				PublishDate: &now,
			}
			results, totalCount, err := stateMachine.ListBundles(ctx, 0, 10, &filters)

			Convey("Then it should return the expected bundles without error", func() {
				So(err, ShouldBeNil)
				So(results, ShouldResemble, expectedBundles)
				So(totalCount, ShouldEqual, len(expectedBundles))
			})
		})
	})
}

func TestGetBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedBundle := &models.Bundle{
			ID:         "bundle-123",
			Title:      "Example Bundle",
			CreatedAt:  &yesterday,
			UpdatedAt:  &today,
			BundleType: models.BundleTypeScheduled,
			State:      models.BundleStatePublished,
		}

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				if bundleID == expectedBundle.ID {
					return expectedBundle, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When GetBundle is called with a valid ID", func() {
			result, err := stateMachine.GetBundle(ctx, expectedBundle.ID)

			Convey("Then it should return the expected bundle without error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldResemble, expectedBundle)
			})
		})

		Convey("When GetBundle is called with an invalid ID", func() {
			result, err := stateMachine.GetBundle(ctx, "invalid-id")

			Convey("Then it should return an error and nil result", func() {
				So(err, ShouldEqual, apierrors.ErrBundleNotFound)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestUpdateBundleETag(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		oldBundle := &models.Bundle{
			ID:        "bundle-123",
			Title:     "Example Bundle",
			CreatedAt: &yesterday,
			UpdatedAt: &today,
			LastUpdatedBy: &models.User{
				Email: "old-email",
			},
			BundleType: models.BundleTypeScheduled,
			State:      models.BundleStatePublished,
			ETag:       "12345",
		}

		expectedBundle := &models.Bundle{
			ID:        oldBundle.ID,
			Title:     oldBundle.Title,
			CreatedAt: oldBundle.CreatedAt,
			UpdatedAt: &tomorrow,
			LastUpdatedBy: &models.User{
				Email: "new-email",
			},
			BundleType: oldBundle.BundleType,
			State:      oldBundle.State,
			ETag:       "new-etag",
		}

		mockedDatastore := &storetest.StorerMock{
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				if bundleID == oldBundle.ID {
					oldBundle.ETag = "new-etag"
					oldBundle.UpdatedAt = &tomorrow
					oldBundle.LastUpdatedBy.Email = email
					return oldBundle, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When UpdateBundleETag is called with a valid ID and email", func() {
			result, err := stateMachine.UpdateBundleETag(ctx, expectedBundle.ID, "new-email")

			Convey("Then it should return the updated bundle without error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldResemble, expectedBundle)
			})
		})
	})
}

func TestCheckBundleExists(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsFunc: func(ctx context.Context, bundleID string) (bool, error) {
				if bundleID == "existing-bundle" {
					return true, nil
				}
				return false, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CheckBundleExists is called with an existing ID", func() {
			exists, err := stateMachine.CheckBundleExists(ctx, "existing-bundle")

			Convey("Then it should return true without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
			})
		})

		Convey("When CheckBundleExists is called with a non-existing ID", func() {
			exists, err := stateMachine.CheckBundleExists(ctx, "non-existing-bundle")

			Convey("Then it should return false without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestCreateContentItem(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedContentItem := &models.ContentItem{
			ID:          "ContentItem1",
			BundleID:    "bundle1",
			ContentType: models.ContentTypeDataset,
			Metadata: models.Metadata{
				DatasetID: "dataset1",
				EditionID: "2025",
				VersionID: 1,
				Title:     "Test Dataset 1",
			},
			State: utils.PtrContentItemState(models.StateApproved),
			Links: models.Links{
				Edit:    "/edit/datasets/dataset1/editions/2025/versions/1",
				Preview: "/preview/datasets/dataset1/editions/2025/versions/1",
			},
		}

		mockedDatastore := &storetest.StorerMock{
			CreateContentItemFunc: func(ctx context.Context, item *models.ContentItem) error {
				if item.ID == expectedContentItem.ID {
					return nil
				}
				return errors.New("something failed")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateContentItem is called with a valid content item", func() {
			err := stateMachine.CreateContentItem(ctx, expectedContentItem)

			Convey("Then it should return no error", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When CreateContentItem is called with an invalid content item", func() {
			err := stateMachine.CreateContentItem(ctx, &models.ContentItem{ID: "simulated-error"})

			Convey("Then it should return an error", func() {
				So(err.Error(), ShouldEqual, "something failed")
			})
		})
	})
}

func TestCheckAllBundleContentsAreApproved(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
				if bundleID == "all-approved-bundle" {
					return true, nil
				}
				return false, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CheckAllBundleContentsAreApproved is called with a bundleID that has all content items approved", func() {
			result, err := stateMachine.CheckAllBundleContentsAreApproved(ctx, "all-approved-bundle")

			Convey("Then it should return true without error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeTrue)
			})
		})

		Convey("When CheckAllBundleContentsAreApproved is called with a bundleID that doesn't have all content items approved", func() {
			result, err := stateMachine.CheckAllBundleContentsAreApproved(ctx, "non-approved-bundle")

			Convey("Then it should return false without error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeFalse)
			})
		})
	})
}

func TestCheckContentItemExistsByDatasetEditionVersion(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			CheckContentItemExistsByDatasetEditionVersionFunc: func(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
				if datasetID == "dataset1" && editionID == "2025" && versionID == 1 {
					return true, nil
				}
				return false, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CheckContentItemExistsByDatasetEditionVersion is called with an existing content item", func() {
			exists, err := stateMachine.CheckContentItemExistsByDatasetEditionVersion(ctx, "dataset1", "2025", 1)

			Convey("Then it should return true without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
			})
		})

		Convey("When CheckContentItemExistsByDatasetEditionVersion is called with a non-existing content item", func() {
			exists, err := stateMachine.CheckContentItemExistsByDatasetEditionVersion(ctx, "dataset2", "2025", 1)

			Convey("Then it should return false without error", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestCreateEventFromBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateEventFromBundle is called with a valid bundle", func() {
			errorObject, err := stateMachine.CreateEventFromBundle(ctx, &fullyPopulatedBundle, "test@email.com", models.ActionCreate)

			Convey("Then it should return no error", func() {
				So(errorObject, ShouldBeNil)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCreateEventFromBundle_ConvertBundleToBundleEvent_Failure(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateEventFromBundle is called with a nil bundle", func() {
			errorObject, err := stateMachine.CreateEventFromBundle(ctx, nil, "test@email.com", models.ActionCreate)

			Convey("Then it should return an input bundle cannot be nil error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "input bundle cannot be nil")
			})

			Convey("And errorObject should be an internal server error", func() {
				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errorObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestCreateEventFromBundle_CreateBundleEvent_Failure(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create bundle event")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateEventFromBundle is called with a valid bundle", func() {
			errorObject, err := stateMachine.CreateEventFromBundle(ctx, &fullyPopulatedBundle, "test@email.com", models.ActionCreate)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create bundle event")
			})

			Convey("And errorObject should be an internal server error", func() {
				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errorObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestCreateBundleEvent(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedEvent := &models.Event{
			CreatedAt: &today,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				if event.CreatedAt != nil {
					return nil
				}
				return errors.New("failed to create event")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundleEvent is called with a valid event", func() {
			err := stateMachine.CreateBundleEvent(ctx, expectedEvent)

			Convey("Then it should return no error", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When CreateBundleEvent is called with and has an error", func() {
			err := stateMachine.CreateBundleEvent(ctx, &models.Event{})

			Convey("Then it should return an error", func() {
				So(err.Error(), ShouldEqual, "failed to create event")
			})
		})
	})
}

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			CreatedBy:   &models.User{Email: "user@example.com"},
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateDraft,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return false, nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return bundleToCreate, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it shouldn't return any errors", func() {
				So(err, ShouldBeNil)
				So(errObject, ShouldBeNil)
			})

			Convey("And it should return the created bundle with a status code of 201", func() {
				So(statusCode, ShouldEqual, 201)
				So(createdBundle, ShouldResemble, createdBundle)
			})
		})
	})
}

func TestCreateBundle_Failure_Transition(t *testing.T) {
	Convey("Given a valid bundle", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateApproved,
		}

		stateMachine := &application.StateMachineBundleAPI{}

		Convey("When CreateBundle is called and the bundle is not in state DRAFT", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return the expected error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "bundle state must be DRAFT when creating a new bundle")

				code := models.CodeBadRequest
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionStateNotAllowedToTransition,
				}
				So(errObject, ShouldNotBeNil)
				So(errObject, ShouldResemble, expectedErr)
			})

			Convey("And it should return a status code of 400 and nil for createdBundle", func() {
				So(statusCode, ShouldEqual, 400)
				So(createdBundle, ShouldBeNil)
			})
		})
	})
}

func TestCreateBundle_Failure_CheckBundleExistsByTitle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateDraft,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return true, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and a bundle with the same title already exists", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error indicating the bundle already exists", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, apierrors.ErrBundleTitleAlreadyExists)

				code := models.CodeConflict
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionBundleTitleAlreadyExist,
					Source: &models.Source{
						Field: "/title",
					},
				}
				So(errObject, ShouldNotBeNil)
				So(errObject, ShouldResemble, expectedErr)
			})

			Convey("And it should return a status code of 409 and nil for createdBundle", func() {
				So(statusCode, ShouldEqual, 409)
				So(createdBundle, ShouldBeNil)
			})
		})
	})
}

func TestCreateBundle_Failure_Backend_CheckBundleExistsByTitle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateDraft,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return false, errors.New("backend error")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the backend returns an error when checking for existing bundles", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error indicating the backend failure", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "backend error")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldNotBeNil)
				So(errObject, ShouldResemble, expectedErr)
			})

			Convey("And it should return a status code of 500 and nil for createdBundle", func() {
				So(statusCode, ShouldEqual, 500)
				So(createdBundle, ShouldBeNil)
			})
		})
	})
}

func TestCreateBundle_Failure_CreateBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateDraft,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return errors.New("failed to create bundle")
			},
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return false, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the datastore fails to create the bundle", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error indicating the failure", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create bundle")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldNotBeNil)
				So(errObject, ShouldResemble, expectedErr)
			})

			Convey("And it should return a status code of 500 and nil for createdBundle", func() {
				So(statusCode, ShouldEqual, 500)
				So(createdBundle, ShouldBeNil)
			})
		})
	})
}

func TestCreateBundle_Failure_CreateEventFromBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			CreatedBy:   &models.User{Email: "user@example.com"},
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateDraft,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return false, nil
			},
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return bundleToCreate, nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create event from bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the datastore fails to create an event from the bundle", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error indicating the failure", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create event from bundle")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldNotBeNil)
				So(errObject, ShouldResemble, expectedErr)
			})

			Convey("And it should return a status code of 500 and nil for createdBundle", func() {
				So(statusCode, ShouldEqual, 500)
				So(createdBundle, ShouldBeNil)
			})
		})
	})
}

func TestCreateBundle_Failure_GetBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       models.BundleStateDraft,
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return false, nil
			},
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return nil, errors.New("failed to retrieve created bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the datastore fails to retrieve the created bundle", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error indicating the failure", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to retrieve created bundle")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldNotBeNil)
				So(errObject, ShouldResemble, expectedErr)
			})

			Convey("And it should return a status code of 500 and nil for createdBundle", func() {
				So(statusCode, ShouldEqual, 500)
				So(createdBundle, ShouldBeNil)
			})
		})
	})
}

func TestCheckBundleExistsByTitle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedBundle := &models.Bundle{
			ID:    "bundle-123",
			Title: "Example Bundle",
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				if title == expectedBundle.Title {
					return true, nil
				}
				return false, apierrors.ErrBundleNotFound
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CheckBundleExistsByTitle is called and a bundle with the same title already exist", func() {
			bundleExist, err := stateMachine.CheckBundleExistsByTitle(ctx, expectedBundle.Title)

			Convey("Then it should return true without error", func() {
				So(err, ShouldBeNil)
				So(bundleExist, ShouldBeTrue)
			})
		})

		Convey("When CheckBundleExistsByTitle is called and a bundle with the same title does not exist", func() {
			bundleExist, err := stateMachine.CheckBundleExistsByTitle(ctx, "Nonexistent Bundle")

			Convey("Then it should return false without error", func() {
				So(bundleExist, ShouldBeFalse)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestCheckBundleExistsByTitle_Failure(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedBundle := &models.Bundle{
			ID:    "bundle-123",
			Title: "Example Bundle",
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsByTitleFunc: func(ctx context.Context, title string) (bool, error) {
				return false, apierrors.ErrBundleNotFound
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CheckBundleExistsByTitle is called and a bundle with the datastore returns an error", func() {
			bundleExist, err := stateMachine.CheckBundleExistsByTitle(ctx, expectedBundle.Title)

			Convey("Then it should return false and an error", func() {
				So(bundleExist, ShouldBeFalse)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, apierrors.ErrBundleNotFound.Error())
			})
		})
	})
}

func TestValidateScheduledAt_Success(t *testing.T) {
	Convey("Given a bundle with a valid ScheduledAt date", t, func() {
		bundle := &models.Bundle{
			ScheduledAt: &tomorrow,
		}

		Convey("When validateScheduledAt is called", func() {
			stateMachine := &application.StateMachineBundleAPI{}
			err := stateMachine.ValidateScheduledAt(bundle)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestValidateScheduledAt_Failure_ScheduledAtNotSet(t *testing.T) {
	Convey("Given a bundle with ScheduledAt not set", t, func() {
		bundle := &models.Bundle{
			BundleType: models.BundleTypeScheduled,
		}

		Convey("When validateScheduledAt is called", func() {
			stateMachine := &application.StateMachineBundleAPI{}
			err := stateMachine.ValidateScheduledAt(bundle)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "scheduled_at is required for scheduled bundles")
			})
		})
	})
}

func TestValidateScheduledAt_Failure_ScheduledAtSet(t *testing.T) {
	Convey("Given a bundle with ScheduledAt set for a manual bundle", t, func() {
		bundle := &models.Bundle{
			BundleType:  models.BundleTypeManual,
			ScheduledAt: &tomorrow,
		}

		Convey("When validateScheduledAt is called", func() {
			stateMachine := &application.StateMachineBundleAPI{}
			err := stateMachine.ValidateScheduledAt(bundle)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "scheduled_at should not be set for manual bundles")
			})
		})
	})
}

func TestValidateScheduledAt_Failure_ScheduledAtInThePast(t *testing.T) {
	Convey("Given a bundle with a ScheduledAt date in the past", t, func() {
		bundle := &models.Bundle{
			ScheduledAt: &yesterday,
		}

		Convey("When validateScheduledAt is called", func() {
			stateMachine := &application.StateMachineBundleAPI{}
			err := stateMachine.ValidateScheduledAt(bundle)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "scheduled_at cannot be in the past")
			})
		})
	})
}

func TestGetBundleContents(t *testing.T) {
	ctx := context.Background()
	authHeaders := datasetAPISDK.Headers{}

	Convey("Given a GetBundleContents application method", t, func() {
		mockedDatastore := &storetest.StorerMock{}
		mockDatasetAPI := &datasetAPIMocks.ClienterMock{}

		app := application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
		}

		Convey("When the bundle is not found", func() {
			mockedDatastore.GetBundleFunc = func(ctx context.Context, id string) (*models.Bundle, error) {
				return nil, errors.New("bundle not found")
			}

			items, total, err := app.GetBundleContents(ctx, "missing-bundle", 0, 10, authHeaders)

			So(err, ShouldNotBeNil)
			So(items, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("When the bundle is published", func() {
			state := models.BundleStatePublished
			mockedDatastore.GetBundleFunc = func(ctx context.Context, id string) (*models.Bundle, error) {
				t := time.Now()
				return &models.Bundle{
					ID:            id,
					BundleType:    models.BundleTypeScheduled,
					Title:         "Published Bundle",
					CreatedBy:     &models.User{Email: "creator@example.com"},
					CreatedAt:     &t,
					LastUpdatedBy: &models.User{Email: "updater@example.com"},
					PreviewTeams:  []models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
					State:         state,
					UpdatedAt:     &t,
				}, nil
			}

			mockedDatastore.ListBundleContentsFunc = func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return []*models.ContentItem{{
					ID:          "1",
					BundleID:    "bundle-1",
					ContentType: models.ContentTypeDataset,
					Metadata:    models.Metadata{DatasetID: "dataset-1"},
				}}, 1, nil
			}

			items, total, err := app.GetBundleContents(ctx, "bundle-1", 0, 10, authHeaders)

			So(err, ShouldBeNil)
			So(items, ShouldHaveLength, 1)
			So(total, ShouldEqual, 1)
		})

		Convey("When the bundle is unpublished and enrichment succeeds", func() {
			state := models.BundleStateDraft
			mockedDatastore.GetBundleFunc = func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    id,
					Title: "Unpublished Bundle",
					State: state,
				}, nil
			}

			mockedDatastore.ListBundleContentsFunc = func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return []*models.ContentItem{{
					ID:          "1",
					BundleID:    "bundle-1",
					ContentType: models.ContentTypeDataset,
					Metadata:    models.Metadata{DatasetID: "dataset-1"},
				}}, 1, nil
			}

			mockDatasetAPI.GetDatasetFunc = func(ctx context.Context, headers datasetAPISDK.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{
					ID:          datasetID,
					Title:       "Dataset Title",
					Description: "Description of dataset",
					Keywords:    []string{"economy", "gdp"},
					Publisher:   &datasetAPIModels.Publisher{Name: "ONS"},
					State:       "published",
				}, nil
			}

			items, total, err := app.GetBundleContents(ctx, "bundle-1", 0, 10, authHeaders)

			So(err, ShouldBeNil)
			So(items, ShouldHaveLength, 1)
			So(items[0].State.String(), ShouldEqual, "published")
			So(items[0].Metadata.Title, ShouldEqual, "Dataset Title")
			So(total, ShouldEqual, 1)
		})

		Convey("When enrichment fails for a dataset", func() {
			state := models.BundleStateDraft
			mockedDatastore.GetBundleFunc = func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    id,
					Title: "Draft Bundle",
					State: state,
				}, nil
			}

			mockedDatastore.ListBundleContentsFunc = func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return []*models.ContentItem{{
					ID:          "1",
					BundleID:    "bundle-1",
					ContentType: models.ContentTypeDataset,
					Metadata:    models.Metadata{DatasetID: "dataset-1"},
				}}, 1, nil
			}

			mockDatasetAPI.GetDatasetFunc = func(ctx context.Context, headers datasetAPISDK.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{}, errors.New("dataset fetch failed")
			}

			items, total, err := app.GetBundleContents(ctx, "bundle-1", 0, 10, authHeaders)

			So(err, ShouldNotBeNil)
			So(items, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("When bundle contents are empty", func() {
			state := models.BundleStateDraft
			mockedDatastore.GetBundleFunc = func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    id,
					Title: "Empty Bundle",
					State: state,
				}, nil
			}

			mockedDatastore.ListBundleContentsFunc = func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return []*models.ContentItem{}, 0, nil
			}

			items, total, err := app.GetBundleContents(ctx, "bundle-1", 0, 10, authHeaders)

			So(err, ShouldBeNil)
			So(items, ShouldBeEmpty)
			So(total, ShouldEqual, 0)
		})
	})
}

func TestDeleteBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				if id == "bundle-1" {
					return &models.Bundle{
						ID:    "bundle-1",
						State: models.BundleStateDraft,
					}, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID: "content-1",
					},
					{
						ID: "content-2",
					},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				if contentItemID == "content-1" || contentItemID == "content-2" {
					return nil
				}
				return errors.New("failed to delete content item")
			},
			DeleteBundleFunc: func(ctx context.Context, id string) error {
				return nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called with an existing bundle ID", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "example@email.com")

			Convey("Then it should return a status code of 204 and no error", func() {
				So(err, ShouldBeNil)
				So(errObject, ShouldBeNil)
				So(statusCode, ShouldEqual, 204)
			})
		})
	})
}

func TestDeleteBundle_Failure_GetBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				if id == "bundle-1" {
					return nil, apierrors.ErrBundleNotFound
				}
				return nil, errors.New("unexpected error")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called with a non-existing bundle ID", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 404 Not Found error", func() {
				So(statusCode, ShouldEqual, 404)
			})

			Convey("And the errorObject/error should indicate the bundle was not found", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, apierrors.ErrBundleNotFound.Error())

				code := models.CodeNotFound
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionNotFound,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})

		Convey("When DeleteBundle is called and the datastore returns an unexpected error", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "unexpected-error", "email@example.com")

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unexpected error")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_Transition(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    "bundle-1",
					State: models.BundleStatePublished,
				}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called with a bundle that is already published", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 409 Conflict error", func() {
				So(statusCode, ShouldEqual, 409)
			})

			Convey("And the errorObject/error should indicate the bundle cannot be deleted", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "cannot update a published bundle")

				code := models.CodeConflict
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionConflict,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_ListBundleContentIDsWithoutLimit(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    "bundle-1",
					State: models.BundleStateDraft,
				}, nil
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return nil, errors.New("failed to list bundle contents")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called and listing bundle contents fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to list bundle contents")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_DeleteContentItem(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    "bundle-1",
					State: models.BundleStateDraft,
				}, nil
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{ID: "content-1"},
					{ID: "content-2"},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return errors.New("failed to delete content item")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called and deleting a content item fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to delete content item")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_CreateEventFromContentItem(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    "bundle-1",
					State: models.BundleStateDraft,
				}, nil
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{ID: "content-1"},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create event from content item")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called and creating an event from content item fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create event from content item")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_DeleteBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    "bundle-1",
					State: models.BundleStateDraft,
				}, nil
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			DeleteBundleFunc: func(ctx context.Context, id string) error {
				return errors.New("failed to delete bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called and deleting the bundle fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to delete bundle")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_CreateEventFromBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    "bundle-1",
					State: models.BundleStateDraft,
				}, nil
			},
			ListBundleContentIDsWithoutLimitFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			DeleteBundleFunc: func(ctx context.Context, id string) error {
				return nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create event from bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called and creating an event from the bundle fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "bundle-1", "email@example.com")

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create event from bundle")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestPutBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := "bundle-123"
		userEmail := "user@example.com"

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
		}

		bundleUpdate := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
		}

		updatedBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			ETag:  "new-etag",
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsSDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				ServiceToken:    "test-token",
				UserAccessToken: "test-token",
			},
		}

		mockedDatastore := &storetest.StorerMock{
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			UpdateBundleETagFunc: func(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
				return updatedBundle, nil
			},
			CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When UpdateBundleWIP is called with no state change", func() {
			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, currentBundle, authEntityData)

			Convey("Then it should update the bundle successfully", func() {
				So(err, ShouldBeNil)
				So(result, ShouldEqual, updatedBundle)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.UpdateBundleETagCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.CreateBundleEventCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestValidateBundleRules_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		currentBundle := &models.Bundle{
			ID:    "bundle-123",
			Title: "Original Title",
		}

		bundleUpdate := &models.Bundle{
			ID:         "bundle-123",
			Title:      "New Title",
			BundleType: models.BundleTypeManual,
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				return false, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When ValidateBundleRules is called with valid data", func() {
			validationErrors := stateMachine.ValidateBundleRules(ctx, bundleUpdate, currentBundle)

			Convey("Then it should return no validation errors", func() {
				So(validationErrors, ShouldBeEmpty)
			})
		})
	})
}

func TestValidateBundleRules_DuplicateTitle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		currentBundle := &models.Bundle{
			ID:    "bundle-123",
			Title: "Original Title",
		}

		bundleUpdate := &models.Bundle{
			ID:         "bundle-123",
			Title:      "Existing Title",
			BundleType: models.BundleTypeManual,
		}

		mockedDatastore := &storetest.StorerMock{
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				if title == "Existing Title" {
					return true, nil
				}
				return false, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When ValidateBundleRules is called with duplicate title", func() {
			validationErrors := stateMachine.ValidateBundleRules(ctx, bundleUpdate, currentBundle)

			Convey("Then it should return a validation error for the title", func() {
				So(validationErrors, ShouldHaveLength, 1)
				So(*validationErrors[0].Code, ShouldEqual, models.CodeInvalidParameters)
				So(validationErrors[0].Source.Field, ShouldEqual, "/title")
			})
		})
	})
}

func TestValidateBundleRules_ScheduledValidation(t *testing.T) {
	Convey("Given a StateMachineBundleAPI", t, func() {
		ctx := context.Background()

		currentBundle := &models.Bundle{
			ID:    "bundle-123",
			Title: "Test Bundle",
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: &storetest.StorerMock{}},
		}

		Convey("When validating SCHEDULED bundle without scheduled_at", func() {
			bundleUpdate := &models.Bundle{
				ID:         "bundle-123",
				Title:      "Test Bundle",
				BundleType: models.BundleTypeScheduled,
			}

			validationErrors := stateMachine.ValidateBundleRules(ctx, bundleUpdate, currentBundle)

			Convey("Then it should return a validation error", func() {
				So(validationErrors, ShouldHaveLength, 1)
				So(*validationErrors[0].Code, ShouldEqual, models.CodeInvalidParameters)
				So(validationErrors[0].Source.Field, ShouldEqual, "/scheduled_at")
			})
		})

		Convey("When validating MANUAL bundle with scheduled_at", func() {
			futureTime := time.Now().Add(24 * time.Hour)
			bundleUpdate := &models.Bundle{
				ID:          "bundle-123",
				Title:       "Test Bundle",
				BundleType:  models.BundleTypeManual,
				ScheduledAt: &futureTime,
			}

			validationErrors := stateMachine.ValidateBundleRules(ctx, bundleUpdate, currentBundle)

			Convey("Then it should return a validation error", func() {
				So(validationErrors, ShouldHaveLength, 1)
				So(*validationErrors[0].Code, ShouldEqual, models.CodeInvalidParameters)
				So(validationErrors[0].Source.Field, ShouldEqual, "/scheduled_at")
			})
		})

		Convey("When validating SCHEDULED bundle with past scheduled_at", func() {
			pastTime := time.Now().Add(-24 * time.Hour)
			bundleUpdate := &models.Bundle{
				ID:          "bundle-123",
				Title:       "Test Bundle",
				BundleType:  models.BundleTypeScheduled,
				ScheduledAt: &pastTime,
			}

			validationErrors := stateMachine.ValidateBundleRules(ctx, bundleUpdate, currentBundle)

			Convey("Then it should return a validation error", func() {
				So(validationErrors, ShouldHaveLength, 1)
				So(*validationErrors[0].Code, ShouldEqual, models.CodeInvalidParameters)
				So(validationErrors[0].Source.Field, ShouldEqual, "/scheduled_at")
			})
		})
	})
}

func TestUpdateContentItemsWithDatasetInfo_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := "bundle-123"

		contentItems := []*models.ContentItem{
			{
				ID:       "content-1",
				BundleID: bundleID,
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					Title:     "Old Title",
				},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return contentItems, nil
			},
			UpdateContentItemDatasetInfoFunc: func(ctx context.Context, contentItemID, title, state string) error {
				return nil
			},
		}

		mockDatasetAPI := &datasetAPIMocks.ClienterMock{
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{
					ID:    datasetID,
					Title: "New Dataset Title",
					State: "published",
				}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
		}

		Convey("When UpdateContentItemsWithDatasetInfo is called", func() {
			authHeaders := datasetAPISDK.Headers{ServiceToken: "test-token"}
			err := stateMachine.UpdateContentItemsWithDatasetInfo(ctx, bundleID, authHeaders)

			Convey("Then it should update content items successfully", func() {
				So(err, ShouldBeNil)
				So(len(mockedDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.UpdateContentItemDatasetInfoCalls()), ShouldEqual, 1)
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].Title, ShouldEqual, "New Dataset Title")
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].State, ShouldEqual, "published")
			})
		})
	})
}

func TestUpdateContentItemsWithDatasetInfo_NoContentItems(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with no content items", t, func() {
		ctx := context.Background()
		bundleID := "bundle-123"

		mockedDatastore := &storetest.StorerMock{
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When UpdateContentItemsWithDatasetInfo is called", func() {
			authHeaders := datasetAPISDK.Headers{ServiceToken: "test-token"}
			err := stateMachine.UpdateContentItemsWithDatasetInfo(ctx, bundleID, authHeaders)

			Convey("Then it should complete without error", func() {
				So(err, ShouldBeNil)
				So(len(mockedDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestUpdateContentItemsWithDatasetInfo_DatasetAPIFailure(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with failing dataset API", t, func() {
		ctx := context.Background()
		bundleID := "bundle-123"

		contentItems := []*models.ContentItem{
			{
				ID:       "content-1",
				BundleID: bundleID,
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
				},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return contentItems, nil
			},
		}

		mockDatasetAPI := &datasetAPIMocks.ClienterMock{
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{}, errors.New("dataset API failure")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
		}

		Convey("When UpdateContentItemsWithDatasetInfo is called", func() {
			authHeaders := datasetAPISDK.Headers{ServiceToken: "test-token"}
			err := stateMachine.UpdateContentItemsWithDatasetInfo(ctx, bundleID, authHeaders)

			Convey("Then it should complete with error and log the failure", func() {
				So(err, ShouldNotBeNil)
				So(len(mockedDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.UpdateContentItemDatasetInfoCalls()), ShouldEqual, 0)
			})
		})
	})
}
