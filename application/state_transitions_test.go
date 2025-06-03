package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/datasets"
	datasetsmock "github.com/ONSdigital/dis-bundle-api/datasets/mocks"
	eventsmocks "github.com/ONSdigital/dis-bundle-api/events/mocks"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetsmodels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

type ContentItemVersion struct {
	ContentItem *models.ContentItem
	Version     *datasetsmodels.Version
}

type MockDataBuilder struct {
	bundles      map[string]*models.Bundle
	contentItems map[string][]*ContentItemVersion
}

func NewMockDataBuilder() *MockDataBuilder {
	return &MockDataBuilder{
		bundles:      make(map[string]*models.Bundle),
		contentItems: make(map[string][]*ContentItemVersion),
	}
}

func (b *MockDataBuilder) WithBundle(id string, state *models.BundleState) *MockDataBuilder {
	b.bundles[id] = &models.Bundle{ID: id, State: state}
	return b
}

func (b *MockDataBuilder) WithContentItem(bundleID, itemID string, state *models.State, version int, versionState *string) *MockDataBuilder {
	item := ContentItemVersion{
		ContentItem: &models.ContentItem{
			ID:       itemID,
			BundleID: bundleID,
			State:    state,
		},
		Version: &datasetsmodels.Version{
			Version: version,
			ID:      fmt.Sprintf("bundle-%s-item-%s-version-%d", bundleID, itemID, version),
			State:   state.String(),
		},
	}

	if versionState != nil {
		item.Version.State = *versionState
	}

	b.contentItems[bundleID] = append(b.contentItems[bundleID], &item)
	return b
}

func (b *MockDataBuilder) Build() (map[string]*models.Bundle, map[string][]*ContentItemVersion) {
	return b.bundles, b.contentItems
}

const (
	bundleIdSuccess         = "success-bundle"
	bundleIdMismatchedState = "mismatched-state-bundle"
	bundleIdNoContentItems  = "no-content-items-bundle"

	contentItemIdMatchingState                       = "matching-state-content-id"
	contentItemIdMatchingStateMismatchedVersionState = "matching-content-state-mismatched-version-state-id"
	contentItemIdMismatchedState                     = "mismatched-state-content-id"
)

func createSuccessScenario() (map[string]*models.Bundle, map[string][]*ContentItemVersion) {
	approvedState := models.StateApproved
	bundleStateApproved := models.BundleStateApproved
	draftState := models.StateDraft
	draftStateString := draftState.String()

	return NewMockDataBuilder().
		WithBundle(bundleIdSuccess, &bundleStateApproved).
		WithContentItem(bundleIdSuccess, "ContentItemA", &approvedState, 1, nil).
		WithContentItem(bundleIdSuccess, "ContentItemB", &approvedState, 1, nil).
		WithBundle(bundleIdMismatchedState, &bundleStateApproved).
		WithContentItem(bundleIdMismatchedState, contentItemIdMatchingState, &approvedState, 1, nil).
		WithContentItem(bundleIdMismatchedState, contentItemIdMismatchedState, &draftState, 1, nil).
		WithContentItem(bundleIdMismatchedState, contentItemIdMatchingStateMismatchedVersionState, &approvedState, 1, &draftStateString).
		WithBundle(bundleIdNoContentItems, &bundleStateApproved).
		Build()
}

func createSuccessMockStorer(contentItems map[string][]*ContentItemVersion, bundles map[string]*models.Bundle) *storetest.StorerMock {
	return &storetest.StorerMock{
		GetContentsForBundleFunc: func(ctx context.Context, bundleID string) ([]models.ContentItem, error) {
			if versions, exists := contentItems[bundleID]; exists {
				items := make([]models.ContentItem, len(versions))
				for i, version := range versions {
					items[i] = *version.ContentItem
				}
				return items, nil
			}
			return nil, nil
		},
		UpdateBundleContentItemStateFunc: func(ctx context.Context, contentItemID string, state models.BundleState) error {
			for _, bundleItems := range contentItems {
				for _, version := range bundleItems {
					if version.ContentItem.ID == contentItemID {
						contentItemState, err := models.GetMatchingStateForBundleState(state)
						if err != nil {
							return err
						}
						version.ContentItem.State = contentItemState
						return nil
					}
				}
			}

			return errors.New("not found content item")
		},
		UpdateBundleStateFunc: func(ctx context.Context, bundleID string, state models.BundleState) error {
			if bundle, exists := bundles[bundleID]; exists {
				bundle.State = &state
				return nil
			}

			return errors.New("not found bundle")
		},
		CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}
}

