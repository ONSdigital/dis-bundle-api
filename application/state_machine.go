package application

import (
	"context"
	"errors"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type StateMachine struct {
	states      map[string]State
	transitions map[string][]string
	datastore   store.Datastore
	ctx         context.Context
}

type Transition struct {
	Label               string
	TargetState         State
	AllowedSourceStates []string
}

type State struct {
	Name string
}

func (s State) String() string {
	return s.Name
}

func getStateByName(stateName string) (*State, bool) {
	switch stateName {
	case "draft":
		return &Draft, true
	case "in_review":
		return &InReview, true
	case "approved":
		return &Approved, true
	case "published":
		return &Published, true
	default:
		return nil, false
	}
}

func NewStateMachine(ctx context.Context, states []State, transitions []Transition, datastore store.Datastore) *StateMachine {
	statesMap := make(map[string]State)
	for _, state := range states {
		statesMap[state.String()] = state
	}

	transitionsMap := make(map[string][]string)
	for _, transition := range transitions {
		transitionsMap[transition.TargetState.String()] = transition.AllowedSourceStates
	}

	StateMachine := &StateMachine{
		states:      statesMap,
		transitions: transitionsMap,
		datastore:   datastore,
		ctx:         ctx,
	}

	return StateMachine
}

func (sm *StateMachine) Transition(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, currentBundle, bundleUpdate *models.Bundle) error {
	var valid bool

	match := false

	for state, transitions := range sm.transitions {
		if state == bundleUpdate.State {
			for i := range transitions {
				if currentBundle.State != transitions[i] {
					continue
				}
				match = true

				_, valid = getStateByName(state)
				if !valid {
					return errors.New("incorrect state value")
				}

				if currentBundle.State == InReview.String() && bundleUpdate.State == Approved.String() {
					allApproved := checkAllBundleContentsAreApproved(currentBundle.Contents)
					if !allApproved {
						return errors.New("not all bundle contents are approved")
					}
				}
				break
			}
		}
	}

	if !match {
		return errors.New("state not allowed to transition")
	}

	return nil
}
