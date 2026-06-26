package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	storetest "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
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
					So(policyInfo.Condition.Attribute, ShouldEqual, conditionAttributeDatasetEdition)
					So(policyInfo.Condition.Operator.String(), ShouldEqual, conditionOperatorStringEquals)
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
					So(policyInfo.Condition.Attribute, ShouldEqual, conditionAttributeDatasetEdition)
					So(policyInfo.Condition.Operator.String(), ShouldEqual, conditionOperatorStringEquals)
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

func TestAddPolicyConditionsForContentItem(t *testing.T) {
	Convey("Given a StateMachineBundleAPI", t, func() {
		ctx := context.Background()

		Convey("When adding policies for a bundle with no preview teams", func() {
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

			err := stateMachineBundleAPI.AddPolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

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

			err := stateMachineBundleAPI.AddPolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

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

			err := stateMachineBundleAPI.AddPolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

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

			err := stateMachineBundleAPI.AddPolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 0)
			})
		})

		Convey("When adding for multiple preview teams", func() {
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

			err := stateMachineBundleAPI.AddPolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 2)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 2)
			})
		})
	})
}

func TestRemovePolicyConditionsForContentItem(t *testing.T) {
	Convey("Given a StateMachineBundleAPI", t, func() {
		ctx := context.Background()

		Convey("When removing policies for a bundle with no preview teams", func() {
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

			err := stateMachineBundleAPI.RemovePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

			Convey("Then no error is returned and no API calls are made", func() {
				So(err, ShouldBeNil)
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

			err := stateMachineBundleAPI.RemovePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

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

			err := stateMachineBundleAPI.RemovePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 0)
			})
		})

		Convey("When removing for multiple preview teams", func() {
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return &permissionsAPIModels.Policy{
						ID: id,
						Condition: permissionsAPIModels.Condition{
							Attribute: "dataset_edition",
							Operator:  "StringEquals",
							Values:    []string{"dataset-1", "dataset-1/edition-1"},
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

			err := stateMachineBundleAPI.RemovePolicyConditionsForContentItem(ctx, "auth-token", bundle, contentItem)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 2)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 2)
			})
		})
	})
}

func TestAddPolicyConditionsForAddedPreviewTeams(t *testing.T) {
	Convey("Given a StateMachineBundleAPI", t, func() {
		ctx := context.Background()

		Convey("When no teams are added", func() {
			mockDatastore := &storetest.StorerMock{}
			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore: store.Datastore{Backend: mockDatastore},
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then no error is returned and no API calls are made", func() {
				So(err, ShouldBeNil)
				So(len(mockDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a team is added but bundle has no content items", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{}, nil
				},
			}
			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore: store.Datastore{Backend: mockDatastore},
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			err := stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then no error is returned and GetPolicy is not called", func() {
				So(err, ShouldBeNil)
				So(len(mockDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a team is added with content items", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							ID:       "content-1",
							BundleID: "bundle-1",
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
								VersionID: 1,
							},
						},
					}, nil
				},
			}

			existingPolicy := &permissionsAPIModels.Policy{
				Condition: permissionsAPIModels.Condition{
					Attribute: "dataset_edition",
					Operator:  "StringEquals",
					Values:    []string{},
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return existingPolicy, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					So(policy.Condition.Values, ShouldContain, "dataset-1")
					So(policy.Condition.Values, ShouldContain, "dataset-1/edition-1")
					So(len(policy.Condition.Values), ShouldEqual, 2)
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			err := stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then values are added to the policy", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a team is added and policy already has some values", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
							},
						},
					}, nil
				},
			}

			existingPolicy := &permissionsAPIModels.Policy{
				Condition: permissionsAPIModels.Condition{
					Attribute: "dataset_edition",
					Operator:  "StringEquals",
					Values:    []string{"dataset-2", "dataset-2/edition-2"},
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return existingPolicy, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					So(policy.Condition.Values, ShouldContain, "dataset-1")
					So(policy.Condition.Values, ShouldContain, "dataset-1/edition-1")
					So(policy.Condition.Values, ShouldContain, "dataset-2")
					So(policy.Condition.Values, ShouldContain, "dataset-2/edition-2")
					So(len(policy.Condition.Values), ShouldEqual, 4)
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			err := stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then new values are added without duplicates", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a team is added and values already exist (no duplicates)", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
							},
						},
					}, nil
				},
			}

			existingPolicy := &permissionsAPIModels.Policy{
				Condition: permissionsAPIModels.Condition{
					Attribute: "dataset_edition",
					Operator:  "StringEquals",
					Values:    []string{"dataset-1", "dataset-1/edition-1"},
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return existingPolicy, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					So(len(policy.Condition.Values), ShouldEqual, 2)
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			err := stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then no duplicate values are added", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 0)
			})
		})

		Convey("When GetPolicy fails", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
							},
						},
					}, nil
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("get policy error")
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			err := stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "get policy error")
			})
		})
	})
}

