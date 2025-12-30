package application

import (
	"context"

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
