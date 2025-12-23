package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateBundlePolicies(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()

		type testCase struct {
			name          string
			previewTeams  *[]models.PreviewTeam
			role          models.Role
			expectedErr   error
			expectedCalls int
		}

		cases := []testCase{
			{
				name: "single preview team",
				previewTeams: &[]models.PreviewTeam{
					{ID: "team-1"},
				},
				role:          models.RoleDatasetsPreviewer,
				expectedErr:   nil,
				expectedCalls: 1,
			},
			{
				name: "multiple preview teams",
				previewTeams: &[]models.PreviewTeam{
					{ID: "team-1"},
					{ID: "team-2"},
				},
				role:          models.RoleDatasetsPreviewer,
				expectedErr:   nil,
				expectedCalls: 2,
			},
			{
				name: "invalid role",
				previewTeams: &[]models.PreviewTeam{
					{ID: "team-1"},
				},
				role:          models.Role("invalid-role"),
				expectedErr:   apierrors.ErrInvalidRole,
				expectedCalls: 0,
			},
			{
				name:          "nil preview teams",
				previewTeams:  nil,
				role:          models.RoleDatasetsPreviewer,
				expectedErr:   nil,
				expectedCalls: 0,
			},
			{
				name:          "empty preview teams slice",
				previewTeams:  &[]models.PreviewTeam{},
				role:          models.RoleDatasetsPreviewer,
				expectedErr:   nil,
				expectedCalls: 0,
			},
		}

		for _, tc := range cases {
			Convey("When CreateBundlePolicies is called with: "+tc.name, func() {
				mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
					PostPolicyFunc: func(ctx context.Context, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
						return nil, nil
					},
				}

				stateMachineBundleAPI := &StateMachineBundleAPI{
					PermissionsAPIClient: mockPermissionsAPIClient,
				}

				Convey("Then the expected error and number of calls are returned", func() {
					err := stateMachineBundleAPI.CreateBundlePolicies(ctx, tc.previewTeams, tc.role)
					So(err, ShouldEqual, tc.expectedErr)
					So(len(mockPermissionsAPIClient.PostPolicyCalls()), ShouldEqual, tc.expectedCalls)
				})
			})
		}

		Convey("When PostPolicy returns an error", func() {
			errExpectedFailure := errors.New("expected failure")
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				PostPolicyFunc: func(ctx context.Context, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
					return nil, errExpectedFailure
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			previewTeams := &[]models.PreviewTeam{
				{ID: "team-1"},
			}

			Convey("Then the error is returned", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldEqual, errExpectedFailure)
				So(len(mockPermissionsAPIClient.PostPolicyCalls()), ShouldEqual, 1)
			})
		})
	})
}
