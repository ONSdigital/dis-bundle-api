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

func TestUpdatePolicyConditionsForContentItem(t *testing.T) {
	Convey("Given a StateMachineBundleAPI", t, func() {
		ctx := context.Background()

		Convey("When updating policies for a bundle with no preview teams", func() {
			stateMachineBundleAPI := &StateMachineBundleAPI{}

			bundle := &models.Bundle{
				PreviewTeams: nil,
			}

			contentItem := &models.ContentItem{
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
				},
			}

			err := stateMachineBundleAPI.UpdatePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem, true)

			Convey("Then no error is returned and no API calls are made", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When adding first content item (empty condition)", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{
						ID:        "team-123",
						Condition: permissionsAPIModels.Condition{},
					}, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			previewTeam := models.PreviewTeam{ID: "team-123"}
			previewTeams := []models.PreviewTeam{previewTeam}

			bundle := &models.Bundle{
				PreviewTeams: &previewTeams,
			}

			contentItem := &models.ContentItem{
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
				},
			}

			err := stateMachineBundleAPI.UpdatePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem, true)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When adding content to existing policy with values", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{
						ID: "team-123",
						Condition: permissionsAPIModels.Condition{
							Attribute: "dataset_edition",
							Operator:  "StringEquals",
							Values:    []string{"existing-dataset", "existing-dataset/existing-edition"},
						},
					}, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			previewTeam := models.PreviewTeam{ID: "team-123"}
			previewTeams := []models.PreviewTeam{previewTeam}

			bundle := &models.Bundle{
				PreviewTeams: &previewTeams,
			}

			contentItem := &models.ContentItem{
				Metadata: models.Metadata{
					DatasetID: "new-dataset",
					EditionID: "new-edition",
				},
			}

			err := stateMachineBundleAPI.UpdatePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem, true)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When removing content item values", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{
						ID: "team-123",
						Condition: permissionsAPIModels.Condition{
							Attribute: "dataset_edition",
							Operator:  "StringEquals",
							Values:    []string{"dataset-1", "dataset-1/edition-1", "dataset-2", "dataset-2/edition-2"},
						},
					}, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			previewTeam := models.PreviewTeam{ID: "team-123"}
			previewTeams := []models.PreviewTeam{previewTeam}

			bundle := &models.Bundle{
				PreviewTeams: &previewTeams,
			}

			contentItem := &models.ContentItem{
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
				},
			}

			err := stateMachineBundleAPI.UpdatePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem, false)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When GetPolicy fails", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("internal server error")
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			previewTeam := models.PreviewTeam{ID: "team-123"}
			previewTeams := []models.PreviewTeam{previewTeam}

			bundle := &models.Bundle{
				PreviewTeams: &previewTeams,
			}

			contentItem := &models.ContentItem{
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
				},
			}

			err := stateMachineBundleAPI.UpdatePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem, true)

			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 0)
			})
		})

		Convey("When updating multiple preview teams", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{
						ID:        id,
						Condition: permissionsAPIModels.Condition{},
					}, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			previewTeams := []models.PreviewTeam{
				{ID: "team-alpha"},
				{ID: "team-beta"},
			}

			bundle := &models.Bundle{
				PreviewTeams: &previewTeams,
			}

			contentItem := &models.ContentItem{
				Metadata: models.Metadata{
					DatasetID: "dataset-1",
					EditionID: "edition-1",
				},
			}

			err := stateMachineBundleAPI.UpdatePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem, true)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 2)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 2)
			})
		})
	})
}
