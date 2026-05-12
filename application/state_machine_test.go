package application

import (
	"context"
	"strings"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"

	slackMock "github.com/ONSdigital/dis-bundle-api/slack/mocks"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	currentBundleWithStateDraft    = &models.Bundle{State: models.BundleStateDraft, LastUpdatedBy: &models.User{}}
	currentBundleWithStateInReview = &models.Bundle{State: models.BundleStateInReview, LastUpdatedBy: &models.User{}}
	currentBundleWithStateApproved = &models.Bundle{State: models.BundleStateApproved, LastUpdatedBy: &models.User{}}

	bundleUpdateWithStateDraft     = &models.Bundle{State: models.BundleStateDraft}
	bundleUpdateWithStateInReview  = &models.Bundle{State: models.BundleStateInReview}
	bundleUpdateWithStateApproved  = &models.Bundle{State: models.BundleStateApproved}
	bundleUpdateWithStatePublished = &models.Bundle{State: models.BundleStatePublished}
)

const (
	mockBundleID = "test-bundle-1234"
	userEmail    = "user@example.com"
)

func getMockStates() []State {
	return []State{
		Draft,
		InReview,
		Approved,
		Published,
	}
}

func getMockTransitions() []Transition {
	return []Transition{
		{
			Label:               "DRAFT",
			TargetState:         Draft,
			AllowedSourceStates: []string{"IN_REVIEW", "APPROVED"},
		},
		{
			Label:               "IN_REVIEW",
			TargetState:         InReview,
			AllowedSourceStates: []string{"DRAFT", "APPROVED"},
		},
		{
			Label:               "APPROVED",
			TargetState:         Approved,
			AllowedSourceStates: []string{"IN_REVIEW"},
		},
		{
			Label:               "PUBLISHED",
			TargetState:         Published,
			AllowedSourceStates: []string{"APPROVED"},
		},
	}
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
			BundleID: mockBundleID,
			State:    (*models.State)(&state),
			Metadata: models.Metadata{
				DatasetID: mockVersions[0].DatasetID,
				EditionID: mockVersions[0].Edition,
				VersionID: mockVersions[0].Version,
			},
		},
		{
			ID:       "another-valid-content-item",
			BundleID: mockBundleID,
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

func TestTransition_Success(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
	mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
	mockSlackClient := &slackMock.ClienterMock{}

	authEntityData := &models.AuthEntityData{
		EntityData: &permissionsAPISDK.EntityData{
			UserID: userEmail,
		},
		Headers: datasetAPISDK.Headers{
			AccessToken: "test-token",
		},
	}

	mockBundle := &models.Bundle{
		ID: mockBundleID,
		LastUpdatedBy: &models.User{
			Email: "email@ons.com",
		},
	}

	mockedDatastore := &storetest.StorerMock{
		CreateEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}

	Convey("When transitioning from 'DRAFT' to 'IN_REVIEW'", t, func() {
		mockedDatastore.UpdateBundleFunc = func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
			return bundleUpdateWithStateInReview, nil
		}

		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient, mockPermissionsAPIClient, mockSlackClient, "")

		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateInReview.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateInReview.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateInReview.State)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'APPROVED'", t, func() {
		fromState := models.BundleStateInReview
		mockBundle.State = fromState

		mockContentItems := createMockVersionsAndContentItems(fromState)

		mockedDatastore.UpdateBundleFunc = func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
			return bundleUpdateWithStateApproved, nil
		}
		mockedDatastore.GetBundleContentsForBundleFunc = func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
			contentItems := make([]models.ContentItem, len(mockContentItems))
			for index := range contentItems {
				contentItems[index] = *mockContentItems[index]
			}
			return &contentItems, nil
		}
		mockedDatastore.UpdateContentItemStateFunc = func(ctx context.Context, contentItemID, state string) error {
			return nil
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{
			PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
				return nil
			},
		}

		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient, mockPermissionsAPIClient, mockSlackClient, "")

		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateApproved.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateApproved.State)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED'", t, func() {
		fromState := models.BundleStateApproved
		mockBundle.State = fromState

		mockContentItems := createMockVersionsAndContentItems(fromState)

		mockedDatastore.UpdateBundleFunc = func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
			return bundleUpdateWithStatePublished, nil
		}
		mockedDatastore.GetBundleContentsForBundleFunc = func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
			contentItems := make([]models.ContentItem, len(mockContentItems))
			for index := range contentItems {
				contentItems[index] = *mockContentItems[index]
			}
			return &contentItems, nil
		}
		mockedDatastore.UpdateContentItemStateFunc = func(ctx context.Context, contentItemID, state string) error {
			return nil
		}

		mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{
			PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
				return nil
			},
		}

		mockSlackClient := &slackMock.ClienterMock{
			SendPublishLogFunc: func(ctx context.Context, summary string, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{}, nil
			},
			UpdatePublishLogFunc: func(ctx context.Context, ref *slack.MessageRef, summary string, fields []slack.Field) (*slack.MessageRef, error) {
				return &slack.MessageRef{}, nil
			},
		}

		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient, mockPermissionsAPIClient, mockSlackClient, "")
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStatePublished.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStatePublished.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStatePublished.State)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'DRAFT'", t, func() {
		mockedDatastore := &storetest.StorerMock{

			UpdateBundleFunc: func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
				return bundleUpdateWithStateDraft, nil
			},
			CreateEventFunc: func(ctx context.Context, event *models.Event) error {
				return nil
			},
		}
		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient, mockPermissionsAPIClient, mockSlackClient, "")
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateDraft.State, *authEntityData)
		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateDraft.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateDraft.State)
		})
	})
}

func TestTransition_failure(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	mockedDatastore := &storetest.StorerMock{}
	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
	mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
	mockSlackClient := &slackMock.ClienterMock{}

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient, mockPermissionsAPIClient, mockSlackClient, "")

	authEntityData := &models.AuthEntityData{
		EntityData: &permissionsAPISDK.EntityData{
			UserID: userEmail,
		},
		Headers: datasetAPISDK.Headers{
			AccessToken: "test-token",
		},
	}

	Convey("When transitioning to a state that is not in the transition list", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, bundleUpdateWithStateInReview, bundleUpdateWithStatePublished.State, *authEntityData)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "state not allowed to transition")
			So(bundle, ShouldBeNil)
		})
	})

}
