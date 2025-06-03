package application

import (
	"context"
	"fmt"

	"slices"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type TransitionHandler func(ctx context.Context, api *StateMachineBundleAPI, bundle *models.Bundle, targetState models.BundleState) *models.Error

type StateMachineTransition struct {
	sourceState       models.BundleState
	targetState       models.BundleState
	transitionHandler TransitionHandler
}

func CreateStateMachineTransition(sourceState, targetState models.BundleState, transitionHandler TransitionHandler) StateMachineTransition {
	return StateMachineTransition{
		sourceState:       sourceState,
		targetState:       targetState,
		transitionHandler: transitionHandler,
	}
}

type StateMachine struct {
	states             map[string]models.BundleState
	transitions        map[models.BundleState][]models.BundleState
	transitionHandlers []StateMachineTransition
	datastore          store.Datastore
	ctx                context.Context
}

type Transition struct {
	Label               string
	TargetState         models.BundleState
	AllowedSourceStates []models.BundleState
}

func NewStateMachine(ctx context.Context, states []models.BundleState, transitions []Transition, datastore store.Datastore, transitionHandlers []StateMachineTransition) *StateMachine {
	statesMap := make(map[string]models.BundleState)
	for _, state := range states {
		statesMap[state.String()] = state
	}

	transitionsMap := make(map[models.BundleState][]models.BundleState)
	for _, transition := range transitions {
		transitionsMap[transition.TargetState] = transition.AllowedSourceStates
	}

	StateMachine := &StateMachine{
		states:             statesMap,
		transitions:        transitionsMap,
		datastore:          datastore,
		ctx:                ctx,
		transitionHandlers: transitionHandlers,
	}

	return StateMachine
}

func (sm *StateMachine) Transition(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	allowedTransitions, exists := sm.transitions[targetState]
	if !exists {
		return models.CreateModelError(models.CodeBadRequest, fmt.Sprintf("incorrect state value: no transitions found for state %s", targetState))
	}

	if !sm.canTransitionFromCurrentState(allowedTransitions, bundle) {
		return models.CreateModelError(models.CodeBadRequest, fmt.Sprintf("state %s not allowed to transition to %s", bundle.State, targetState))
	}

	transitionHandler, err := sm.getTransitionHandler(bundle, targetState)

	if err != nil {
		return models.CreateModelError(models.CodeBadRequest, err.Error())
	}

	handler := *transitionHandler
	handlerErr := handler(ctx, stateMachineBundleAPI, bundle, targetState)
	return handlerErr
}

func (sm *StateMachine) canTransitionFromCurrentState(allowedTransitions []models.BundleState, currentBundle *models.Bundle) bool {
	return slices.Contains(allowedTransitions, *currentBundle.State)
}

func (sm *StateMachine) getTransitionHandler(bundle *models.Bundle, targetState models.BundleState) (*TransitionHandler, error) {
	for _, transition := range sm.transitionHandlers {
		if transition.sourceState == *bundle.State && transition.targetState == targetState {
			return &transition.transitionHandler, nil
		}
	}

	return nil, fmt.Errorf("could not find a matching state machine transition handler for updating %s to %s", bundle.State, targetState)
}
