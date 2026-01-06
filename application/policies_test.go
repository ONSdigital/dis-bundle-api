package application

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	getMethod = "GET"
	putMethod = "PUT"
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
					PostPolicyWithIDFunc: func(ctx context.Context, headers permissionsAPISDK.Headers, id string, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
						return nil, nil
					},
				}

				stateMachineBundleAPI := &StateMachineBundleAPI{
					PermissionsAPIClient: mockPermissionsAPIClient,
				}

				Convey("Then the expected error and number of calls are returned", func() {
					err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", tc.previewTeams, tc.role)
					So(err, ShouldEqual, tc.expectedErr)
					So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, tc.expectedCalls)
				})
			})
		}

		Convey("When PostPolicyWithID returns an error", func() {
			errExpectedFailure := errors.New("expected failure")
			mockPermissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
				PostPolicyWithIDFunc: func(ctx context.Context, headers permissionsAPISDK.Headers, id string, policy permissionsAPIModels.PolicyInfo) (*permissionsAPIModels.Policy, error) {
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
				err := stateMachineBundleAPI.CreateBundlePolicies(ctx, "auth-token", previewTeams, models.RoleDatasetsPreviewer)
				So(err, ShouldEqual, errExpectedFailure)
				So(len(mockPermissionsAPIClient.PostPolicyWithIDCalls()), ShouldEqual, 1)
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
			permissionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case getMethod:
					policy := permissionsAPIModels.Policy{
						ID:        "team-123",
						Condition: permissionsAPIModels.Condition{},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(policy)
				case putMethod:
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer permissionsServer.Close()

			permissionsAPIClient := permissionsAPISDK.NewClient(permissionsServer.URL)

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: permissionsAPIClient,
				PermissionsAPIURL:    permissionsServer.URL,
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
			})
		})

		Convey("When adding content to existing policy with values", func() {
			permissionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case getMethod:
					policy := permissionsAPIModels.Policy{
						ID: "team-123",
						Condition: permissionsAPIModels.Condition{
							Attribute: "dataset_edition",
							Operator:  "StringEquals",
							Values:    []string{"existing-dataset", "existing-dataset/existing-edition"},
						},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(policy)
				case putMethod:
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer permissionsServer.Close()

			permissionsAPIClient := permissionsAPISDK.NewClient(permissionsServer.URL)

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: permissionsAPIClient,
				PermissionsAPIURL:    permissionsServer.URL,
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
			})
		})

		Convey("When removing content item values", func() {
			permissionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case getMethod:
					policy := permissionsAPIModels.Policy{
						ID: "team-123",
						Condition: permissionsAPIModels.Condition{
							Attribute: "dataset_edition",
							Operator:  "StringEquals",
							Values:    []string{"dataset-1", "dataset-1/edition-1", "dataset-2", "dataset-2/edition-2"},
						},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(policy)
				case putMethod:
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer permissionsServer.Close()

			permissionsAPIClient := permissionsAPISDK.NewClient(permissionsServer.URL)

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: permissionsAPIClient,
				PermissionsAPIURL:    permissionsServer.URL,
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
			})
		})

		Convey("When GetPolicy fails", func() {
			permissionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer permissionsServer.Close()

			permissionsAPIClient := permissionsAPISDK.NewClient(permissionsServer.URL)

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: permissionsAPIClient,
				PermissionsAPIURL:    permissionsServer.URL,
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
			})
		})

		Convey("When updating multiple preview teams", func() {
			permissionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case getMethod:
					policy := permissionsAPIModels.Policy{
						ID:        r.URL.Path[len("/v1/policies/"):],
						Condition: permissionsAPIModels.Condition{},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(policy)
				case putMethod:
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer permissionsServer.Close()

			permissionsAPIClient := permissionsAPISDK.NewClient(permissionsServer.URL)

			stateMachineBundleAPI := &StateMachineBundleAPI{
				PermissionsAPIClient: permissionsAPIClient,
				PermissionsAPIURL:    permissionsServer.URL,
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
			})
		})
	})
}
