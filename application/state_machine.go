package application

import (
	"context"
	"fmt"
	"net/http"

	"slices"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type TransitionHandler func(ctx context.Context, api *StateMachineBundleAPI, r *http.Request, bundle *models.Bundle, targetState models.BundleState) *models.Error

type StateMachine struct {
	states      map[string]models.BundleState
	transitions map[models.BundleState][]Transition
	datastore   store.Datastore
	ctx         context.Context
}

type Transition struct {
	Label               string
	TargetState         models.BundleState
	AllowedSourceStates []models.BundleState
	Handler             TransitionHandler
}

func NewStateMachine(ctx context.Context, states []models.BundleState, transitions []Transition, datastore store.Datastore) *StateMachine {
	statesMap := make(map[string]models.BundleState)
	for _, state := range states {
		statesMap[state.String()] = state
	}

	transitionsMap := make(map[models.BundleState][]Transition)
	for _, transition := range transitions {
		transitionsMap[transition.TargetState] = append(transitionsMap[transition.TargetState], transition)
	}

	StateMachine := &StateMachine{
		states:      statesMap,
		transitions: transitionsMap,
		datastore:   datastore,
		ctx:         ctx,
	}

	return StateMachine
}

func (sm *StateMachine) Transition(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, r *http.Request, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	transitionHandler, err := sm.getTransitionHandler(bundle, targetState)

	if err != nil {
		return err
	}

	handler := *transitionHandler

	err = handler(ctx, stateMachineBundleAPI, r, bundle, targetState)
	return err
}

func (sm *StateMachine) canTransitionFromCurrentState(allowedTransitions []models.BundleState, currentBundle *models.Bundle) bool {
	return slices.Contains(allowedTransitions, *currentBundle.State)
}

func (sm *StateMachine) getTransitionHandler(bundle *models.Bundle, targetState models.BundleState) (*TransitionHandler, *models.Error) {
	transitions, exists := sm.transitions[targetState]
	if !exists {
		return nil, models.CreateModelError(models.CodeBadRequest, fmt.Sprintf("incorrect state value: no transitions found for state %s", targetState))
	}

	for _, transition := range transitions {
		if sm.canTransitionFromCurrentState(transition.AllowedSourceStates, bundle) {
			return &transition.Handler, nil
		}
	}

	return nil, models.CreateModelError(models.CodeBadRequest, fmt.Sprintf("no valid transition from %s to %s", *bundle.State, targetState))
}
