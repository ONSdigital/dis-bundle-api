Feature: Add a dataset item to a bundle - POST /bundles/{id}/contents

    Background:
        Given I have these bundles:
        """
        [
            {
                "id": "bundle-1",
                "bundle_type": "SCHEDULED",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "created_at": "2025-06-09T07:00:00Z",
                "last_updated_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "scheduled_at": "2025-05-05T08:00:00Z",
                "state": "DRAFT",
                "title": "bundle-1",
                "updated_at": "2025-06-10T07:00:00Z",
                "managed_by": "WAGTAIL",
                "e_tag": "original-etag"
            }
        ]
        """
        And I have these content items:
        """
        [
            {
                "id": "content-item-2",
                "bundle_id": "bundle-1",
                "content_type": "DATASET",
                "metadata": {
                    "dataset_id": "dataset2",
                    "edition_id": "edition2",
                    "version_id": 2,
                    "title": "Test Dataset 2"
                },
                "links": {
                    "edit": "edit/link",
                    "preview": "preview/link"
                }
            }
        ]
        """
    
    Scenario: POST /bundles/{id}/contents successfully
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "bundle_id": "bundle-1",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            """
        Then the HTTP status code should be "201"
        And the response header "Content-Type" should be "application/json"
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        And the response header "Location" should contain "/bundles/bundle-1/contents/"
        Then I should receive the following ContentItem JSON response:
            """
            {
                "bundle_id": "bundle-1",
                "content_type": "DATASET",
                "metadata": {
                    "dataset_id": "dataset1",
                    "edition_id": "edition1",
                    "version_id": 1,
                    "title": "Test Dataset"
                },
                "id": "new-uuid",
                "links": {
                    "edit": "edit/link",
                    "preview": "preview/link"
                }
            }
            """
    
    Scenario: POST /bundles/{id}/contents with an invalid body (missing content_type)
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "bundle_id": "bundle-1",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            """
        Then I should receive the following JSON response with status "400":
            """
            {
                "errors": [
                    {
                        "code": "missing_parameters",
                        "description": "Unable to process request due to a malformed or invalid request body or query parameter",
                        "source": {
                            "field": "/content_type"
                        }
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with a non-existent bundle
        Given I am an admin user
        When I POST "/bundles/bundle-missing/contents"
            """
                {
                    "bundle_id": "bundle-missing",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            """
        Then I should receive the following JSON response with status "404":
            """
            {
                "errors": [
                    {
                        "code": "not_found",
                        "description": "Bundle not found"
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with a dataset that doesn't exist
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "bundle_id": "bundle-1",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "fail-get-version",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            """
        Then I should receive the following JSON response with status "404":
            """
            {
                "errors": [
                    {
                        "code": "not_found",
                        "description": "Dataset version not found"
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with an existing content item with the same dataset
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "bundle_id": "bundle-1",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset2",
                        "edition_id": "edition2",
                        "version_id": 2,
                        "title": "Test Dataset 2"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            """
        Then I should receive the following JSON response with status "409":
            """
            {
                "errors": [
                    {
                        "code": "conflict",
                        "description": "Content item already exists for the given dataset, edition and version"
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with no authentication
        Given I am not authenticated
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "bundle_id": "bundle-1",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            """
        Then the HTTP status code should be "401"
        And the response body should be empty