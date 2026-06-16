package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/slack"
	slackMock "github.com/ONSdigital/dis-bundle-api/slack/mocks"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPIMocks "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	today     = time.Now()
	yesterday = today.Add(-24 * time.Hour)
	tomorrow  = today.Add(24 * time.Hour)
)

const (
	bundle1   = "bundle-1"
	bundle123 = "bundle-123"
	userEmail = "user@example.com"
)

var authEntityData = &models.AuthEntityData{
	EntityData: &permissionsAPISDK.EntityData{
		UserID: "test-user-id",
		Groups: []string{"group1", "group2"},
	},
}

func TestListBundles(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedBundles := []*models.Bundle{
			{
				ID:         bundle123,
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
			bundleFilters := filters.BundleFilters{
				PublishDate: &now,
			}
			results, totalCount, err := stateMachine.ListBundles(ctx, 0, 10, &bundleFilters)

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
			ID:         bundle123,
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
			ID:        bundle123,
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

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          bundle123,
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			CreatedBy:   &models.User{Email: userEmail},
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
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
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
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate, authEntityData)

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

func TestCreateBundle_Failure_CheckBundleExistsByTitle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          bundle123,
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
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate, authEntityData)

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
			ID:          bundle123,
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
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate, authEntityData)

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
			ID:          bundle123,
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
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate, authEntityData)

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

func TestCreateBundle_Failure_CreateEvent(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          bundle123,
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			CreatedBy:   &models.User{Email: userEmail},
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
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create event from bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the datastore fails to create an event", func() {
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate, authEntityData)

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
			ID:          bundle123,
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
			statusCode, createdBundle, errObject, err := stateMachine.CreateBundle(ctx, bundleToCreate, authEntityData)

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
			ID:    bundle123,
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
			ID:    bundle123,
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
					PreviewTeams:  &[]models.PreviewTeam{{ID: "team1"}, {ID: "team2"}},
					State:         state,
					UpdatedAt:     &t,
				}, nil
			}

			mockedDatastore.ListBundleContentsFunc = func(ctx context.Context, id string, offset, limit int) ([]*models.ContentItem, int, error) {
				return []*models.ContentItem{{
					ID:          "1",
					BundleID:    bundle1,
					ContentType: models.ContentTypeDataset,
					Metadata:    models.Metadata{DatasetID: "dataset-1"},
				}}, 1, nil
			}

			items, total, err := app.GetBundleContents(ctx, bundle1, 0, 10, authHeaders)

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
					BundleID:    bundle1,
					ContentType: models.ContentTypeDataset,
					Metadata:    models.Metadata{DatasetID: "dataset-1", EditionID: "edition-1", VersionID: 1},
				}}, 1, nil
			}

			mockDatasetAPI.GetDatasetFunc = func(ctx context.Context, headers datasetAPISDK.Headers, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{
					ID:          datasetID,
					Title:       "Dataset Title",
					Description: "Description of dataset",
					Keywords:    []string{"economy", "gdp"},
					Publisher:   &datasetAPIModels.Publisher{Name: "ONS"},
					State:       "published",
				}, nil
			}
			mockDatasetAPI.GetVersionFunc = func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
				return datasetAPIModels.Version{
					State: "associated",
				}, nil
			}

			items, total, err := app.GetBundleContents(ctx, bundle1, 0, 10, authHeaders)

			So(err, ShouldBeNil)
			So(items, ShouldHaveLength, 1)
			So(items[0].State.String(), ShouldEqual, "associated")
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
					BundleID:    bundle1,
					ContentType: models.ContentTypeDataset,
					Metadata:    models.Metadata{DatasetID: "dataset-1"},
				}}, 1, nil
			}

			mockDatasetAPI.GetDatasetFunc = func(ctx context.Context, headers datasetAPISDK.Headers, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{}, errors.New("dataset fetch failed")
			}

			items, total, err := app.GetBundleContents(ctx, bundle1, 0, 10, authHeaders)

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

			items, total, err := app.GetBundleContents(ctx, bundle1, 0, 10, authHeaders)

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
				if id == bundle1 {
					return &models.Bundle{
						ID:    bundle1,
						State: models.BundleStateDraft,
						PreviewTeams: &[]models.PreviewTeam{
							{ID: "preview-team-1"},
						},
					}, nil
				}
				return nil, apierrors.ErrBundleNotFound
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID:       "content-1",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-1",
							EditionID: "edition-1",
						},
					},
					{
						ID:       "content-2",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-2",
							EditionID: "edition-2",
						},
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
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}
		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return &permissionsAPIModels.Policy{
					ID: "preview-team-1",
					Condition: permissionsAPIModels.Condition{
						Values: []string{"dataset-1", "dataset-1/edition-1", "dataset-2", "dataset-2/edition-2"},
					},
				}, nil
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}
		stateMachine := &application.StateMachineBundleAPI{
			Datastore:            store.Datastore{Backend: mockedDatastore},
			PermissionsAPIClient: mockPermissionsClient,
		}

		Convey("When DeleteBundle is called with an existing bundle ID", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

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
				if id == bundle1 {
					return nil, apierrors.ErrBundleNotFound
				}
				return nil, errors.New("unexpected error")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called with a non-existing bundle ID", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

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
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, "unexpected-error", authEntityData)

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
					ID:    bundle1,
					State: models.BundleStatePublished,
				}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called with a bundle that is already published", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

			Convey("Then it should return a 409 Conflict error", func() {
				So(statusCode, ShouldEqual, 409)
			})

			Convey("And the errorObject/error should indicate the bundle cannot be deleted", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "cannot delete a published bundle")

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

