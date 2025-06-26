package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
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

func TestGetStateByName_Success(t *testing.T) {
	Convey("Given a valid state name", t, func() {
		Convey("When the state name is 'DRAFT'", func() {
			state, found := getStateByName("DRAFT")

			Convey("Then it should return the DRAFT state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "DRAFT")
			})
		})

		Convey("When the state name is 'IN_REVIEW'", func() {
			state, found := getStateByName("IN_REVIEW")

			Convey("Then it should return the IN_REVIEW state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "IN_REVIEW")
			})
		})

		Convey("When the state name is 'APPROVED'", func() {
			state, found := getStateByName("APPROVED")

			Convey("Then it should return the APPROVED state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "APPROVED")
			})
		})

		Convey("When the state name is 'PUBLISHED'", func() {
			state, found := getStateByName("PUBLISHED")

			Convey("Then it should return the Published state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "PUBLISHED")
			})
		})
	})
}

func TestGetStateByName_Failure(t *testing.T) {
	Convey("Given an invalid state name", t, func() {
		Convey("When the state name is 'UNKNOWN'", func() {
			state, found := getStateByName("UNKNOWN")

			Convey("Then it should return nil and found should be false", func() {
				So(found, ShouldBeFalse)
				So(state, ShouldBeNil)
			})
		})
	})
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

	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}
	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient)

	Convey("When transitioning from 'DRAFT' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'APPROVED' with bundle contents APPROVED", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateDraft)
		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'PUBLISHED'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStatePublished)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'APPROVED' to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStateDraft)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from nil current bundle to 'DRAFT'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, nil, bundleUpdateWithStateDraft)

		Convey("Then the transition should not fail", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from any state that is not 'PUBLISHED' to nil", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, nil)

		Convey("Then the transition should not fail", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestTransition_failure(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	mockedDatastore := &storetest.StorerMock{}
	mockDatasetAPIClient := &datasetAPISDKMock.ClienterMock{}

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, mockDatasetAPIClient)

	Convey("When transitioning from a state that is not in the transition list", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateUnknown, bundleUpdateWithStateInReview)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "state not allowed to transition")
		})
	})

	Convey("When transitioning to a state that is not in the transition list", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "state not allowed to transition")
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
				err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "not all bundle contents are approved")
			})
		})

		Convey("And CheckAllBundleContentsAreApproved returns an error", func() {
			stateMachineBundleAPI.Datastore.Backend = &storetest.StorerMock{
				CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
					return false, errors.New("datastore error")
				},
			}

			Convey("Then the transition should fail with an error", func() {
				err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateApproved)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "datastore error")
			})
		})
	})

	Convey("When the state machine has a transition that contains an invalid state", t, func() {
		stateMachineBundleAPI.StateMachine.transitions["UNKNOWN"] = []string{"DRAFT"}
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "incorrect state value")
		})
	})

	Convey("When transitioning from nil current bundle to 'APPROVED'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, nil, bundleUpdateWithStateApproved)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "bundle state must be DRAFT when creating a new bundle")
		})
	})

	Convey("When transitioning from 'PUBLISHED' to nil", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStatePublished, nil)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "cannot update a published bundle")
		})
	})
}
