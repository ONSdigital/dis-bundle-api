package application

import (
	"context"
	"net/http"
	"testing"

	datasetsmocks "github.com/ONSdigital/dis-bundle-api/datasets/mocks"
	eventsmocks "github.com/ONSdigital/dis-bundle-api/events/mocks"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	currentDraftBundleId        = "Bundle-Draft-Id"
	currentInReviewBundleId     = "Bundle-InReview-Id"
	currentApprovedBundleId     = "Bundle-Approved-Id"
	currentPublishedBundleId    = "Bundle-Published-Id"
	currentUnknownStateBundleId = "Bundle-Unknown-Id"
)

var (
	bundleStateDraft                = models.BundleStateDraft
	bundleStateInReview             = models.BundleStateInReview
	bundleStateApproved             = models.BundleStateApproved
	bundleStatePublished            = models.BundleStatePublished
	bundleStateUnknown              = models.BundleState("UNKNOWN")
	currentBundleWithStateDraft     = &models.Bundle{State: &bundleStateDraft, ID: currentDraftBundleId}
	currentBundleWithStateInReview  = &models.Bundle{State: &bundleStateInReview, ID: currentInReviewBundleId}
	currentBundleWithStateApproved  = &models.Bundle{State: &bundleStateApproved, ID: currentApprovedBundleId}
	currentBundleWithStatePublished = &models.Bundle{State: &bundleStateApproved, ID: currentPublishedBundleId}
	currentBundleWithStateUnknown   = &models.Bundle{State: &bundleStateUnknown, ID: currentUnknownStateBundleId}
)

func getMockStates() []models.BundleState {
	return []models.BundleState{
		bundleStateDraft,
		bundleStateInReview,
		bundleStateApproved,
		bundleStatePublished,
	}
}

func getMockTransitions(handler TransitionHandler) []Transition {
	return []Transition{
		{
			Label:               "DRAFT",
			TargetState:         bundleStateDraft,
			AllowedSourceStates: []models.BundleState{"IN_REVIEW", "APPROVED"},
			Handler:             handler,
		},
		{
			Label:               "IN_REVIEW",
			TargetState:         bundleStateInReview,
			AllowedSourceStates: []models.BundleState{"DRAFT", "APPROVED"},
			Handler:             handler,
		},
		{
			Label:               "APPROVED",
			TargetState:         bundleStateApproved,
			AllowedSourceStates: []models.BundleState{"IN_REVIEW"},
			Handler:             handler,
		},
		{
			Label:               "PUBLISHED",
			TargetState:         bundleStatePublished,
			AllowedSourceStates: []models.BundleState{"APPROVED"},
			Handler:             handler,
		},
	}
}

func createMockSuccessfulTransitionHandler() TransitionHandler {
	return func(ctx context.Context, api *StateMachineBundleAPI, r *http.Request, bundle *models.Bundle, targetState models.BundleState) *models.Error {
		return nil
	}
}

const (
	DefaultErrorCode        = models.CodeInternalServerError
	DefaultErrorDescription = "Test error description"
)

func createMockFailureTransitionHandler() TransitionHandler {
	return func(ctx context.Context, api *StateMachineBundleAPI, r *http.Request, bundle *models.Bundle, targetState models.BundleState) *models.Error {
		code := DefaultErrorCode

		description := DefaultErrorDescription

		return models.CreateModelError(code, description)
	}
}

func TestTransition_success(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions(createMockSuccessfulTransitionHandler())
	mockedDatastore := &storetest.StorerMock{
		CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
			return true, nil
		},
		GetContentsForBundleFunc: func(ctx context.Context, bundleID string) ([]models.ContentItem, error) {
			var items []models.ContentItem
			return items, nil
		},
	}

	mockHttpRequest := http.Request{}
	mockDatasetsApi := datasetsmocks.CreateDatasetsClientMock()
	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

	Convey("When transitioning from 'DRAFT' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateDraft, bundleStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateInReview, bundleStateDraft)
		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateApproved, bundleStatePublished)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateApproved, bundleStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateApproved, bundleStateDraft)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestTransition_failure(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions(createMockFailureTransitionHandler())

	mockHttpRequest := http.Request{}
	mockedDatastore := &storetest.StorerMock{}
	mockDatasetsApi := datasetsmocks.CreateDatasetsClientMock()

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetsApi, eventsmocks.CreateSuccessMockBundleEventsManager())

	Convey("When transitioning from a state that is not in the transition list", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateUnknown, bundleStateInReview)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Description, ShouldContainSubstring, "no valid transition")
		})
	})

	Convey("When transitioning to a state that is not in the transition list", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateDraft, bundleStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Description, ShouldContainSubstring, "no transitions found for state ")
		})
	})

	Convey("When the state machine has a transition that contains an invalid state", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHttpRequest, currentBundleWithStateDraft, bundleStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Description, ShouldStartWith, "incorrect state value")
		})
	})
}