func TestDeleteBundle_Failure_GetBundleContentsForBundle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    bundle1,
					State: models.BundleStateDraft,
				}, nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return nil, errors.New("failed to get bundle contents")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When DeleteBundle is called and listing bundle contents fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to get bundle contents")

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
					ID:    bundle1,
					State: models.BundleStateDraft,
				}, nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID:       "content-1",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-1",
							EditionID: "edition-1",
						},
					},
					{
						ID:       "content-2",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-2",
							EditionID: "edition-2",
						},
					},
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
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

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

func TestDeleteBundle_Failure_GetPolicy(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    bundle1,
					State: models.BundleStateDraft,
					PreviewTeams: &[]models.PreviewTeam{
						{ID: "preview-team-1"},
					},
				}, nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID:       "content-1",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-1",
							EditionID: "edition-1",
						},
					},
					{
						ID:       "content-2",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-2",
							EditionID: "edition-2",
						},
					},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return nil, errors.New("404 Not Found")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:            store.Datastore{Backend: mockedDatastore},
			PermissionsAPIClient: mockPermissionsClient,
		}

		Convey("When DeleteBundle is called and getting the policy associated with the bundle's preview team fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

			Convey("Then it should return a 404 Not Found", func() {
				So(statusCode, ShouldEqual, 404)
			})

			Convey("And the errorObject/error should indicate a Not Found error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "404 Not Found")

				code := models.CodeNotFound
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionNotFound,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})
	})
}

func TestDeleteBundle_Failure_PutPolicy(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    bundle1,
					State: models.BundleStateDraft,
					PreviewTeams: &[]models.PreviewTeam{
						{ID: "preview-team-1"},
					},
				}, nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID:       "content-1",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-1",
							EditionID: "edition-1",
						},
					},
					{
						ID:       "content-2",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-2",
							EditionID: "edition-2",
						},
					},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return &permissionsAPIModels.Policy{
					ID: "preview-team-1",
					Condition: permissionsAPIModels.Condition{
						Values: []string{"dataset-1", "dataset-1/edition-1", "dataset-2", "dataset-2/edition-2"},
					},
				}, nil
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return errors.New("failed to update policy")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:            store.Datastore{Backend: mockedDatastore},
			PermissionsAPIClient: mockPermissionsClient,
		}

		Convey("When DeleteBundle is called and updating the policy associated with the bundle's preview team fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to update policy")

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

