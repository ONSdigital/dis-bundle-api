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
	currentDraftBundleID        = "Bundle-Draft-ID"
	currentInReviewBundleID     = "Bundle-InReview-ID"
	currentApprovedBundleID     = "Bundle-Approved-ID"
	currentPublishedBundleID    = "Bundle-Published-ID"
	currentUnknownStateBundleID = "Bundle-Unknown-ID"
)

var (
	bundleStateDraft               = models.BundleStateDraft
	bundleStateInReview            = models.BundleStateInReview
	bundleStateApproved            = models.BundleStateApproved
	bundleStatePublished           = models.BundleStatePublished
	bundleStateUnknown             = models.BundleState("UNKNOWN")
	currentBundleWithStateDraft    = &models.Bundle{State: &bundleStateDraft, ID: currentDraftBundleID}
	currentBundleWithStateInReview = &models.Bundle{State: &bundleStateInReview, ID: currentInReviewBundleID}
	currentBundleWithStateApproved = &models.Bundle{State: &bundleStateApproved, ID: currentApprovedBundleID}
	currentBundleWithStateUnknown  = &models.Bundle{State: &bundleStateUnknown, ID: currentUnknownStateBundleID}
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

	mockHTTPRequest := http.Request{}
	mockDatasetsAPI := datasetsmocks.CreateDatasetsClientMock()
	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetsAPI, eventsmocks.CreateSuccessMockBundleEventsManager())

	Convey("When transitioning from 'DRAFT' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateDraft, bundleStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateInReview, bundleStateDraft)
		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateApproved, bundleStatePublished)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateApproved, bundleStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateApproved, bundleStateDraft)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestTransition_failure(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions(createMockFailureTransitionHandler())

	mockHTTPRequest := http.Request{}
	mockedDatastore := &storetest.StorerMock{}
	mockDatasetsAPI := datasetsmocks.CreateDatasetsClientMock()

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetsAPI, eventsmocks.CreateSuccessMockBundleEventsManager())

	Convey("When transitioning from a state that is not in the transition list", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateUnknown, bundleStateInReview)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Description, ShouldContainSubstring, "no valid transition")
		})
	})

	Convey("When transitioning to a state that is not in the transition list", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateDraft, bundleStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Description, ShouldContainSubstring, "no transitions found for state ")
		})
	})

	Convey("When the state machine has a transition that contains an invalid state", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, &mockHTTPRequest, currentBundleWithStateDraft, bundleStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Description, ShouldStartWith, "incorrect state value")
		})
	})
}
