package application

import (
	"context"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	currentBundleWithStateDraft                          = &models.Bundle{State: Draft.String()}
	currentBundleWithStateInReview                       = &models.Bundle{State: InReview.String()}
	currentBundleWithStateInReviewAndContentsApproved    = &models.Bundle{State: InReview.String(), Contents: []models.BundleContent{{State: Approved.String()}}}
	currentBundleWithStateInReviewAndContentsNotApproved = &models.Bundle{State: InReview.String(), Contents: []models.BundleContent{{State: Draft.String()}}}
	currentBundleWithStateApproved                       = &models.Bundle{State: Approved.String()}
	currentBundleWithStateUnknown                        = &models.Bundle{State: "UNKNOWN"}

	bundleUpdateWithStateDraft     = &models.Bundle{State: Draft.String()}
	bundleUpdateWithStateInReview  = &models.Bundle{State: InReview.String()}
	bundleUpdateWithStateApproved  = &models.Bundle{State: Approved.String()}
	bundleUpdateWithStatePublished = &models.Bundle{State: Published.String()}
	bundleUpdateWithStateUnknown   = &models.Bundle{State: "UNKNOWN"}
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

func TestGetStateByName(t *testing.T) {
	Convey("Given a state name", t, func() {
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

	mockedDatastore := &storetest.StorerMock{}

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine)

	Convey("When transitioning from 'DRAFT' to 'IN_REVIEW'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'IN_REVIEW' to 'APPROVED' with bundle contents APPROVED", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReviewAndContentsApproved, bundleUpdateWithStateApproved)

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
}

func TestTransition_failure(t *testing.T) {
	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	mockedDatastore := &storetest.StorerMock{}

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine)

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
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReviewAndContentsNotApproved, bundleUpdateWithStateApproved)
		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not all bundle contents are approved")
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
}