func TestDeleteBundle_Failure_CreateEvent(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return &models.Bundle{
					ID:    bundle1,
					State: models.BundleStateDraft,
					PreviewTeams: &[]models.PreviewTeam{
						{ID: "preview-team-1"},
					},
				}, nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID:       "content-1",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-1",
							EditionID: "edition-1",
						},
					},
					{
						ID:       "content-2",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-2",
							EditionID: "edition-2",
						},
					},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return errors.New("failed to create event")
			},
		}
		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return &permissionsAPIModels.Policy{
					ID: "preview-team-1",
					Condition: permissionsAPIModels.Condition{
						Values: []string{"dataset-1", "dataset-1/edition-1", "dataset-2", "dataset-2/edition-2"},
					},
				}, nil
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}
		stateMachine := &application.StateMachineBundleAPI{
			Datastore:            store.Datastore{Backend: mockedDatastore},
			PermissionsAPIClient: mockPermissionsClient,
		}

		Convey("When DeleteBundle is called and creating an event for the content item fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create event")

				code := models.CodeInternalError
				expectedErr := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				So(errObject, ShouldResemble, expectedErr)
			})
		})

		Convey("When DeleteBundle is called and creating an event for the bundle fails", func() {
			mockedDatastore.CreateEventFunc = func(ctx context.Context, event *models.Event) error {
				if event.Bundle != nil {
					return errors.New("failed to create event")
				}
				return nil
			}
			mockedDatastore.DeleteBundleFunc = func(ctx context.Context, id string) error {
				return nil
			}

			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

			Convey("Then it should return a 500 Internal Server Error", func() {
				So(statusCode, ShouldEqual, 500)
			})

			Convey("And the errorObject/error should indicate an internal server error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to create event")

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
					ID:    bundle1,
					State: models.BundleStateDraft,
					PreviewTeams: &[]models.PreviewTeam{
						{ID: "preview-team-1"},
					},
				}, nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{
					{
						ID:       "content-1",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-1",
							EditionID: "edition-1",
						},
					},
					{
						ID:       "content-2",
						BundleID: bundle1,
						Metadata: models.Metadata{
							DatasetID: "dataset-2",
							EditionID: "edition-2",
						},
					},
				}, nil
			},
			DeleteContentItemFunc: func(ctx context.Context, contentItemID string) error {
				return nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			DeleteBundleFunc: func(ctx context.Context, id string) error {
				return errors.New("failed to delete bundle")
			},
		}
		mockPermissionsClient := &permissionsAPISDKMock.ClienterMock{
			GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
				return &permissionsAPIModels.Policy{
					ID: "preview-team-1",
					Condition: permissionsAPIModels.Condition{
						Values: []string{"dataset-1", "dataset-1/edition-1", "dataset-2", "dataset-2/edition-2"},
					},
				}, nil
			},
			PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
				return nil
			},
		}
		stateMachine := &application.StateMachineBundleAPI{
			Datastore:            store.Datastore{Backend: mockedDatastore},
			PermissionsAPIClient: mockPermissionsClient,
		}

		Convey("When DeleteBundle is called and deleting the bundle fails", func() {
			statusCode, errObject, err := stateMachine.DeleteBundle(ctx, bundle1, authEntityData)

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

func TestPutBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123
		userEmail := userEmail

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			ETag:  "old-etag",
		}

		bundleUpdate := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			ETag:  "new-etag",
		}

		updatedBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			ETag:  "new-etag",
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				AccessToken: "test-token",
			},
		}

		var states = make([]application.State, 0, 1)
		states = append(states, application.Draft)

		var transitions = make([]application.Transition, 0, 1)
		transitions = append(transitions, application.Transition{
			Label:               "DRAFT",
			TargetState:         application.Draft,
			AllowedSourceStates: []string{"IN_REVIEW", "DRAFT"},
		})

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return currentBundle, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:    store.Datastore{Backend: mockedDatastore},
			StateMachine: application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
		}

		Convey("When UpdateBundleWIP is called with no state change", func() {
			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, currentBundle.ETag)

			Convey("Then it should update the bundle successfully", func() {
				So(err, ShouldBeNil)
				So(result.ID, ShouldEqual, updatedBundle.ID)
				So(result.Title, ShouldEqual, updatedBundle.Title)
				So(result.State, ShouldEqual, updatedBundle.State)
				So(result.ETag, ShouldEqual, updatedBundle.ETag)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 1)
			})
		})
	})
}

func createMockVersionsAndContentItems(state models.BundleState) []*models.ContentItem {
	mockVersions := []*datasetAPIModels.Version{
		{
			ID:        "valid-version-1",
			Version:   1,
			DatasetID: "dataset-id-1",
			Edition:   "edition-id-1",
			State:     strings.ToLower(state.String()),
		},
		{
			ID:        "valid-version-2",
			Version:   1,
			DatasetID: "dataset-id-2",
			Edition:   "edition-id-2",
			State:     strings.ToLower(state.String()),
		},
	}

	mockContentItems := []*models.ContentItem{
		{
			ID:       "valid-content-item",
			BundleID: bundle123,
			State:    (*models.State)(&state),
			Metadata: models.Metadata{
				DatasetID: mockVersions[0].DatasetID,
				EditionID: mockVersions[0].Edition,
				VersionID: mockVersions[0].Version,
			},
		},
		{
			ID:       "another-valid-content-item",
			BundleID: bundle123,
			State:    (*models.State)(&state),
			Metadata: models.Metadata{
				DatasetID: mockVersions[1].DatasetID,
				EditionID: mockVersions[1].Edition,
				VersionID: mockVersions[1].Version,
			},
		},
	}

	return mockContentItems
}