func TestRemovePolicyConditionsForRemovedPreviewTeams(t *testing.T) {
	Convey("Given a StateMachineBundleAPI", t, func() {
		ctx := context.Background()

		Convey("When no teams are removed", func() {
			mockDatastore := &storetest.StorerMock{}
			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore: store.Datastore{Backend: mockDatastore},
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then no error is returned and no API calls are made", func() {
				So(err, ShouldBeNil)
				So(len(mockDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a team is removed but bundle has no content items", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{}, nil
				},
			}
			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore: store.Datastore{Backend: mockDatastore},
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then no error is returned and GetPolicy is not called", func() {
				So(err, ShouldBeNil)
				So(len(mockDatastore.GetContentItemsByBundleIDCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a team is removed and dataset not used elsewhere", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
							},
						},
					}, nil
				},
				GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
					return []*models.Bundle{}, nil
				},
			}

			existingPolicy := &permissionsAPIModels.Policy{
				Condition: permissionsAPIModels.Condition{
					Attribute: "dataset_edition",
					Operator:  "StringEquals",
					Values:    []string{"dataset-1", "dataset-1/edition-1"},
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return existingPolicy, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					So(policy.Condition.Values, ShouldBeNil)
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then both dataset and edition values are removed and values becomes nil", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a team is removed but dataset used in another bundle", func() {
			bundle2 := &models.Bundle{
				ID:    "bundle-2",
				Title: "Bundle 2",
			}

			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					if bundleID == "bundle-1" {
						return []*models.ContentItem{
							{
								Metadata: models.Metadata{
									DatasetID: "dataset-1",
									EditionID: "edition-1",
								},
							},
						}, nil
					}
					if bundleID == "bundle-2" {
						return []*models.ContentItem{
							{
								Metadata: models.Metadata{
									DatasetID: "dataset-1",
									EditionID: "edition-2",
								},
							},
						}, nil
					}
					return []*models.ContentItem{}, nil
				},
				GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
					return []*models.Bundle{bundle2}, nil
				},
			}

			existingPolicy := &permissionsAPIModels.Policy{
				Condition: permissionsAPIModels.Condition{
					Attribute: "dataset_edition",
					Operator:  "StringEquals",
					Values:    []string{"dataset-1", "dataset-1/edition-1", "dataset-1/edition-2"},
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return existingPolicy, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					So(policy.Condition.Values, ShouldContain, "dataset-1")
					So(policy.Condition.Values, ShouldContain, "dataset-1/edition-2")
					So(policy.Condition.Values, ShouldNotContain, "dataset-1/edition-1")
					So(len(policy.Condition.Values), ShouldEqual, 2)
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then only edition is removed, dataset value remains", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a team is removed but policy has no matching values", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
							},
						},
					}, nil
				},
				GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
					return []*models.Bundle{}, nil
				},
			}

			existingPolicy := &permissionsAPIModels.Policy{
				Condition: permissionsAPIModels.Condition{
					Attribute: "dataset_edition",
					Operator:  "StringEquals",
					Values:    []string{"different-dataset", "different-dataset/edition"},
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return existingPolicy, nil
				},
				PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
					return nil
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then PutPolicy is not called since values didn't change", func() {
				So(err, ShouldBeNil)
				So(len(mockPermissionsAPIClient.GetPolicyCalls()), ShouldEqual, 1)
				So(len(mockPermissionsAPIClient.PutPolicyCalls()), ShouldEqual, 0)
			})
		})

		Convey("When GetPolicy fails", func() {
			mockDatastore := &storetest.StorerMock{
				GetContentItemsByBundleIDFunc: func(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
					return []*models.ContentItem{
						{
							Metadata: models.Metadata{
								DatasetID: "dataset-1",
								EditionID: "edition-1",
							},
						},
					}, nil
				},
				GetBundlesByPreviewTeamIDFunc: func(ctx context.Context, teamID string) ([]*models.Bundle, error) {
					return []*models.Bundle{}, nil
				},
			}

			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
					return nil, errors.New("get policy error")
				},
			}

			stateMachineBundleAPI := &StateMachineBundleAPI{
				Datastore:            store.Datastore{Backend: mockDatastore},
				PermissionsAPIClient: mockPermissionsAPIClient,
			}

			currentTeams := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updatedTeams := []models.PreviewTeam{{ID: "team-1"}}

			err := stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, "auth-token", "bundle-1", &currentTeams, &updatedTeams)

			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "get policy error")
			})
		})
	})
}

