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
                "created_at": "2025-01-01T07:00:00Z",
                "last_updated_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "scheduled_at": "2025-01-03T07:00:00Z",
                "state": "DRAFT",
                "title": "bundle-1",
                "updated_at": "2025-01-02T07:00:00Z",
                "managed_by": "WAGTAIL",
                "e_tag": "original-etag"
            },
            {
                "id": "bundle-2",
                "bundle_type": "MANUAL",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "created_at": "2025-01-01T07:00:00Z",
                "last_updated_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "title": "bundle-2",
                "updated_at": "2025-01-02T07:00:00Z",
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
    And I have these dataset versions:
        """
        [
            {
                "id": "version-1",
                "version": 1,
                "dataset_id": "dataset1",
                "edition": "edition1",
                "state": "approved" 
            },
            {
                "id": "version-2",
                "version": 2,
                "dataset_id": "dataset2",
                "edition": "edition2",
                "state": "approved" 
            }
        ]
        """
        
    Scenario: POST /bundles/{id}/contents successfully for SCHEDULED bundle
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
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
        And the total number of events should be 2
        And the number of events with action "CREATE" and datatype "content_item" should be 1
        And the number of events with action "UPDATE" and datatype "bundle" should be 1
        And the release date for the dataset version with id "version-1" should be "2025-01-03T07:00:00.000Z"

    Scenario: POST /bundles/{id}/contents successfully for MANUAL bundle
        Given I am an admin user
        When I POST "/bundles/bundle-2/contents"
            """
                {
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
        And the response header "Location" should contain "/bundles/bundle-2/contents/"
        Then I should receive the following ContentItem JSON response:
            """
            {
                "bundle_id": "bundle-2",
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
        And the total number of events should be 2
        And the number of events with action "CREATE" and datatype "content_item" should be 1
        And the number of events with action "UPDATE" and datatype "bundle" should be 1
        And the release date for the dataset version with id "version-1" should be ""
    
    Scenario: POST /bundles/{id}/contents with an invalid body (invalid content_type and missing edit link)
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "content_type": "INVALID",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition1",
                        "version_id": 1,
                        "title": "Test Dataset"
                    },
                    "links": {
                        "preview": "preview/link"
                    }
                }
            """
        Then I should receive the following JSON response with status "400":
            """
            {
                "errors": [
                    {
                        "code": "InvalidParameters",
                        "description": "Unable to process request due to a malformed or invalid request body or query parameter.",
                        "source": {
                            "field": "/content_type"
                        }
                    },
                    {
                        "code": "MissingParameters",
                        "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                        "source": {
                            "field": "/links/edit"
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
                        "code": "NotFound",
                        "description": "The requested resource does not exist."
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with a dataset that doesn't exist
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset-not-found",
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
                        "code": "NotFound",
                        "description": "The requested resource does not exist.",
                        "source": {
                            "field": "/metadata/dataset_id"
                        }
                    }
                ]
            }
            """

    Scenario: POST /bundles/{id}/contents with an edition that doesn't exist
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition-not-found",
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
                        "code": "NotFound",
                        "description": "The requested resource does not exist.",
                        "source": {
                            "field": "/metadata/edition_id"
                        }
                    }
                ]
            }
            """

    Scenario: POST /bundles/{id}/contents with a version that doesn't exist
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset1",
                        "edition_id": "edition1",
                        "version_id": 404,
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
                        "code": "NotFound",
                        "description": "The requested resource does not exist.",
                        "source": {
                            "field": "/metadata/version_id"
                        }
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with an existing content item with the same dataset
        Given I am an admin user
        When I POST "/bundles/bundle-1/contents"
            """
                {
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
                        "code": "Conflict",
                        "description": "Change rejected due to a conflict with the current resource state. A common cause is attempting to change a bundle that is already locked pending publication or has already been published."
                    }
                ]
            }
            """
    
    Scenario: POST /bundles/{id}/contents with no authentication
        Given I am not authenticated
        When I POST "/bundles/bundle-1/contents"
            """
                {
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