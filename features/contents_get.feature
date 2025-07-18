Feature: Get all content items in a bundle - GET /bundles/{id}/contents

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
                    "state": "PUBLISHED",
                    "title": "bundle-1",
                    "updated_at": "2025-06-10T07:00:00Z",
                    "managed_by": "WAGTAIL",
                    "e_tag": "original-etag"
                },
                {
                    "id": "bundle-2",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-06-09T08:00:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "890m231k-98df-11ec-b909-0242ac120003"
                        }
                    ],
                    "scheduled_at": "2025-05-05T09:00:00Z",
                    "state": "DRAFT",
                    "title": "bundle-2",
                    "updated_at": "2025-06-10T08:00:00Z",
                    "managed_by": "WAGTAIL",
                    "e_tag": "original-etag"
                },
                {
                    "id": "bundle-3",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-06-09T09:00:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "890m231k-98df-11ec-b909-0242ac120009"
                        }
                    ],
                    "scheduled_at": "2025-05-05T06:00:00Z",
                    "state": "DRAFT",
                    "title": "bundle-3",
                    "updated_at": "2025-06-10T06:00:00Z",
                    "managed_by": "WAGTAIL",
                    "e_tag": "original-etag"
                }
            ]
            """
        And I have these content items:
            """
            [
                {
                    "id": "content-item-1",
                    "bundle_id": "bundle-1",
                    "content_type": "DATASET",
                    "state": "APPROVED",
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
                },
                {
                    "id": "content-item-1-secondary",
                    "bundle_id": "bundle-1",
                    "content_type": "DATASET",
                    "state": "APPROVED",
                    "metadata": {
                        "dataset_id": "dataset2",
                        "edition_id": "edition2",
                        "version_id": 1,
                        "title": "Test Dataset2"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-2",
                    "bundle_id": "bundle-2",
                    "content_type": "DATASET",
                    "state": "APPROVED",
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
                },
                {
                    "id": "content-item-3",
                    "bundle_id": "bundle-3",
                    "content_type": "DATASET",
                    "state": "APPROVED",
                    "metadata": {
                        "dataset_id": "dataset-id-does-not-exist",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                }
            ]
            """

    Scenario: GET /bundles/{id}/contents successfully
        Given I am an admin user
        When I GET "/bundles/bundle-1/contents"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        Then I should receive the following JSON response:
            """
            {
                "items": [
                    {
                        "id": "content-item-1-secondary",
                        "bundle_id": "bundle-1",
                        "content_type": "DATASET",
                        "state": "APPROVED",
                        "metadata": {
                            "dataset_id": "dataset2",
                            "edition_id": "edition2",
                            "version_id": 1,
                            "title": "Test Dataset2"
                        },
                        "links": {
                            "edit": "edit/link",
                            "preview": "preview/link"
                        }
                    },
                    {
                        "id": "content-item-1",
                        "bundle_id": "bundle-1",
                        "content_type": "DATASET",
                        "state": "APPROVED",
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
                ],
                "count": 2,
                "offset": 0,
                "limit": 20,
                "total_count": 2
            }
            """

    Scenario: GET /bundles/{id}/contents successfully when bundle state is not published
        Given I am an admin user
        When I GET "/bundles/bundle-2/contents"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        Then I should receive the following JSON response:
            """
            {
                "items": [
                    {
                        "id": "content-item-2",
                        "bundle_id": "bundle-2",
                        "content_type": "DATASET",
                        "state": "edition-confirmed",
                        "metadata": {
                            "dataset_id": "dataset1",
                            "edition_id": "edition1",
                            "version_id": 1,
                            "title": "Test Dataset title"
                        },
                        "links": {
                            "edit": "edit/link",
                            "preview": "preview/link"
                        }
                    }
                ],
                "count": 1,
                "offset": 0,
                "limit": 20,
                "total_count": 1
            }
            """

    Scenario: GET /bundles/{id}/contents with non-existent bundle
        Given I am an admin user
        When I GET "/bundles/non-existent-bundle/contents"
        Then I should receive the following JSON response with status "404":
            """
            {
                "errors": [
                    {
                        "code": "NotFound",
                        "description": "The requested resource does not exist."
                    }
                ]
            }
            """

    Scenario: GET /bundles/{id}/contents with invalid pagination parameter
        Given I am an admin user
        When I GET "/bundles/bundle-1/contents?offset=-1"
        Then I should receive the following JSON response with status "400":
            """
            {
                "errors": [
                    {
                        "code": "BadRequest",
                        "description": "Unable to process request due to a malformed or invalid request body or query parameter.",
                        "source": {
                            "parameter": " offset"
                        }
                    }
                ]
            }
            """

    Scenario: GET /bundles/{id}/contents with dataset that does not exist
        Given I am an admin user
        When I GET "/bundles/bundle-3/contents"
        Then the HTTP status code should be "404"
        Then I should receive the following JSON response:
            """
            {
                "errors": [
                    {
                        "code": "NotFound",
                        "description": "The requested resource does not exist.",
                        "source": {
                            "field": "/metadata/dataset_id"
                        }
                    }
                ]
            }
            """

    Scenario: GET /bundles/{id}/contents with no authentication
        Given I am not authenticated
        When I GET "/bundles/bundle-1/contents"
        Then the HTTP status code should be "401"
        And the response body should be empty
