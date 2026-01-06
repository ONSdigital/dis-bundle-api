package application

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"

	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
)

// CreateBundlePolicies creates a new policy for each preview team with the provided role
func (s *StateMachineBundleAPI) CreateBundlePolicies(ctx context.Context, authToken string, previewTeams *[]models.PreviewTeam, role models.Role) error {
	if previewTeams == nil || len(*previewTeams) == 0 {
		return nil
	}

	if !models.ValidateRole(role) {
		return apierrors.ErrInvalidRole
	}

	for _, team := range *previewTeams {
		policyInfo := permissionsAPIModels.PolicyInfo{
			Entities: []string{
				"groups/" + team.ID,
			},
			Role: role.String(),
		}

		_, err := s.PermissionsAPIClient.PostPolicyWithID(ctx, permissionsAPISDK.Headers{Authorization: authToken}, team.ID, policyInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdatePolicyConditionsForContentItem updates policy conditions for all preview teams in a bundle
// when a content item is added (isAdd=true) or removed (isAdd=false)
func (s *StateMachineBundleAPI) UpdatePolicyConditionsForContentItem(ctx context.Context, authToken string, bundle *models.Bundle, contentItem *models.ContentItem, isAdd bool) error {
	if bundle.PreviewTeams == nil || len(*bundle.PreviewTeams) == 0 {
		return nil
	}

	for _, team := range *bundle.PreviewTeams {
		if err := s.updatePolicyConditionForTeam(ctx, authToken, team.ID, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, isAdd); err != nil {
			return err
		}
	}

	return nil
}

// updatePolicyConditionForTeam updates the policy condition for a single preview team
func (s *StateMachineBundleAPI) updatePolicyConditionForTeam(ctx context.Context, authToken, teamID, datasetID, editionID string, isAdd bool) error {
	policy, err := s.PermissionsAPIClient.GetPolicy(ctx, teamID)
	if err != nil {
		return err
	}

	datasetValue := datasetID
	datasetEditionValue := datasetID + "/" + editionID

	if isAdd {
		if policy.Condition.Attribute == "" {
			policy.Condition = permissionsAPIModels.Condition{
				Attribute: "dataset_edition",
				Operator:  "StringEquals",
				Values:    []string{datasetValue, datasetEditionValue},
			}
		} else {
			policy.Condition.Values = append(policy.Condition.Values, datasetValue, datasetEditionValue)
		}
	} else {
		policy.Condition.Values = removeConditionValues(policy.Condition.Values, datasetValue, datasetEditionValue)
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+authToken)
	opts := permissionsAPISDK.Options{Headers: headers}
	authClient := permissionsAPISDK.NewClientWithOptions(s.PermissionsAPIURL, opts)

	err = authClient.PutPolicy(ctx, teamID, *policy)
	if err != nil {
		return err
	}

	return nil
}

// removeValuesFromSlice removes specific values from the Condition array
func removeConditionValues(values []string, toRemove ...string) []string {
	result := []string{}
	removeMap := make(map[string]bool)

	for _, val := range toRemove {
		removeMap[val] = true
	}

	for _, val := range values {
		if !removeMap[val] {
			result = append(result, val)
		}
	}

	return result
}
