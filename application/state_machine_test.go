package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsSDK "github.com/ONSdigital/dp-permissions-api/sdk"

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

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})

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

func TestTransitionBundle_Success(t *testing.T) {
	fromState := models.BundleStateApproved
	targetState := models.BundleStatePublished

	mockBundle := &models.Bundle{
		ID:    mockBundleID,
		State: fromState,
		LastUpdatedBy: &models.User{
			Email: "email@ons.com",
		},
	}

	mockVersions := []*datasetAPIModels.Version{
		{
			ID:        "valid-version-1",
			Version:   1,
			DatasetID: "dataset-id-1",
			Edition:   "edition-id-1",
			State:     strings.ToLower(mockBundle.State.String()),
		},
		{
			ID:        "valid-version-2",
			Version:   1,
			DatasetID: "dataset-id-2",
			Edition:   "edition-id-2",
			State:     strings.ToLower(mockBundle.State.String()),
		},
	}

	validContentItemState := models.State(mockBundle.State.String())

	mockContentItems := []*models.ContentItem{
		{
			ID:       "valid-content-item",
			BundleID: mockBundleID,
			State:    &validContentItemState,
			Metadata: models.Metadata{
				DatasetID: mockVersions[0].DatasetID,
				EditionID: mockVersions[0].Edition,
				VersionID: mockVersions[0].Version,
			},
		},
		{
			ID:       "another-valid-content-item",
			BundleID: mockBundleID,
			State:    &validContentItemState,
			Metadata: models.Metadata{
				DatasetID: mockVersions[1].DatasetID,
				EditionID: mockVersions[1].Edition,
				VersionID: mockVersions[1].Version,
			},
		},
	}
	mockAuthEntityData := models.AuthEntityData{
		EntityData: &permissionsSDK.EntityData{
			UserID: mockUserID,
		},
		Headers: datasetAPISDK.Headers{
			ServiceToken:    mockServiceToken,
			UserAccessToken: mockServiceToken,
		},
	}

	var createdEvents []*models.Event

	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	getVersionFunc := func(_ context.Context, _ datasetAPISDK.Headers, datasetID, editionID, versionID string) (*datasetAPIModels.Version, error) {
		for index := range mockVersions {
			mockVersion := mockVersions[index]

			mockVersionID := strconv.Itoa(mockVersion.Version)

			if mockVersion.DatasetID == datasetID && mockVersion.Edition == editionID && versionID == mockVersionID {
				return mockVersion, nil
			}
		}

		return nil, errors.New("not found version")
	}

	mockedDatastore := &storetest.StorerMock{
		GetBundleContentsForBundleFunc: func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
			if bundleID == mockBundleID {
				contentItems := make([]models.ContentItem, len(mockContentItems))
				for index := range contentItems {
					contentItems[index] = *mockContentItems[index]
				}
				return &contentItems, nil
			}

			return nil, nil
		},
		UpdateBundleFunc: func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
			if id != mockBundleID {
				return nil, apierrors.ErrBundleNotFound
			}

			mockBundle.State = update.State
			mockBundle.LastUpdatedBy = update.LastUpdatedBy
			mockBundle.ETag = update.ETag

			return mockBundle, nil
		},
		CreateBundleEventFunc: func(ctx context.Context, event *models.Event) error {
			createdEvents = append(createdEvents, event)
			return nil
		},
		UpdateContentItemStateFunc: func(ctx context.Context, contentItemID, state string) error {
			for index := range mockContentItems {
				contentItem := mockContentItems[index]

				if contentItem.ID == contentItemID {
					updatedState := models.State(state)
					contentItem.State = &updatedState
					return nil
				}
			}

			return errors.New("not found content item")
		},
	}
	mockdatasetAPIClient := datasetAPISDKMock.ClienterMock{
		GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
			version, err := getVersionFunc(ctx, headers, datasetID, editionID, versionID)

			if err != nil {
				return datasetAPIModels.Version{}, nil
			}

			return *version, nil
		},
		PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
			version, err := getVersionFunc(ctx, headers, datasetID, editionID, versionID)

			if err != nil {
				return err
			}

			version.State = state
			return nil
		},
	}

	stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
	stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockdatasetAPIClient)
	bundle, err := stateMachine.TransitionBundle(ctx, stateMachineBundleAPI, mockBundle, &targetState, &mockAuthEntityData)

	Convey("When TransitionBundle is called with a valid transition and valid bundle", t, func() {
		So(err, ShouldBeNil)
		Convey("Then the bundle should be updated", func() {
			So(mockBundle.State.String(), ShouldEqual, targetState.String())
			So(mockBundle.LastUpdatedBy.Email, ShouldEqual, mockUserID)

			Convey("And the returned bundle should match", func() {
				So(mockBundle, ShouldEqual, bundle)
			})
		})

		contentItemsThatShouldBeUpdated := []*models.ContentItem{mockContentItems[0], mockContentItems[1]}

		Convey("And the content items should be updated if the state matched", func() {
			for _, contentItem := range contentItemsThatShouldBeUpdated {
				So(contentItem.State.String(), ShouldEqual, targetState.String())
			}
		})

		versionsThatShouldBeUpdated := []int{0, 1}

		Convey("And the versions should be updated if the state matched", func() {
			for index := range versionsThatShouldBeUpdated {
				mockVersion := mockVersions[versionsThatShouldBeUpdated[index]]

				So(strings.ToLower(mockVersion.State), ShouldEqual, strings.ToLower(targetState.String()))
			}
		})

		Convey("And events should be created", func() {
			So(createdEvents, ShouldHaveLength, 3)

			validateCreatedEvents(contentItemsThatShouldBeUpdated, createdEvents, mockBundle)
		})
	})
}

