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
	. "github.com/smartystreets/goconvey/convey"
)

var (
	today     = time.Now()
	yesterday = today.Add(-24 * time.Hour)
	tomorrow  = today.Add(24 * time.Hour)
)

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
				State:      ptrBundleState(models.BundleStatePublished),
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
			State:      ptrBundleState(models.BundleStatePublished),
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
			State:      ptrBundleState(models.BundleStatePublished),
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
			State: ptrContentItemState(models.StateApproved),
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
			BundleType:  models.BundleTypeScheduled,
			State:       ptrBundleState(models.BundleStateDraft),
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return bundleToCreate, nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called", func() {
			createdBundle, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return the created bundle without error", func() {
				So(err, ShouldBeNil)
				So(createdBundle, ShouldResemble, bundleToCreate)
			})
		})
	})
}

func TestCreateBundle_ValidationFailure(t *testing.T) {
	Convey("Given a bundle with an invalid ScheduledAt date", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &yesterday,
			BundleType:  models.BundleTypeScheduled,
			State:       ptrBundleState(models.BundleStateDraft),
		}

		mockedDatastore := &storetest.StorerMock{}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called", func() {
			createdBundle, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error indicating ScheduledAt cannot be in the past", func() {
				So(err, ShouldNotBeNil)
				So(createdBundle, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "scheduled_at cannot be in the past")
			})
		})
	})
}

func TestCreateBundle_FailureWhenSavingBundle(t *testing.T) {
	Convey("Given a valid bundle", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       ptrBundleState(models.BundleStateDraft),
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return errors.New("failed to create bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the bundle could not be created", func() {
			createdBundle, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(createdBundle, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to create bundle")
			})
		})
	})
}

func TestCreateBundle_FailureToGetBundle(t *testing.T) {
	Convey("Given a valid bundle", t, func() {
		ctx := context.Background()
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &tomorrow,
			BundleType:  models.BundleTypeScheduled,
			State:       ptrBundleState(models.BundleStateDraft),
		}

		mockedDatastore := &storetest.StorerMock{
			CreateBundleFunc: func(ctx context.Context, bundle *models.Bundle) error {
				return nil
			},
			GetBundleFunc: func(ctx context.Context, id string) (*models.Bundle, error) {
				return nil, errors.New("failed to get bundle")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When CreateBundle is called and the bundle could not be retrieved", func() {
			createdBundle, err := stateMachine.CreateBundle(ctx, bundleToCreate)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(createdBundle, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to get bundle")
			})
		})
	})
}

func TestCheckBundleExistsByTitle(t *testing.T) {
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

func TestValidateScheduledAt_Success(t *testing.T) {
	Convey("Given a bundle with a valid ScheduledAt date", t, func() {
		bundle := &models.Bundle{
			ScheduledAt: &tomorrow,
		}

		Convey("When validateScheduledAt is called", func() {
			err := application.ValidateScheduledAt(bundle)

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
			err := application.ValidateScheduledAt(bundle)

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
			err := application.ValidateScheduledAt(bundle)

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
			err := application.ValidateScheduledAt(bundle)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "scheduled_at cannot be in the past")
			})
		})
	})
}

func ptrBundleState(s models.BundleState) *models.BundleState {
	return &s
}

func ptrContentItemState(s models.State) *models.State {
	return &s
}
