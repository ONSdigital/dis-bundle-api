package application

import (
	"context"
	"errors"
	"slices"

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
	//var valid bool

	match := false
	var nextState *State
	var ok bool

	for state, transitions := range sm.transitions {
		if state == targetState.String() {
			for i := range transitions {
				if currentBundle.State.String() == transitions[i] {
					//continue

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

// IsValidTransition validates whether the sourceState can transition to the targetState. If not, an error is returned
func (sm *StateMachine) IsValidTransition(ctx context.Context, sourceState, targetState *models.BundleState) error {
	allowedSourceStates, exists := sm.transitions[targetState.String()]

	if !exists {
		return apierrors.ErrInvalidTransition
	}

	if !slices.Contains(allowedSourceStates, sourceState.String()) {
		return apierrors.ErrInvalidTransition
	}

	return nil
}

// func (sm *StateMachine) TransitionBundle(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, bundle *models.Bundle, targetState *models.BundleState, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
// 	fmt.Println("Entering transition bundle at: ", time.Now().String()+" for bundle id "+bundle.ID)
// 	fmt.Println("Number of go routines at start transition: ", runtime.NumGoroutine())

// 	if err := sm.IsValidTransition(ctx, &bundle.State, targetState); err != nil {
// 		return nil, err
// 	}

// 	contents, err := stateMachineBundleAPI.Datastore.GetBundleContentsForBundle(ctx, bundle.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if contents == nil || len(*contents) == 0 {
// 		return nil, apierrors.ErrBundleHasNoContentItems
// 	}

// 	if targetState.String() == models.BundleStateApproved.String() || targetState.String() == models.BundleStatePublished.String() {
// 		// cores := runtime.NumCPU()
// 		// runtime.GOMAXPROCS(cores)
// 		var wg sync.WaitGroup
// 		//	var wg sync.WaitGroup
// 		// numWorkers := len(*contents)
// 		// wg.Add(numWorkers)
// 		ch := make(chan int)
// 		fmt.Println("Starting go routines at: ", time.Now().String())
// 		for index := range *contents {
// 			wg.Add(1)
// 			//defer wg.Done()
// 			fmt.Println("Starting loop: "+strconv.Itoa(index)+"at: ", time.Now().String())
// 			contentItem := &(*contents)[index]
// 			go sm.transitionContentItem(ctx, contentItem, stateMachineBundleAPI, targetState, authEntityData, ch, &wg)
// 			fmt.Println("Ending loop: "+strconv.Itoa(index)+"at: ", time.Now().String())

// 		}
// 		wg.Wait()
// 		close(ch)

// 	}

// 	bundle.State = *targetState
// 	bundle.LastUpdatedBy.Email = authEntityData.GetUserEmail()

// 	updatedBundle, err := stateMachineBundleAPI.Datastore.UpdateBundle(ctx, bundle.ID, bundle)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err = stateMachineBundleAPI.CreateEvent(ctx, authEntityData, models.ActionUpdate, updatedBundle, nil); err != nil {
// 		log.Error(ctx, "failed to create event", err, log.Data{"bundle_id": updatedBundle.ID})
// 		return nil, err
// 	}
// 	fmt.Println("Exiting transition bundle at: ", time.Now().String())
// 	fmt.Println("Number of go routines at end transition: ", runtime.NumGoroutine())

// 	return updatedBundle, nil
// }

// func (sm *StateMachine) transitionContentItem(ctx context.Context, contentItem *models.ContentItem, stateMachineBundleAPI *StateMachineBundleAPI, targetState *models.BundleState, authEntityData *models.AuthEntityData, ch chan int, wg *sync.WaitGroup) error {
// 	defer wg.Done()
// 	// if err := stateMachineBundleAPI.updateVersionStateForContentItem(ctx, contentItem, targetState, authEntityData.Headers); err != nil {
// 	// 	return err
// 	// }

// 	fmt.Println("ABOUT TO EXECUTE PUT DATASET STUFF")
// 	if err := sm.datasetAPIClient.PutVersionState(ctx, authEntityData.Headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, strconv.Itoa(contentItem.Metadata.VersionID), strings.ToLower(targetState.String())); err != nil {
// 		return err
// 	}

// 	if err := stateMachineBundleAPI.Datastore.UpdateContentItemState(ctx, contentItem.ID, targetState.String()); err != nil {
// 		return err
// 	}

// 	if err := stateMachineBundleAPI.CreateEvent(ctx, authEntityData, models.ActionUpdate, nil, contentItem); err != nil {
// 		log.Error(ctx, "failed to create event", err, log.Data{"bundle_id": contentItem.BundleID, "content_item_id": contentItem.ID})
// 		return err
// 	}

// 	ch <- 42

// 	return nil
// }