func TestPutBundleState_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123
		userEmail := userEmail

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateApproved,
			ETag:  "old-etag",
		}

		bundleUpdate := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStatePublished,
			ETag:  "new-etag",
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				AccessToken: "test-token",
			},
		}

		var states = make([]application.State, 0, 1)
		states = append(states, application.Draft)

		var transitions = make([]application.Transition, 0, 1)
		transitions = append(transitions, application.Transition{
			Label:               "PUBLISHED",
			TargetState:         application.Published,
			AllowedSourceStates: []string{"APPROVED"},
		})

		mockContentItems := createMockVersionsAndContentItems(models.BundleStateApproved)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return currentBundle, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetBundleContentsForBundleFunc: func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
				contentItems := make([]models.ContentItem, len(mockContentItems))
				for index := range contentItems {
					contentItems[index] = *mockContentItems[index]
				}
				return &contentItems, nil
			},
			UpdateContentItemStateFunc: func(ctx context.Context, contentItemID, state string) error {
				return nil
			},
		}

		mockDatasetAPIClient := &datasetAPIMocks.ClienterMock{
			PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
				return nil
			},
		}

		mockSlackClient := &slackMock.ClienterMock{
			SendPublishLogFunc: func(ctx context.Context, summary string, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{
					ChannelID: "example-channel",
					Timestamp: "example-timestamp",
				}, nil
			},
			UpdatePublishLogFunc: func(ctx context.Context, ref *slack.MessageRef, summary string, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:             store.Datastore{Backend: mockedDatastore},
			StateMachine:          application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
			DataBundleSlackClient: mockSlackClient,
			DatasetAPIClient:      mockDatasetAPIClient,
		}

		Convey("When UpdateBundleState is called to publish a bundle which is a valid transition", func() {
			result, err := stateMachine.UpdateBundleState(ctx, bundleID, currentBundle.ETag, bundleUpdate.State, authEntityData)
			Convey("Then the bundle and it's content items should be published and slack alerts should be sent", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.State, ShouldEqual, models.BundleStatePublished)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 3)
				So(len(mockSlackClient.SendPublishLogCalls()), ShouldEqual, 1)
				So(len(mockSlackClient.UpdatePublishLogCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestPutBundleState_ContentItemFails(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123
		userEmail := userEmail

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateApproved,
			ETag:  "old-etag",
		}

		bundleUpdate := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStatePublished,
			ETag:  "new-etag",
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				AccessToken: "test-token",
			},
		}

		var states = make([]application.State, 0, 1)
		states = append(states, application.Draft)

		var transitions = make([]application.Transition, 0, 1)
		transitions = append(transitions, application.Transition{
			Label:               "PUBLISHED",
			TargetState:         application.Published,
			AllowedSourceStates: []string{"APPROVED"},
		})

		mockContentItems := createMockVersionsAndContentItems(models.BundleStateApproved)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return currentBundle, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetBundleContentsForBundleFunc: func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
				contentItems := make([]models.ContentItem, len(mockContentItems))
				for index := range contentItems {
					contentItems[index] = *mockContentItems[index]
				}
				return &contentItems, nil
			},
			UpdateContentItemStateFunc: func(ctx context.Context, contentItemID, state string) error {
				return nil
			},
		}

		mockDatasetAPIClient := &datasetAPIMocks.ClienterMock{
			PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
				return errors.New("state not allowed to transition")
			},
		}

		mockSlackClient := &slackMock.ClienterMock{
			SendPublishLogFunc: func(ctx context.Context, summary string, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{
					ChannelID: "example-channel",
					Timestamp: "example-timestamp",
				}, nil
			},
			SendAlarmFunc: func(ctx context.Context, summary string, err error, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{
					ChannelID: "example-channel",
					Timestamp: "example-timestamp",
				}, nil
			},
			UpdatePublishLogAsAlarmFunc: func(ctx context.Context, ref *slack.MessageRef, summary string, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{
					ChannelID: "example-channel",
					Timestamp: "example-timestamp",
				}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:             store.Datastore{Backend: mockedDatastore},
			StateMachine:          application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
			DataBundleSlackClient: mockSlackClient,
			DatasetAPIClient:      mockDatasetAPIClient,
		}

		Convey("When UpdateBundleState is called to publish a bundle which has content items that will fail", func() {
			result, err := stateMachine.UpdateBundleState(ctx, bundleID, currentBundle.ETag, bundleUpdate.State, authEntityData)
			Convey("Then the bundle will continue to publish but slack alerts should be sent for the failing content items", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.State, ShouldEqual, models.BundleStatePublished)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 1)
				So(len(mockSlackClient.SendPublishLogCalls()), ShouldEqual, 1)
				So(len(mockSlackClient.UpdatePublishLogAsAlarmCalls()), ShouldEqual, 1)
				So(len(mockSlackClient.SendAlarmCalls()), ShouldEqual, 2)
			})
		})
	})
}

