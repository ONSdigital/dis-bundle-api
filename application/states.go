package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// list of states for the state machine
var (
	Draft = State{
		Name: "DRAFT",
		EnterFunc: func(ctx context.Context, sm *StateMachine, smAPI *StateMachineBundleAPI, bundle *models.Bundle, auth *models.AuthEntityData) error {
			fmt.Println("enter function called DRAFT")
			return nil
		},
	}

	InReview = State{
		Name: "IN_REVIEW",
		EnterFunc: func(ctx context.Context, sm *StateMachine, smAPI *StateMachineBundleAPI, bundle *models.Bundle, auth *models.AuthEntityData) error {
			return nil
		},
	}

	Approved = State{
		Name: "APPROVED",
		EnterFunc: func(ctx context.Context, sm *StateMachine, smAPI *StateMachineBundleAPI, bundle *models.Bundle, auth *models.AuthEntityData) error {
			// Check all contents are approved
			allApproved, err := smAPI.CheckAllBundleContentsAreApproved(ctx, bundle.ID)
			fmt.Println("allApproved is :", allApproved)
			if err != nil {
				log.Error(ctx, "error checking if all bundle contents are approved", err, log.Data{"bundle_id": bundle.ID})
				return err
			}
			if !allApproved {
				return errors.New("not all bundle contents are approved")
			}

			// Cascade to content items
			contents, err := smAPI.Datastore.GetBundleContentsForBundle(ctx, bundle.ID)
			if err != nil {
				return err
			}
			for i := range *contents {
				if err := sm.transitionContentItem(ctx, &(*contents)[i], smAPI, models.BundleStateApproved, auth); err != nil {
					log.Warn(ctx, fmt.Sprintf("Error occurred transitioning content item for bundle: %s", err.Error()),
						log.Data{"bundle-id": bundle.ID, "content-item-id": (*contents)[i].ID})
					return err
				}
			}
			return nil
		},
	}

	Published = State{
		Name: "PUBLISHED",
		EnterFunc: func(ctx context.Context, sm *StateMachine, smAPI *StateMachineBundleAPI, bundle *models.Bundle, auth *models.AuthEntityData) error {
			// Ensure bundle has content
			contents, err := smAPI.Datastore.GetBundleContentsForBundle(ctx, bundle.ID)
			if err != nil {
				return err
			}
			if contents == nil || len(*contents) == 0 {
				return apierrors.ErrBundleHasNoContentItems
			}

			// Cascade to content items
			for i := range *contents {
				if err := sm.transitionContentItem(ctx, &(*contents)[i], smAPI, models.BundleStatePublished, auth); err != nil {
					log.Warn(ctx, fmt.Sprintf("Error occurred transitioning content item for bundle: %s", err.Error()),
						log.Data{"bundle-id": bundle.ID, "content-item-id": (*contents)[i].ID})
					return err
				}
			}
			return nil
		},
	}
)
