package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"

	slackMock "github.com/ONSdigital/dis-bundle-api/slack/mocks"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	currentBundleWithStateDraft     = &models.Bundle{State: models.BundleStateDraft}
	currentBundleWithStateInReview  = &models.Bundle{State: models.BundleStateInReview}
	currentBundleWithStateApproved  = &models.Bundle{State: models.BundleStateApproved}
	currentBundleWithStatePublished = &models.Bundle{State: models.BundleStatePublished}
	currentBundleWithStateUnknown   = &models.Bundle{State: models.BundleState("UNKNOWN")}

	bundleUpdateWithStateDraft     = &models.Bundle{State: models.BundleStateDraft}
	bundleUpdateWithStateInReview  = &models.Bundle{State: models.BundleStateInReview}
	bundleUpdateWithStateApproved  = &models.Bundle{State: models.BundleStateApproved}
	bundleUpdateWithStatePublished = &models.Bundle{State: models.BundleStatePublished}
	bundleUpdateWithStateUnknown   = &models.Bundle{State: models.BundleState("UNKNOWN")}
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

func TestTransition_success(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	mockedDatastore := &storetest.StorerMock{
		CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
			return true, nil
		},
	}

	userEmail := "user@example.com"

	authEntityData := &models.AuthEntityData{
		EntityData: &permissionsAPISDK.EntityData{
			UserID: userEmail,
		},
		Headers: datasetAPISDK.Headers{
			AccessToken: "test-token",
		},
	}

	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
	mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{}
	mockSlackClient := &slackMock.ClienterMock{}
	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient, mockPermissionsAPIClient, mockSlackClient, "")

	Convey("When transitioning from 'DRAFT' to 'IN_REVIEW'", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateInReview.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateInReview.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateInReview.State)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'APPROVED' with bundle contents APPROVED", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateApproved.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateApproved.State)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'DRAFT'", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateDraft.State, *authEntityData)
		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateDraft.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateDraft.State)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED'", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStatePublished.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStatePublished.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStatePublished.State)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'IN_REVIEW'", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStateInReview.State, *authEntityData)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
			So(bundle, ShouldNotBeNil)
			So(bundle.ID, ShouldEqual, bundleUpdateWithStateInReview.ID)
			So(bundle.State, ShouldEqual, bundleUpdateWithStateInReview.State)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'DRAFT'", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStateDraft.State, *authEntityData)

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

	userEmail := "user@example.com"

	authEntityData := &models.AuthEntityData{
		EntityData: &permissionsAPISDK.EntityData{
			UserID: userEmail,
		},
		Headers: datasetAPISDK.Headers{
			AccessToken: "test-token",
		},
	}

	Convey("When transitioning from a state that is not in the transition list", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateUnknown, bundleUpdateWithStateInReview.State, *authEntityData)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "state not allowed to transition")
			So(bundle, ShouldBeNil)
		})
	})

	Convey("When transitioning to a state that is not in the transition list", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateUnknown.State, *authEntityData)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "state not allowed to transition")
			So(bundle, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'APPROVED' with bundle contents not APPROVED", t, func() {
		Convey("And CheckAllBundleContentsAreApproved returns false", func() {
			stateMachineBundleAPI.Datastore.Backend = &storetest.StorerMock{
				CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
					return false, nil
				},
			}

			Convey("Then the transition should fail", func() {
				bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved.State, *authEntityData)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "not all bundle contents are approved")
				So(bundle, ShouldBeNil)
			})
		})

		Convey("And CheckAllBundleContentsAreApproved returns an error", func() {
			stateMachineBundleAPI.Datastore.Backend = &storetest.StorerMock{
				CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
					return false, errors.New("datastore error")
				},
			}

			Convey("Then the transition should fail with an error", func() {
				bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved.State, *authEntityData)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "datastore error")
				So(bundle, ShouldBeNil)
			})
		})
	})

	Convey("When the state machine has a transition that contains an invalid state", t, func() {
		stateMachineBundleAPI.StateMachine.transitions["UNKNOWN"] = []string{"DRAFT"}
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateUnknown.State, *authEntityData)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "incorrect state value")
			So(bundle, ShouldBeNil)
		})
	})

	Convey("When transitioning from nil current bundle to 'APPROVED'", t, func() {
		bundle, err := stateMachine.Transition(ctx, stateMachineBundleAPI, nil, bundleUpdateWithStateApproved.State, *authEntityData)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "bundle state must be DRAFT when creating a new bundle")
			So(bundle, ShouldBeNil)
		})
	})

	// Convey("When transitioning from 'PUBLISHED' to nil", t, func() {
	// 	err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStatePublished, nil)

	// 	Convey("Then the transition should fail", func() {
	// 		So(err, ShouldNotBeNil)
	// 		So(err.Error(), ShouldEqual, "cannot update a published bundle")
	// 	})
	// })
}

func TestIsValidTransition(t *testing.T) {
	validTransitions := []struct {
		fromState      models.BundleState
		toState        models.BundleState
		expectedResult *error
	}{
		{models.BundleStateDraft, models.BundleStateInReview, nil},
		{models.BundleStateInReview, models.BundleStateDraft, nil},
		{models.BundleStateInReview, models.BundleStateApproved, nil},
		{models.BundleStateApproved, models.BundleStatePublished, nil},
		{models.BundleStateApproved, models.BundleStateInReview, nil},
	}

	t.Parallel()

	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	mockedDatastore := &storetest.StorerMock{
		CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
			return true, nil
		},
	}

	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore}, mockDatasetAPIClient)

	for index := range validTransitions {
		tc := validTransitions[index]
		t.Run(fmt.Sprintf("When validating a valid transition from %s to %s", tc.fromState, tc.toState), func(t *testing.T) {
			t.Parallel()

			Convey("Then no error should be returned", t, func() {
				err := stateMachine.IsValidTransition(ctx, &tc.fromState, &tc.toState)

				So(err, ShouldBeNil)
			})
		})
	}

	invalidTransitions := []struct {
		fromState      models.BundleState
		toState        models.BundleState
		expectedResult *error
	}{
		{models.BundleStateDraft, models.BundleStateApproved, nil},
		{models.BundleStateDraft, models.BundleStatePublished, nil},
		{models.BundleStateInReview, models.BundleStatePublished, nil},

		// Published bundle cannot transition
		{models.BundleStatePublished, models.BundleStateInReview, nil},
		{models.BundleStatePublished, models.BundleStateApproved, nil},
		{models.BundleStatePublished, models.BundleStateDraft, nil},
	}

	for index := range invalidTransitions {
		tc := invalidTransitions[index]
		t.Run(fmt.Sprintf("When validating an invalid transition from %s to %s", tc.fromState, tc.toState), func(t *testing.T) {
			t.Parallel()

			Convey("Then an error should be returned", t, func() {
				err := stateMachine.IsValidTransition(ctx, &tc.fromState, &tc.toState)

				So(err, ShouldNotBeNil)

				Convey("And the error should be an invalid transition error", func() {
					So(err, ShouldEqual, apierrors.ErrInvalidTransition)
				})
			})
		})
	}
}

const (
	mockUserID       = "mock-user-id"
	mockServiceToken = "mock-service-token"
	mockBundleID     = "test-bundle-1234"
)