func TestFindRemovedTeams(t *testing.T) {
	Convey("When finding removed teams", t, func() {
		Convey("With no teams removed", func() {
			current := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updated := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			removed := findRemovedTeams(&current, &updated)

			Convey("Then no teams are returned", func() {
				So(len(removed), ShouldEqual, 0)
			})
		})

		Convey("With one team removed", func() {
			current := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updated := []models.PreviewTeam{{ID: "team-1"}}

			removed := findRemovedTeams(&current, &updated)

			Convey("Then one team is returned", func() {
				So(len(removed), ShouldEqual, 1)
				So(removed[0].ID, ShouldEqual, "team-2")
			})
		})

		Convey("With all teams removed", func() {
			current := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updated := []models.PreviewTeam{}

			removed := findRemovedTeams(&current, &updated)

			Convey("Then all teams are returned", func() {
				So(len(removed), ShouldEqual, 2)
			})
		})

		Convey("With nil current teams", func() {
			updated := []models.PreviewTeam{{ID: "team-1"}}

			removed := findRemovedTeams(nil, &updated)

			Convey("Then no teams are returned", func() {
				So(len(removed), ShouldEqual, 0)
			})
		})

		Convey("With nil updated teams", func() {
			current := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			removed := findRemovedTeams(&current, nil)

			Convey("Then all current teams are returned", func() {
				So(len(removed), ShouldEqual, 2)
			})
		})
	})
}

func TestFindAddedTeams(t *testing.T) {
	Convey("When finding added teams", t, func() {
		Convey("With no teams added", func() {
			current := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}
			updated := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			added := findAddedTeams(&current, &updated)

			Convey("Then no teams are returned", func() {
				So(len(added), ShouldEqual, 0)
			})
		})

		Convey("With one team added", func() {
			current := []models.PreviewTeam{{ID: "team-1"}}
			updated := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			added := findAddedTeams(&current, &updated)

			Convey("Then one team is returned", func() {
				So(len(added), ShouldEqual, 1)
				So(added[0].ID, ShouldEqual, "team-2")
			})
		})

		Convey("With all teams added", func() {
			current := []models.PreviewTeam{}
			updated := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			added := findAddedTeams(&current, &updated)

			Convey("Then all teams are returned", func() {
				So(len(added), ShouldEqual, 2)
			})
		})

		Convey("With nil current teams", func() {
			updated := []models.PreviewTeam{{ID: "team-1"}, {ID: "team-2"}}

			added := findAddedTeams(nil, &updated)

			Convey("Then all updated teams are returned", func() {
				So(len(added), ShouldEqual, 2)
			})
		})

		Convey("With nil updated teams", func() {
			current := []models.PreviewTeam{{ID: "team-1"}}

			added := findAddedTeams(&current, nil)

			Convey("Then no teams are returned", func() {
				So(len(added), ShouldEqual, 0)
			})
		})
	})
}