func validateCreatedEvents(contentItemsThatShouldBeUpdated []*models.ContentItem, createdEvents []*models.Event, mockBundle *models.Bundle) {
	var bundleEvents []*models.Event
	var contentItemEvents []*models.Event

	expectedContentItems := map[string]*models.ContentItem{
		contentItemsThatShouldBeUpdated[0].ID: contentItemsThatShouldBeUpdated[0],
		contentItemsThatShouldBeUpdated[1].ID: contentItemsThatShouldBeUpdated[1],
	}

	for index := range createdEvents {
		event := createdEvents[index]

		So(event.Action, ShouldEqual, models.ActionUpdate)
		if event.ContentItem == nil {
			bundleEvents = append(bundleEvents, event)
			So(event.Bundle.ID, ShouldEqual, mockBundle.ID)
		} else if event.ContentItem != nil {
			_, exists := expectedContentItems[event.ContentItem.ID]
			So(exists, ShouldBeTrue)
			delete(expectedContentItems, event.ContentItem.ID)
			contentItemEvents = append(contentItemEvents, event)
		} else {
			panic("both bundle and content item were nil in the event")
		}
	}

	So(bundleEvents, ShouldHaveLength, 1)
	So(contentItemEvents, ShouldHaveLength, 2)
}

func TestTransitionBundle_Failure(t *testing.T) {
	fromState := models.BundleStateApproved
	targetState := models.BundleStatePublished

	mockBundle := &models.Bundle{
		ID:    mockBundleID,
		State: fromState,
		LastUpdatedBy: &models.User{
			Email: "email@ons.com",
		},
	}

	mockVersions := []*datasetAPIModels.Version{
		{
			ID:        "valid-version-1",
			Version:   1,
			DatasetID: "dataset-id-1",
			Edition:   "edition-id-1",
			State:     strings.ToLower(mockBundle.State.String()),
		},
		{
			ID:        "valid-version-2",
			Version:   1,
			DatasetID: "dataset-id-2",
			Edition:   "edition-id-2",
			State:     strings.ToLower(mockBundle.State.String()),
		},
	}

	validContentItemState := models.State(mockBundle.State.String())

	mockContentItems := []*models.ContentItem{
		{
			ID:       "valid-content-item",
			BundleID: mockBundleID,
			State:    &validContentItemState,
			Metadata: models.Metadata{
				DatasetID: mockVersions[0].DatasetID,
				EditionID: mockVersions[0].Edition,
				VersionID: mockVersions[0].Version,
			},
		},
		{
			ID:       "another-valid-content-item",
			BundleID: mockBundleID,
			State:    &validContentItemState,
			Metadata: models.Metadata{
				DatasetID: mockVersions[1].DatasetID,
				EditionID: mockVersions[1].Edition,
				VersionID: mockVersions[1].Version,
			},
		},
	}

	mockAuthEntityData := models.AuthEntityData{
		EntityData: &permissionsSDK.EntityData{
			UserID: mockUserID,
		},
		Headers: datasetAPISDK.Headers{
			ServiceToken:    mockServiceToken,
			UserAccessToken: mockServiceToken,
		},
	}

	var createdEvents []*models.Event

	ctx := context.Background()

	states := getMockStates()
	transitions := getMockTransitions()

	getVersionFunc := func(_ context.Context, _ datasetAPISDK.Headers, datasetID, editionID, versionID string) (*datasetAPIModels.Version, error) {
		for index := range mockVersions {
			mockVersion := mockVersions[index]

			mockVersionID := strconv.Itoa(mockVersion.Version)

			if mockVersion.DatasetID == datasetID && mockVersion.Edition == editionID && versionID == mockVersionID {
				return mockVersion, nil
			}
		}

		return nil, errors.New("not found version")
	}

	getBundleContentsFunc := func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
		if bundleID == mockBundleID {
			contentItems := make([]models.ContentItem, len(mockContentItems))
			for index := range contentItems {
				contentItems[index] = *mockContentItems[index]
			}
			return &contentItems, nil
		}

		return nil, nil
	}

	updateBundleFunc := func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
		if id != mockBundleID {
			return nil, apierrors.ErrBundleNotFound
		}

		mockBundle.State = update.State
		mockBundle.LastUpdatedBy = update.LastUpdatedBy
		mockBundle.ETag = update.ETag

		return mockBundle, nil
	}

	createBundleEventsFunc := func(ctx context.Context, event *models.Event) error {
		createdEvents = append(createdEvents, event)
		return nil
	}

	updateContentItemFunc := func(ctx context.Context, contentItemID, state string) error {
		for index := range mockContentItems {
			contentItem := mockContentItems[index]

			if contentItem.ID == contentItemID {
				updatedState := models.State(state)
				contentItem.State = &updatedState
				return nil
			}
		}

		return errors.New("not found content item")
	}
	mockedDatastore := &storetest.StorerMock{
		GetBundleContentsForBundleFunc: getBundleContentsFunc,
		UpdateBundleFunc:               updateBundleFunc,
		CreateBundleEventFunc:          createBundleEventsFunc,
		UpdateContentItemStateFunc:     updateContentItemFunc,
	}

	mockdatasetAPIClient := datasetAPISDKMock.ClienterMock{
		GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (datasetAPIModels.Version, error) {
			version, err := getVersionFunc(ctx, headers, datasetID, editionID, versionID)

			if err != nil {
				return datasetAPIModels.Version{}, nil
			}

			return *version, nil
		},
		PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
			version, err := getVersionFunc(ctx, headers, datasetID, editionID, versionID)

			if err != nil {
				return err
			}

			version.State = state
			return nil
		},
	}

	t.Run("When attempting a valid bundle transition state/But an error is returned attempting to get content items", func(t *testing.T) {
		dbError := errors.New("database error")
		mockedDatastore.GetBundleContentsForBundleFunc = func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
			return nil, dbError
		}

		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockdatasetAPIClient)
		bundle, err := stateMachine.TransitionBundle(ctx, stateMachineBundleAPI, mockBundle, &targetState, &mockAuthEntityData)
		Convey("Then", t, func() {
			Convey("the error should be returned", func() {
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, dbError)
			})

			Convey("And the bundle returned should be nil", func() {
				So(bundle, ShouldBeNil)
			})

			Convey("And the bundle state should not be updated", func() {
				So(mockBundle.State.String(), ShouldNotEqual, targetState.String())
			})
		})
	})

	t.Run("When attempting a valid bundle transition state/But no content items are found", func(t *testing.T) {
		mockedDatastore.GetBundleContentsForBundleFunc = func(ctx context.Context, bundleID string) (*[]models.ContentItem, error) {
			return nil, nil
		}

		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockdatasetAPIClient)
		bundle, err := stateMachine.TransitionBundle(ctx, stateMachineBundleAPI, mockBundle, &targetState, &mockAuthEntityData)
		Convey("Then", t, func() {
			Convey("Then a not found error should be returned", func() {
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, apierrors.ErrBundleHasNoContentItems)
			})

			Convey("And the bundle returned should be nil", func() {
				So(bundle, ShouldBeNil)
			})

			Convey("And the bundle state should not be updated", func() {
				So(mockBundle.State.String(), ShouldNotEqual, targetState.String())
			})
		})
	})

	t.Run("When attempting a valid bundle transition state/But an error occurs updating the bundle", func(t *testing.T) {
		mockedDatastore.GetBundleContentsForBundleFunc = getBundleContentsFunc
		updateBundleError := errors.New("update bundle error")
		mockedDatastore.UpdateBundleFunc = func(ctx context.Context, id string, update *models.Bundle) (*models.Bundle, error) {
			return nil, updateBundleError
		}
		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockdatasetAPIClient)

		bundleInstance := *mockBundle

		bundle, err := stateMachine.TransitionBundle(ctx, stateMachineBundleAPI, &bundleInstance, &targetState, &mockAuthEntityData)
		Convey("Then", t, func() {
			Convey("the error should be returned", func() {
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, updateBundleError)
			})

			Convey("the bundle returned should be nil", func() {
				So(bundle, ShouldBeNil)
			})

			Convey("And the bundle state should not be updated", func() {
				So(mockBundle.State.String(), ShouldNotEqual, targetState.String())
			})
		})
	})

	t.Run("When attempting a valid bundle transition state/But an error occurs creating the bundle event", func(t *testing.T) {
		mockedDatastore.GetBundleContentsForBundleFunc = getBundleContentsFunc
		mockedDatastore.UpdateBundleFunc = updateBundleFunc

		createEventError := errors.New("create event error")

		// previous test would have published all versions so we need to reset them
		for index := range mockVersions {
			mockVersions[index].State = strings.ToLower(fromState.String())
		}

		mockedDatastore.CreateBundleEventFunc = func(ctx context.Context, event *models.Event) error {
			// To avoid throwing the error when content items are updated
			if event.Bundle != nil {
				return createEventError
			}
			return nil
		}

		stateMachine := NewStateMachine(ctx, states, transitions, store.Datastore{Backend: mockedDatastore})
		stateMachineBundleAPI := Setup(store.Datastore{Backend: mockedDatastore}, stateMachine, &mockdatasetAPIClient)

		bundleInstance := *mockBundle
		bundle, err := stateMachine.TransitionBundle(ctx, stateMachineBundleAPI, &bundleInstance, &targetState, &mockAuthEntityData)
		Convey("Then", t, func() {
			Convey("the error should be returned", func() {
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, createEventError)
			})

			Convey("the bundle returned should be nil", func() {
				So(bundle, ShouldBeNil)
			})

			Convey("And the bundle state should be updated", func() {
				So(mockBundle.State.String(), ShouldEqual, targetState.String())
			})
		})
	})
}
