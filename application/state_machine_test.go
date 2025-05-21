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
	currentBundleWithStateUnknown                        = &models.Bundle{State: "unknown"}

	bundleUpdateWithStateDraft     = &models.Bundle{State: Draft.String()}
	bundleUpdateWithStateInReview  = &models.Bundle{State: InReview.String()}
	bundleUpdateWithStateApproved  = &models.Bundle{State: Approved.String()}
	bundleUpdateWithStatePublished = &models.Bundle{State: Published.String()}
	bundleUpdateWithStateUnknown   = &models.Bundle{State: "unknown"}
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
			Label:               "draft",
			TargetState:         Draft,
			AllowedSourceStates: []string{"in_review", "approved"},
		},
		{
			Label:               "in_review",
			TargetState:         InReview,
			AllowedSourceStates: []string{"draft", "approved"},
		},
		{
			Label:               "approved",
			TargetState:         Approved,
			AllowedSourceStates: []string{"in_review"},
		},
		{
			Label:               "published",
			TargetState:         Published,
			AllowedSourceStates: []string{"approved"},
		},
	}
}

func TestGetStateByName(t *testing.T) {
	Convey("Given a state name", t, func() {
		Convey("When the state name is 'draft'", func() {
			state, found := getStateByName("draft")

			Convey("Then it should return the Draft state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "draft")
			})
		})

		Convey("When the state name is 'in_review'", func() {
			state, found := getStateByName("in_review")

			Convey("Then it should return the InReview state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "in_review")
			})
		})

		Convey("When the state name is 'approved'", func() {
			state, found := getStateByName("approved")

			Convey("Then it should return the Approved state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "approved")
			})
		})

		Convey("When the state name is 'published'", func() {
			state, found := getStateByName("published")

			Convey("Then it should return the Published state", func() {
				So(found, ShouldBeTrue)
				So(state, ShouldNotBeNil)
				So(state.Name, ShouldEqual, "published")
			})
		})

		Convey("When the state name is 'unknown'", func() {
			state, found := getStateByName("unknown")

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

	Convey("When transitioning from 'draft' to 'in_review'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'in_review' to 'approved' with bundle contents approved", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReviewAndContentsApproved, bundleUpdateWithStateApproved)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'in_review' to 'draft'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReview, bundleUpdateWithStateDraft)
		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'approved' to 'published'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStatePublished)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'approved' to 'in_review'", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateApproved, bundleUpdateWithStateInReview)

		Convey("Then the transition should be successful", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When transitioning from 'approved' to 'draft'", t, func() {
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

	Convey("When transitioning from 'in_review' to 'approved' with bundle contents not approved", t, func() {
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateInReviewAndContentsNotApproved, bundleUpdateWithStateApproved)
		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not all bundle contents are approved")
		})
	})

	Convey("When the state machine has a transition that contains an invalid state", t, func() {
		stateMachineBundleAPI.StateMachine.transitions["unknown"] = []string{"draft"}
		err := stateMachine.Transition(ctx, stateMachineBundleAPI, currentBundleWithStateDraft, bundleUpdateWithStateUnknown)

		Convey("Then the transition should fail", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "incorrect state value")
		})
	})
}
