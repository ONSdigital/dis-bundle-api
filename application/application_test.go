package application_test

import (
	"context"
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