func TestPutBundlePolicy_Failures(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123
		userEmail := userEmail

		var previewTeams = make([]models.PreviewTeam, 0, 1)
		previewTeams = append(previewTeams, models.PreviewTeam{ID: "test-team"})

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			ETag:  "old-etag",
		}

		bundleUpdate := &models.Bundle{
			ID:           bundleID,
			State:        models.BundleStateDraft,
			ETag:         "new-etag",
			PreviewTeams: &previewTeams,
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				AccessToken: "test-token",
			},
		}

		var states = make([]application.State, 0, 1)
		states = append(states, application.Draft)

		var transitions = make([]application.Transition, 0, 1)
		transitions = append(transitions, application.Transition{
			Label:               "DRAFT",
			TargetState:         application.Draft,
			AllowedSourceStates: []string{"IN_REVIEW", "DRAFT"},
		})

		contentItems := createMockVersionsAndContentItems(bundleUpdate.State)

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return currentBundle, nil
			},
			UpdateBundleFunc: func(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
				return bundle, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return contentItems, nil
			},
		}

		Convey("When PutBundle is called with a preview team added but the creation of the policy fails", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("GetPolicyError")
				},
			}

			stateMachine := &application.StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockedDatastore},
				StateMachine:         application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, currentBundle.ETag)

			Convey("Then the bundle should not be updated", func() {
				So(err, ShouldEqual, errors.New("failed to create bundle policies"))
				So(result, ShouldBeNil)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 0)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 0)
			})
		})

		Convey("When PutBundle is called with a preview team added but the request fails", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{}, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					return errors.New("UpdatePolicyError")
				},
			}

			stateMachine := &application.StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockedDatastore},
				StateMachine:         application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, currentBundle.ETag)

			Convey("Then the bundle should not be updated", func() {
				So(err, ShouldEqual, errors.New("failed to add policy conditions for added preview teams"))
				So(result, ShouldBeNil)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 0)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestPutBundle_Fails(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123
		userEmail := userEmail

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			Title: "Old title",
			ETag:  "old-etag",
		}

		bundleUpdate := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateDraft,
			Title: "New title which already exists",
			ETag:  "new-etag",
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				AccessToken: "test-token",
			},
		}

		var states = make([]application.State, 0, 1)
		states = append(states, application.Draft)

		var transitions = make([]application.Transition, 0, 1)
		transitions = append(transitions, application.Transition{
			Label:               "DRAFT",
			TargetState:         application.Draft,
			AllowedSourceStates: []string{"IN_REVIEW", "DRAFT"},
		})

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return currentBundle, nil
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				return true, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:    store.Datastore{Backend: mockedDatastore},
			StateMachine: application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
		}

		Convey("When PutBundle is called but the title is duplicated", func() {
			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, currentBundle.ETag)

			Convey("Then the bundle should not be updated", func() {
				So(err, ShouldEqual, errors.New("bundle with the same title already exists"))
				So(result, ShouldBeNil)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 0)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestApproveBundle_Failure(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123
		userEmail := userEmail

		currentBundle := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateInReview,
			Title: "Bundle title",
			ETag:  "old-etag",
		}

		bundleUpdate := &models.Bundle{
			ID:    bundleID,
			State: models.BundleStateApproved,
			Title: "Bundle title",
			ETag:  "new-etag",
		}

		authEntityData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{
				UserID: userEmail,
			},
			Headers: datasetAPISDK.Headers{
				AccessToken: "test-token",
			},
		}

		var states = make([]application.State, 0, 1)
		states = append(states, application.Draft)

		var transitions = make([]application.Transition, 0, 1)
		transitions = append(transitions, application.Transition{
			Label:               "APPROVED",
			TargetState:         application.Approved,
			AllowedSourceStates: []string{"IN_REVIEW"},
		})

		version := datasetAPIModels.Version{

			ID:        "valid-version-1",
			Version:   1,
			DatasetID: "dataset-id-1",
			Edition:   "edition-id-1",
			State:     "associated",
		}

		contentItems := []models.ContentItem{
			{
				ID:       "valid-content-item",
				BundleID: bundle123,
				Metadata: models.Metadata{
					DatasetID: "dataset-id-1",
					EditionID: "edition-id-1",
					VersionID: 1,
				},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc: func(ctx context.Context, bundleID string) (*models.Bundle, error) {
				return currentBundle, nil
			},
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) {
				return false, nil
			},
			GetBundleContentsForBundleFunc: func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
				return &contentItems, nil
			},
		}

		mockDatasetAPI := &datasetAPIMocks.ClienterMock{
			GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
				return version, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
			StateMachine:     application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
		}

		Convey("When PutBundle is called but the content items are not approved", func() {
			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, currentBundle.ETag)

			Convey("Then the bundle should not be updated", func() {
				So(err, ShouldEqual, errors.New("version state expected to be APPROVED when transitioning bundle to APPROVED"))
				So(result, ShouldBeNil)
				So(len(mockedDatastore.UpdateBundleCalls()), ShouldEqual, 0)
				So(len(mockedDatastore.CreateEventCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestUpdateContentItemsWithDatasetInfo_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		bundleID := bundle123

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
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID string) (datasetAPIModels.Dataset, error) {
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
			authHeaders := datasetAPISDK.Headers{AccessToken: "test-token"}
			err := stateMachine.UpdateContentItemsWithDatasetInfo(ctx, bundleID, authHeaders)

			Convey("Then it should update content items successfully", func() {
				So(err, ShouldBeNil)
				So(len(mockedDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.UpdateContentItemDatasetInfoCalls()), ShouldEqual, 1)
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].Title, ShouldEqual, "New Dataset Title")
				So(mockedDatastore.UpdateContentItemDatasetInfoCalls()[0].State, ShouldEqual, "PUBLISHED")
			})
		})
	})
}

