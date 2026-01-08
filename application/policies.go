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
