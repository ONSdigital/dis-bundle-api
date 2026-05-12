package application

import (
	"context"
	"errors"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

type StateMachine struct {
	states           map[string]State
	transitions      map[string][]string
	datastore        store.Datastore
	ctx              context.Context
	datasetAPIClient datasetAPISDK.Clienter
}

type Transition struct {
	Label               string
	TargetState         State
	AllowedSourceStates []string
}

type State struct {
	Name      string
	EnterFunc func(ctx context.Context, smBundle StateMachineBundleAPI, bundle *models.Bundle, authEntityData *models.AuthEntityData) (*models.Bundle, error)
}

func (s State) String() string {
	return s.Name
}

func NewStateMachine(ctx context.Context, states []State, transitions []Transition, datastore store.Datastore, datasetAPIClient datasetAPISDK.Clienter) *StateMachine {
	statesMap := make(map[string]State)
	for _, state := range states {
		statesMap[state.String()] = state
	}

	transitionsMap := make(map[string][]string)
	for _, transition := range transitions {
		transitionsMap[transition.TargetState.String()] = transition.AllowedSourceStates
	}

	StateMachine := &StateMachine{
		states:           statesMap,
		transitions:      transitionsMap,
		datastore:        datastore,
		ctx:              ctx,
		datasetAPIClient: datasetAPIClient,
	}

	return StateMachine
}

func castStateToState(state string) (*State, bool) {
	switch s := state; s {
	case "PUBLISHED":
		return &Published, true
	case "IN_REVIEW":
		return &InReview, true
	case "APPROVED":
		return &Approved, true
	case "DRAFT":
		return &Draft, true
	default:
		return nil, false
	}
}

func (sm *StateMachine) Transition(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, currentBundle *models.Bundle, targetState models.BundleState, authEntityData models.AuthEntityData) (*models.Bundle, error) {
	match := false
	var nextState *State
	var ok bool

	for state, transitions := range sm.transitions {
		if state == targetState.String() {
			for i := range transitions {
				if currentBundle.State.String() == transitions[i] {
					match = true
					nextState, ok = castStateToState(targetState.String())
					if !ok {
						return nil, errors.New("incorrect state value")
					}
					break
				}
			}
		}
	}

	if !match {
		return nil, apierrors.ErrInvalidTransition
	}

	updatedBundle, err := nextState.EnterFunc(ctx, *stateMachineBundleAPI, currentBundle, &authEntityData)
	if err != nil {
		return nil, err
	}
	return updatedBundle, nil
}