func TestUpdateContentItemsWithDatasetInfo_NoContentItems(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with no content items", t, func() {
		ctx := context.Background()
		bundleID := bundle123

		mockedDatastore := &storetest.StorerMock{
			GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
				return []*models.ContentItem{}, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When UpdateContentItemsWithDatasetInfo is called", func() {
			authHeaders := datasetAPISDK.Headers{AccessToken: "test-token"}
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
		bundleID := bundle123

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
			GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID string) (datasetAPIModels.Dataset, error) {
				return datasetAPIModels.Dataset{}, errors.New("dataset API failure")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
		}

		Convey("When UpdateContentItemsWithDatasetInfo is called", func() {
			authHeaders := datasetAPISDK.Headers{AccessToken: "test-token"}
			err := stateMachine.UpdateContentItemsWithDatasetInfo(ctx, bundleID, authHeaders)

			Convey("Then it should complete with error and log the failure", func() {
				So(err, ShouldNotBeNil)
				So(len(mockedDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
				So(len(mockedDatastore.UpdateContentItemDatasetInfoCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestUpdateDatasetVersionReleaseDate(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()

		mockedDatastore := &storetest.StorerMock{}
		mockDatasetAPI := &datasetAPIMocks.ClienterMock{
			PutVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string, version datasetAPIModels.Version) (datasetAPIModels.Version, error) {
				if datasetID == "0" {
					return datasetAPIModels.Version{}, errors.New("request failed")
				}
				return version, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
		}

		Convey("When UpdateDatasetVersionReleaseDate is called and PutVersion is successful", func() {
			authHeaders := datasetAPISDK.Headers{AccessToken: "test-token"}
			err := stateMachine.UpdateDatasetVersionReleaseDate(ctx, &today, "1", "1", 1, authHeaders)

			Convey("Then there should be no error", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When UpdateDatasetVersionReleaseDate is called and PutVersion fails", func() {
			authHeaders := datasetAPISDK.Headers{AccessToken: "test-token"}
			err := stateMachine.UpdateDatasetVersionReleaseDate(ctx, &today, "0", "1", 1, authHeaders)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "request failed")
			})
		})
	})
}

func TestApproveBundle_RefreshesContentItemMetadataAndLinks(t *testing.T) {
	Convey("Given a StateMachineBundleAPI approving a bundle whose edition was renamed in dataset-api", t, func() {
		ctx := context.Background()
		bundleID := bundle123

		currentBundle := &models.Bundle{
			ID: bundleID, State: models.BundleStateInReview, Title: "Bundle title", ETag: "old-etag",
		}
		bundleUpdate := &models.Bundle{
			ID: bundleID, State: models.BundleStateApproved, Title: "Bundle title", ETag: "new-etag",
		}

		authData := &models.AuthEntityData{
			EntityData: &permissionsAPISDK.EntityData{UserID: userEmail},
			Headers:    datasetAPISDK.Headers{AccessToken: "test-token"},
		}

		states := []application.State{application.Approved}
		transitions := []application.Transition{{
			Label: "APPROVED", TargetState: application.Approved, AllowedSourceStates: []string{"IN_REVIEW"},
		}}

		contentItems := []models.ContentItem{{
			ID: "content-1", BundleID: bundleID,
			Metadata: models.Metadata{DatasetID: "dataset-1", EditionID: "old-edition", VersionID: 1},
			Links:    models.Links{Edit: "/old/edit", Preview: "/old/preview"},
		}}

		version := datasetAPIModels.Version{
			ID: "v1", Version: 1, State: datasetAPIModels.ApprovedState,
			Links: &datasetAPIModels.VersionLinks{
				Dataset: &datasetAPIModels.LinkObject{ID: "dataset-1"},
				Edition: &datasetAPIModels.LinkObject{ID: "new-edition"},
				WebPage: &datasetAPIModels.LinkObject{HRef: "https://example.com/datasets/dataset-1/editions/new-edition/versions/1"},
			},
		}

		mockedDatastore := &storetest.StorerMock{
			GetBundleFunc:                      func(ctx context.Context, id string) (*models.Bundle, error) { return currentBundle, nil },
			CheckBundleExistsByTitleUpdateFunc: func(ctx context.Context, title, excludeID string) (bool, error) { return false, nil },
			GetBundleContentsForBundleFunc:     func(ctx context.Context, id string) (*[]models.ContentItem, error) { return &contentItems, nil },
			UpdateContentItemMetadataAndLinksFunc: func(ctx context.Context, contentItemID, datasetID, editionID, editLink, previewLink string) error {
				return nil
			},
			UpdateContentItemStateFunc: func(ctx context.Context, contentItemID, state string) error { return nil },
			UpdateBundleFunc:           func(ctx context.Context, id string, b *models.Bundle) (*models.Bundle, error) { return b, nil },
			CreateEventFunc:            func(ctx context.Context, e *models.Event) error { return nil },
		}

		mockDatasetAPI := &datasetAPIMocks.ClienterMock{
			GetVersionFunc: func(ctx context.Context, h datasetAPISDK.Headers, dID, eID, vID string) (datasetAPIModels.Version, error) {
				return version, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore:        store.Datastore{Backend: mockedDatastore},
			DatasetAPIClient: mockDatasetAPI,
			StateMachine:     application.NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, nil),
		}

		Convey("When the bundle is approved", func() {
			result, err := stateMachine.PutBundle(ctx, bundleID, bundleUpdate, authData, currentBundle.ETag)

			Convey("Then the content item's metadata and links are refreshed from dataset-api", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.State, ShouldEqual, models.BundleStateApproved)

				calls := mockedDatastore.UpdateContentItemMetadataAndLinksCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].EditionID, ShouldEqual, "new-edition")
				So(calls[0].DatasetID, ShouldEqual, "dataset-1")
				So(calls[0].EditLink, ShouldEqual, "/data-admin/series/dataset-1/editions/new-edition/versions/1")
				So(calls[0].PreviewLink, ShouldEqual, "/datasets/dataset-1/editions/new-edition/versions/1")
			})
		})
	})
}