const (
	errorDescriptionGetContentsForBundle = "error getting content items for bundle"
)

func createErrorMockStorer() *storetest.StorerMock {
	return &storetest.StorerMock{
		GetContentsForBundleFunc: func(ctx context.Context, bundleID string) ([]models.ContentItem, error) {
			return nil, errors.New(errorDescriptionGetContentsForBundle)
		},
	}
}

func createMockDatasetsApi(contentItems map[string][]*ContentItemVersion, getForContentItemsError *string, updateStateForContentItemError *string) datasetsmock.DatasetsVersionsClientMock {
	itemVersionMap := make(map[string]*datasetsmodels.Version)
	for _, versions := range contentItems {
		for _, cv := range versions {
			itemVersionMap[cv.ContentItem.ID] = cv.Version
		}
	}

	return datasetsmock.DatasetsVersionsClientMock{
		GetForContentItemFunc: func(ctx context.Context, r *http.Request, contentItem models.ContentItem) (*datasetsmodels.Version, error) {
			if getForContentItemsError != nil {
				return nil, errors.New(*getForContentItemsError)
			}

			if version, exists := itemVersionMap[contentItem.ID]; exists {
				return version, nil
			}
			return nil, fmt.Errorf("failed to find version for content item %s", contentItem.ID)
		},
		UpdateStateForContentItemFunc: func(ctx context.Context, r *http.Request, contentItem models.ContentItem, targetState models.BundleState) error {
			if updateStateForContentItemError != nil {
				return errors.New(*updateStateForContentItemError)
			}

			if version, exists := itemVersionMap[contentItem.ID]; exists {
				version.State = targetState.String()
				itemVersionMap[contentItem.ID] = version
				return nil
			}
			return fmt.Errorf("content item %s not found", contentItem.ID)
		},
	}
}

func createSuccessMockDatasetApi(contentItems map[string][]*ContentItemVersion) datasetsmock.DatasetsVersionsClientMock {
	return createMockDatasetsApi(contentItems, nil, nil)
}

func TestHandleApprovedToPublished_success(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions(UpdateBundleState)

	bundles, contentItems := createSuccessScenario()

	mockedDatastore := createSuccessMockStorer(contentItems, bundles)
	mockDatasetsVersionsApi := createSuccessMockDatasetApi(contentItems)

	mockDatasetsApi := datasetsmock.DatasetsClientMock{
		VersionsFunc: func() datasets.DatasetsVersionsClient {
			return &mockDatasetsVersionsApi
		},
	}

	mockHttpRequest := http.Request{}
	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED' with all valid matching states", t, func() {
		testBundle := bundles[bundleIdSuccess]
		testContentItemVersion := contentItems[bundleIdSuccess]

		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, testBundle, bundleStatePublished)
		Convey("No errors should occurr", func() {
			So(err, ShouldBeNil)
		})

		Convey("The bundle content items + versions state should be updated", func() {
			for _, contentItemVersion := range testContentItemVersion {
				So(contentItemVersion.Version.State, ShouldEqual, models.BundleStatePublished.String())
				So(contentItemVersion.ContentItem.State.String(), ShouldEqual, models.BundleStatePublished.String())
			}
		})

		Convey("The bundle state should be updated", func() {
			So(*testBundle.State, ShouldEqual, bundleStatePublished)
		})

	})

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED' with state mismatches", t, func() {
		testBundle := bundles[bundleIdMismatchedState]
		testContentItemVersion := contentItems[bundleIdMismatchedState]

		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, testBundle, bundleStatePublished)
		Convey("No errors should occurr", func() {
			So(err, ShouldBeNil)
		})

		Convey("The bundle content item + version states should be updated for matching states", func() {
			for _, contentItemVersion := range testContentItemVersion {
				if contentItemVersion.ContentItem.ID == contentItemIdMatchingState {
					So(contentItemVersion.Version.State, ShouldEqual, models.BundleStatePublished.String())
					So(contentItemVersion.ContentItem.State.String(), ShouldEqual, models.BundleStatePublished.String())
				}
			}
		})

		Convey("The bundle version state should not be updated for content items with mismatched states", func() {
			for _, contentItemVersion := range testContentItemVersion {
				if contentItemVersion.ContentItem.ID == contentItemIdMismatchedState {
					So(contentItemVersion.Version.State, ShouldEqual, models.BundleStateDraft.String())
				}
			}
		})

		Convey("The bundle version state should not be updated for mismatched version states", func() {
			for _, contentItemVersion := range testContentItemVersion {
				if contentItemVersion.ContentItem.ID == contentItemIdMatchingStateMismatchedVersionState {
					So(contentItemVersion.Version.State, ShouldEqual, models.BundleStateDraft.String())
				}
			}
		})
	})
}

