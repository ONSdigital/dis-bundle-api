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
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPIMocks "github.com/ONSdigital/dp-dataset-api/sdk/mocks"

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

func ptrBundleState(s models.BundleState) *models.BundleState {
	return &s
}

func ptrContentItemState(s models.State) *models.State {
	return &s
}

func TestGetBundleContents(t *testing.T) {
	ctx := context.Background()
	authHeaders := sdk.Headers{}

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
					State:         &state,
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
					State: &state,
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

			mockDatasetAPI.GetDatasetFunc = func(ctx context.Context, headers sdk.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
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
					State: &state,
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

			mockDatasetAPI.GetDatasetFunc = func(ctx context.Context, headers sdk.Headers, collectionID, datasetID string) (datasetAPIModels.Dataset, error) {
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
					State: &state,
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
