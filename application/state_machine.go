package application

import (
	"context"
	"errors"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/log.go/v2/log"
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
	case "DRAFT":
		return &Draft, true
	case "IN_REVIEW":
		return &InReview, true
	case "APPROVED":
		return &Approved, true
	case "PUBLISHED":
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
		if state == bundleUpdate.State.String() {
			for i := range transitions {
				if currentBundle.State.String() != transitions[i] {
					continue
				}
				match = true

				_, valid = getStateByName(state)
				if !valid {
					return errors.New("incorrect state value")
				}

				if currentBundle.State.String() == InReview.String() && bundleUpdate.State.String() == Approved.String() {
					allBundleContentsApproved, err := stateMachineBundleAPI.CheckAllBundleContentsAreApproved(ctx, currentBundle.ID)
					if err != nil {
						log.Error(ctx, "error checking if all bundle contents are approved", err, log.Data{"bundle_id": currentBundle.ID})
						return err
					}

					if !allBundleContentsApproved {
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
