package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListBundles(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		now := time.Now().UTC()

		expectedBundles := []*models.Bundle{
			{
				ID:         "bundle-123",
				Title:      "Example Bundle",
				CreatedAt:  &now,
				UpdatedAt:  &now,
				BundleType: models.BundleTypeScheduled,
				State:      ptrBundleState(models.BundleStatePublished),
			},
		}

		mockedDatastore := &storetest.StorerMock{
			ListBundlesFunc: func(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
				return expectedBundles, len(expectedBundles), nil
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When ListBundles is called", func() {
			results, totalCount, err := stateMachine.ListBundles(ctx, 0, 10)

			Convey("Then it should return the expected bundles without error", func() {
				So(err, ShouldBeNil)
				So(results, ShouldResemble, expectedBundles)
				So(totalCount, ShouldEqual, len(expectedBundles))
			})
		})
	})
}

func TestCreateBundle_Success(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()
		now := time.Now().UTC()

		tomorrow := now.Add(24 * time.Hour)
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
		pastTime := time.Now().Add(-24 * time.Hour)
		bundleToCreate := &models.Bundle{
			ID:          "bundle-123",
			Title:       "Example Bundle",
			ScheduledAt: &pastTime,
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
		tomorrow := time.Now().Add(24 * time.Hour)
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
		tomorrow := time.Now().Add(24 * time.Hour)
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

func TestGetBundleByTitle(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with a mocked datastore", t, func() {
		ctx := context.Background()

		expectedBundle := &models.Bundle{
			ID:    "bundle-123",
			Title: "Example Bundle",
		}

		mockedDatastore := &storetest.StorerMock{
			GetBundleByTitleFunc: func(ctx context.Context, title string) (*models.Bundle, error) {
				if title == expectedBundle.Title {
					return expectedBundle, nil
				}
				return nil, errors.New("bundle not found")
			},
		}

		stateMachine := &application.StateMachineBundleAPI{
			Datastore: store.Datastore{Backend: mockedDatastore},
		}

		Convey("When GetBundleByTitle is called and a bundle with the same title already exist", func() {
			bundle, err := stateMachine.GetBundleByTitle(ctx, expectedBundle.Title)

			Convey("Then it should return the expected bundle without error", func() {
				So(err, ShouldBeNil)
				So(bundle, ShouldResemble, expectedBundle)
			})
		})

		Convey("When GetBundleByTitle is called and a bundle with the same title does not exist", func() {
			bundle, err := stateMachine.GetBundleByTitle(ctx, "Nonexistent Bundle")

			Convey("Then it should return an error and nil bundle", func() {
				So(err, ShouldNotBeNil)
				So(bundle, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "bundle not found")
			})
		})
	})
}

func TestValidateScheduledAt_Success(t *testing.T) {
	Convey("Given a bundle with a valid ScheduledAt date", t, func() {
		futureTime := time.Now().Add(24 * time.Hour)
		bundle := &models.Bundle{
			ScheduledAt: &futureTime,
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
		futureTime := time.Now().Add(24 * time.Hour)
		bundle := &models.Bundle{
			BundleType:  models.BundleTypeManual,
			ScheduledAt: &futureTime,
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
		pastTime := time.Now().Add(-24 * time.Hour)
		bundle := &models.Bundle{
			ScheduledAt: &pastTime,
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