func TestHandleApprovedToPublished_failure(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions(UpdateBundleState)

	bundles, contentItems := createSuccessScenario()

	mockHttpRequest := http.Request{}
	Convey("When transitioning from 'APPROVED' to 'PUBLISHED'", t, func() {

		Convey("When no content items are found for the bundle", func() {
			mockedDatastore := createSuccessMockStorer(contentItems, bundles)
			stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})

			mockDatasetsVersionsApi := createSuccessMockDatasetApi(contentItems)

			mockDatasetsApi := datasetsmock.DatasetsClientMock{
				VersionsFunc: func() datasets.DatasetsVersionsClient {
					return &mockDatasetsVersionsApi
				},
			}

			stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

			testBundle := bundles[bundleIdNoContentItems]
			err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, testBundle, bundleStatePublished)
			Convey("An error should be returned", func() {
				expectedErrorCode := models.CodeNotFound
				So(err, ShouldNotBeNil)
				So(err.Description, ShouldEqual, apierrors.ErrorDescriptionNoContentItemsFound)
				So(err.Code, ShouldEqual, &expectedErrorCode)
			})
		})

		Convey("When an error occurs getting bundle content items", func() {
			mockedDatastore := createErrorMockStorer()
			stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})

			mockDatasetsVersionsApi := createSuccessMockDatasetApi(contentItems)

			mockDatasetsApi := datasetsmock.DatasetsClientMock{
				VersionsFunc: func() datasets.DatasetsVersionsClient {
					return &mockDatasetsVersionsApi
				},
			}

			stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

			testBundle := bundles[bundleIdSuccess]
			err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, testBundle, bundleStatePublished)
			Convey("An error should be returned", func() {
				expectedErrorCode := models.CodeInternalServerError
				So(err, ShouldNotBeNil)
				So(err.Description, ShouldEqual, errorDescriptionGetContentsForBundle)
				So(err.Code, ShouldEqual, &expectedErrorCode)
			})
		})

		Convey("When an error occurs getting version from dataset API", func() {
			mockedDatastore := createSuccessMockStorer(contentItems, bundles)
			stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})

			getVersionErrorMessage := "error occurred getting version"
			mockDatasetsVersionsApi := createMockDatasetsApi(contentItems, &getVersionErrorMessage, nil)

			mockDatasetsApi := datasetsmock.DatasetsClientMock{
				VersionsFunc: func() datasets.DatasetsVersionsClient {
					return &mockDatasetsVersionsApi
				},
			}

			stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

			testBundle := bundles[bundleIdSuccess]
			err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, testBundle, bundleStatePublished)
			Convey("An error should be returned", func() {
				expectedErrorCode := models.CodeInternalServerError
				So(err, ShouldNotBeNil)
				So(err.Description, ShouldEqual, getVersionErrorMessage)
				So(err.Code, ShouldEqual, &expectedErrorCode)
			})
		})

		Convey("When an error occurs getting updating the version state in dataset API", func() {
			mockedDatastore := createSuccessMockStorer(contentItems, bundles)
			stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})

			updateVersionStateError := "error occurred updating version"
			mockDatasetsVersionsApi := createMockDatasetsApi(contentItems, nil, &updateVersionStateError)

			mockDatasetsApi := datasetsmock.DatasetsClientMock{
				VersionsFunc: func() datasets.DatasetsVersionsClient {
					return &mockDatasetsVersionsApi
				},
			}

			stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

			testBundle := bundles[bundleIdSuccess]
			err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, testBundle, bundleStatePublished)
			Convey("An error should be returned", func() {
				expectedErrorCode := models.CodeInternalServerError
				So(err, ShouldNotBeNil)
				So(err.Description, ShouldEqual, updateVersionStateError)
				So(err.Code, ShouldEqual, &expectedErrorCode)
			})
		})

	})

}
