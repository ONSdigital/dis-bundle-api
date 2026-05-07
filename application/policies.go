package application

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"

	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
)

const (
	conditionAttributeDatasetEdition = "dataset_edition"
	conditionOperatorStringEquals    = "StringEquals"
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
		policyExists, err := s.CheckPolicyExists(ctx, authToken, team.ID)
		if err != nil {
			return err
		}
		if policyExists {
			continue
		}

		policyInfo := permissionsAPIModels.PolicyInfo{
			Entities: []string{
				"groups/" + team.ID,
			},
			Role: role.String(),
			Condition: permissionsAPIModels.Condition{
				Attribute: conditionAttributeDatasetEdition,
				Operator:  conditionOperatorStringEquals,
			},
		}

		_, err = s.PermissionsAPIClient.PostPolicyWithID(ctx, team.ID, policyInfo, permissionsAPISDK.Headers{Authorization: authToken})
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckPolicyExists checks if a policy with the given ID exists
func (s *StateMachineBundleAPI) CheckPolicyExists(ctx context.Context, authToken, policyID string) (bool, error) {
	_, err := s.PermissionsAPIClient.GetPolicy(ctx, policyID, permissionsAPISDK.Headers{Authorization: authToken})
	if err != nil {
		// Permissions API will return an error containing "404" if the policy is not found
		// Note: The Permissions API SDK does not currently provide an alternative way to check for API response codes
		if strings.Contains(err.Error(), strconv.Itoa(http.StatusNotFound)) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// AddPolicyConditionsForContentItem adds policy conditions for all preview teams in a bundle
// when a content item is added
func (s *StateMachineBundleAPI) AddPolicyConditionsForContentItem(ctx context.Context, authToken string, bundle *models.Bundle, contentItem *models.ContentItem) error {
	if bundle.PreviewTeams == nil || len(*bundle.PreviewTeams) == 0 {
		return nil
	}

	for _, team := range *bundle.PreviewTeams {
		if err := s.addPolicyConditionForTeam(ctx, authToken, team.ID, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID); err != nil {
			return err
		}
	}

	return nil
}

// RemovePolicyConditionsForContentItem removes policy conditions for all preview teams in a bundle
// when a content item is removed
func (s *StateMachineBundleAPI) RemovePolicyConditionsForContentItem(ctx context.Context, authToken string, bundle *models.Bundle, contentItem *models.ContentItem) error {
	if bundle.PreviewTeams == nil || len(*bundle.PreviewTeams) == 0 {
		return nil
	}

	for _, team := range *bundle.PreviewTeams {
		if err := s.removePolicyConditionForTeam(ctx, authToken, team.ID, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID); err != nil {
			return err
		}
	}

	return nil
}

// RemovePolicyConditionsForRemovedPreviewTeams removes policy conditions for teams
// that have been removed from the bundle during a PUT bundle update.
//
//nolint:gocognit,gocyclo // cognitive complexity 45 (> 42) is acceptable for now
func (s *StateMachineBundleAPI) RemovePolicyConditionsForRemovedPreviewTeams(ctx context.Context, authToken, bundleID string, currentTeams, updatedTeams *[]models.PreviewTeam) error {
	removedTeams := findRemovedTeams(currentTeams, updatedTeams)
	if len(removedTeams) == 0 {
		return nil
	}

	contentItems, err := s.Datastore.GetContentItemsByBundleID(ctx, bundleID)
	if err != nil {
		return err
	}
	if len(contentItems) == 0 {
		return nil
	}

	for _, team := range removedTeams {
		otherBundles, err := s.Datastore.GetBundlesByPreviewTeamID(ctx, team.ID)
		if err != nil {
			return err
		}

		datasetsInUse := make(map[string]bool)
		for _, bundle := range otherBundles {
			if bundle.ID == bundleID {
				continue
			}

			otherContentItems, err := s.Datastore.GetContentItemsByBundleID(ctx, bundle.ID)
			if err != nil {
				return err
			}

			for _, otherItem := range otherContentItems {
				datasetsInUse[otherItem.Metadata.DatasetID] = true
			}
		}

		var toRemove []string
		for _, item := range contentItems {
			toRemove = append(toRemove, item.Metadata.DatasetID+"/"+item.Metadata.EditionID)

			if !datasetsInUse[item.Metadata.DatasetID] {
				toRemove = append(toRemove, item.Metadata.DatasetID)
			}
		}

		policy, err := s.PermissionsAPIClient.GetPolicy(ctx, team.ID, permissionsAPISDK.Headers{Authorization: authToken})
		if err != nil {
			return err
		}

		result := removeConditionValues(policy.Condition.Values, toRemove...)

		seen := make(map[string]bool)
		deduplicated := []string{}
		for _, v := range result {
			if !seen[v] {
				seen[v] = true
				deduplicated = append(deduplicated, v)
			}
		}
		result = deduplicated

		valuesChanged := len(result) != len(policy.Condition.Values)
		if !valuesChanged {
			for i, v := range result {
				if v != policy.Condition.Values[i] {
					valuesChanged = true
					break
				}
			}
		}

		if !valuesChanged {
			continue
		}

		if len(result) == 0 {
			policy.Condition.Values = nil
		} else {
			policy.Condition.Values = result
		}

		if err := s.PermissionsAPIClient.PutPolicy(ctx, team.ID, *policy, permissionsAPISDK.Headers{Authorization: authToken}); err != nil {
			return err
		}
	}

	return nil
}

// addPolicyConditionForTeam adds dataset/edition values to a single preview team's policy
func (s *StateMachineBundleAPI) addPolicyConditionForTeam(ctx context.Context, authToken, teamID, datasetID, editionID string) error {
	policy, err := s.PermissionsAPIClient.GetPolicy(ctx, teamID, permissionsAPISDK.Headers{Authorization: authToken})
	if err != nil {
		return err
	}

	if policy.Condition.Attribute == "" {
		policy.Condition = permissionsAPIModels.Condition{
			Attribute: conditionAttributeDatasetEdition,
			Operator:  conditionOperatorStringEquals,
			Values:    []string{datasetID, datasetID + "/" + editionID},
		}
	} else {
		policy.Condition.Values = append(policy.Condition.Values, datasetID, datasetID+"/"+editionID)
	}

	err = s.PermissionsAPIClient.PutPolicy(ctx, teamID, *policy, permissionsAPISDK.Headers{Authorization: authToken})
	if err != nil {
		return err
	}

	return nil
}

// removePolicyConditionForTeam removes dataset/edition values from a single preview team's policy
func (s *StateMachineBundleAPI) removePolicyConditionForTeam(ctx context.Context, authToken, teamID, datasetID, editionID string) error {
	policy, err := s.PermissionsAPIClient.GetPolicy(ctx, teamID, permissionsAPISDK.Headers{Authorization: authToken})
	if err != nil {
		return err
	}

	policy.Condition.Values = removeConditionValues(policy.Condition.Values, datasetID, datasetID+"/"+editionID)

	err = s.PermissionsAPIClient.PutPolicy(ctx, teamID, *policy, permissionsAPISDK.Headers{Authorization: authToken})
	if err != nil {
		return err
	}

	return nil
}

// findRemovedTeams returns teams present in current but absent from updated.
func findRemovedTeams(current, updated *[]models.PreviewTeam) []models.PreviewTeam {
	updatedSet := map[string]bool{}
	if updated != nil {
		for _, t := range *updated {
			updatedSet[t.ID] = true
		}
	}

	var removed []models.PreviewTeam
	if current != nil {
		for _, t := range *current {
			if !updatedSet[t.ID] {
				removed = append(removed, t)
			}
		}
	}
	return removed
}

// removeConditionValues removes specific values from the Condition array
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

// AddPolicyConditionsForAddedPreviewTeams adds policy conditions for teams
// that have been added to the bundle during a PUT bundle update
func (s *StateMachineBundleAPI) AddPolicyConditionsForAddedPreviewTeams(ctx context.Context, authToken, bundleID string, currentTeams, updatedTeams *[]models.PreviewTeam) error {
	addedTeams := findAddedTeams(currentTeams, updatedTeams)
	if len(addedTeams) == 0 {
		return nil
	}

	contentItems, err := s.Datastore.GetContentItemsByBundleID(ctx, bundleID)
	if err != nil {
		return err
	}
	if len(contentItems) == 0 {
		return nil
	}

	for _, team := range addedTeams {
		valuesToAdd := make(map[string]bool)
		for _, item := range contentItems {
			valuesToAdd[item.Metadata.DatasetID] = true
			valuesToAdd[item.Metadata.DatasetID+"/"+item.Metadata.EditionID] = true
		}

		policy, err := s.PermissionsAPIClient.GetPolicy(ctx, team.ID, permissionsAPISDK.Headers{Authorization: authToken})
		if err != nil {
			return err
		}

		existingValues := make(map[string]bool)
		for _, v := range policy.Condition.Values {
			existingValues[v] = true
		}

		for value := range valuesToAdd {
			if !existingValues[value] {
				policy.Condition.Values = append(policy.Condition.Values, value)
			}
		}

		err = s.PermissionsAPIClient.PutPolicy(ctx, team.ID, *policy, permissionsAPISDK.Headers{Authorization: authToken})
		if err != nil {
			return err
		}
	}

	return nil
}

// findAddedTeams returns teams present in updated but absent from current
func findAddedTeams(current, updated *[]models.PreviewTeam) []models.PreviewTeam {
	currentSet := map[string]bool{}
	if current != nil {
		for _, t := range *current {
			currentSet[t.ID] = true
		}
	}

	var added []models.PreviewTeam
	if updated != nil {
		for _, t := range *updated {
			if !currentSet[t.ID] {
				added = append(added, t)
			}
		}
	}
	return added
}
