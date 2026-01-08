package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateBundlePolicies(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		stateMachineBundleAPI := &StateMachineBundleAPI{}

		Convey("When CreateBundlePolicies is called with existing policies", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{}, nil
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			previewTeams := []models.PreviewTeam{
				{ID: "team-1"},
				{ID: "team-2"},
			}

			Convey("Then CreateBundlePolicies does not attempt to create new policies", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 0)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 2)
			})
		})

		Convey("When CreateBundlePolicies is called with non-existing policies", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("404 Not Found")
				},
				PostPolicyWithIDFunc: func(ctx context.Context, id string, policyInfo permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{}, nil
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			previewTeams := []models.PreviewTeam{
				{ID: "team-1"},
				{ID: "team-2"},
			}

			Convey("Then CreateBundlePolicies creates new policies for each preview team", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 2)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 2)
			})
		})

		Convey("When CreateBundlePolicies is called with some existing and some non-existing policies", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					if id == "team-1" {
						return &permissionsAPIModels.Policy{}, nil
					}
					return nil, errors.New("404 Not Found")
				},
				PostPolicyWithIDFunc: func(ctx context.Context, id string, policyInfo permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{}, nil
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			previewTeams := []models.PreviewTeam{
				{ID: "team-1"},
				{ID: "team-2"},
			}

			Convey("Then CreateBundlePolicies creates new policies only for non-existing preview teams", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 2)
			})
		})

		Convey("When CreateBundlePolicies is called and CheckPolicyExists returns an unexpected error", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("unexpected error")
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			previewTeams := []models.PreviewTeam{
				{ID: "team-1"},
			}

			Convey("Then CreateBundlePolicies returns the error", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unexpected error")
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 0)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When CreateBundlePolicies is called and PostPolicyWithID returns an error", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("404 Not Found")
				},
				PostPolicyWithIDFunc: func(ctx context.Context, id string, policyInfo permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("post error")
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			previewTeams := []models.PreviewTeam{
				{ID: "team-1"},
			}

			Convey("Then CreateBundlePolicies returns the error", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "post error")
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When CreateBundlePolicies is called with the following invalid inputs then the correct errors are returned", func() {
			Convey("nil preview teams", func() {
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", nil, models.RoleDatasetsPreviewer)
				So(err, ShouldBeNil)
			})

			Convey("empty preview teams", func() {
				emptyPreviewTeams := []models.PreviewTeam{}
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &emptyPreviewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldBeNil)
			})

			Convey("invalid role", func() {
				previewTeams := []models.PreviewTeam{
					{ID: "team-1"},
				}
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", &previewTeams, models.Role("invalid-role"))
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, apierrors.ErrInvalidRole)
			})
		})
	})
}

func TestCheckPolicyExists(t *testing.T) {
	Convey("Given a StateMachineBundleAPI with mocked dependencies", t, func() {
		ctx := context.Background()
		stateMachineBundleAPI := &StateMachineBundleAPI{}

		Convey("When the policy exists", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{}, nil
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			Convey("Then CheckPolicyExists returns true and no error", func() {
				exists, err := stateMachineBundleAPI.CheckPolicyExists(ctx, "auth-token", "policy-id")
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
			})
		})

		Convey("When the policy does not exist", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("404 Not Found")
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			Convey("Then CheckPolicyExists returns false and no error", func() {
				exists, err := stateMachineBundleAPI.CheckPolicyExists(ctx, "auth-token", "policy-id")
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})

		Convey("When an unexpected error occurs", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("unexpected error")
				},
			}
			stateMachineBundleAPI.PermissionsAPIClient = mockPermissionsAPIClient

			Convey("Then CheckPolicyExists returns the error", func() {
				exists, err := stateMachineBundleAPI.CheckPolicyExists(ctx, "auth-token", "policy-id")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unexpected error")
				So(exists, ShouldBeFalse)
			})
		})
	})
}
