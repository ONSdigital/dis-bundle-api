Feature: Update Bundles functionality - PUT /bundles/{id}/state

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
                    "created_at": "2025-04-03T11:25:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "890m231k-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "scheduled_at": "2025-05-05T08:00:00Z",
                    "state": "IN_REVIEW",
                    "title": "bundle-1",
                    "updated_at": "2025-04-03T11:25:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "bundle-2",
                    "bundle_type": "MANUAL",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-04T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "DRAFT",
                    "title": "bundle-2",
                    "updated_at": "2025-04-04T13:40:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "bundle-3",
                    "bundle_type": "MANUAL",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "IN_REVIEW",
                    "title": "bundle-3",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "bundle-4",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "APPROVED",
                    "title": "bundle-4",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "bundle-5",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "APPROVED",
                    "title": "bundle-5",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "bundle-6",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "APPROVED",
                    "title": "bundle-6",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "bundle-10",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "IN_REVIEW",
                    "title": "bundle-10",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "WAGTAIL"
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
                },
                {
                    "id": "content-item-2",
                    "bundle_id": "bundle-2",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset3",
                        "edition_id": "edition3",
                        "version_id": 3,
                        "title": "Test Dataset 3"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-3",
                    "bundle_id": "bundle-4",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset4",
                        "edition_id": "edition4",
                        "version_id": 1,
                        "title": "Test Dataset 4"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    },
                    "state": "APPROVED"
                },
                {
                    "id": "content-item-12",
                    "bundle_id": "bundle-3",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset5",
                        "edition_id": "edition5",
                        "version_id": 1,
                        "title": "Test Dataset 5"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-13",
                    "bundle_id": "bundle-3",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset6",
                        "edition_id": "edition6",
                        "version_id": 1,
                        "title": "Test Dataset 6"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-14",
                    "bundle_id": "bundle-3",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset7",
                        "edition_id": "edition7",
                        "version_id": 1,
                        "title": "Test Dataset 7"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-15",
                    "bundle_id": "bundle-3",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset8",
                        "edition_id": "edition8",
                        "version_id": 1,
                        "title": "Test Dataset 8"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-16",
                    "bundle_id": "bundle-3",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset9",
                        "edition_id": "edition9",
                        "version_id": 10,
                        "title": "Test Dataset 9"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-17",
                    "bundle_id": "bundle-10",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset10",
                        "edition_id": "edition10",
                        "version_id": 20,
                        "title": "Test Dataset 10"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    }
                },
                {
                    "id": "content-item-18",
                    "bundle_id": "bundle-5",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "dataset10",
                        "edition_id": "edition10",
                        "version_id": 7,
                        "title": "Test Dataset 11"
                    },
                    "links": {
                        "edit": "edit/link",
                        "preview": "preview/link"
                    },
                    "state": "APPROVED"
                }
            ]
            """
        And I have these dataset versions:
            """
            [
                {
                    "id": "version-1",
                    "version": 1,
                    "dataset_id": "dataset4",
                    "edition": "edition4",
                    "state": "approved" 
                },
                {
                    "id": "version-2",
                    "version": 1,
                    "dataset_id": "dataset5",
                    "edition": "edition5",
                    "state": "associated" 
                },
                {
                    "id": "version-3",
                    "version": 1,
                    "dataset_id": "dataset6",
                    "edition": "edition6",
                    "state": "associated" 
                },
                 {
                    "id": "version-4",
                    "version": 1,
                    "dataset_id": "dataset7",
                    "edition": "edition7",
                    "state": "associated" 
                },
                 {
                    "id": "version-5",
                    "version": 1,
                    "dataset_id": "dataset8",
                    "edition": "edition8",
                    "state": "associated" 
                },
                {
                    "id": "version-6",
                    "version": 10,
                    "dataset_id": "dataset9",
                    "edition": "edition9",
                    "state": "associated" 
                },
                {
                    "id": "version-7",
                    "version": 7,
                    "dataset_id": "dataset10",
                    "edition": "edition10",
                    "state": "published" 
                }
            ]
            """
            
    Scenario: PUT /bundles/{id}/state with valid arguments for 'APPROVED' -> 'PUBLISHED'
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-4"
        When I PUT "/bundles/bundle-4/state"
            """
                {
                    "state": "PUBLISHED"
                }
            """
        Then the HTTP status code should be "200"
        And the response should contain the following JSON response with a dynamic timestamp:
            """
                {
                    "id": "bundle-4",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "janedoe@example.com"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "PUBLISHED",
                    "title": "bundle-4",
                    "updated_at": "{{DYNAMIC_TIMESTAMP}}",
                    "managed_by": "WAGTAIL"
                }
            """
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        And bundle "bundle-4" should have state "PUBLISHED"
        And bundle "bundle-4" should not have this etag "etag-bundle-4"
        And these content item states should match:
            """
            [
                {
                    "id": "content-item-3",
                    "state": "PUBLISHED"
                }
            ]
            """
        And these dataset versions states should match:
            """
            [
                {
                    "id": "version-1",
                    "state": "published"
                }
            ]
            """
        And the total number of events should be 2
        And the number of events with action "UPDATE" and datatype "bundle" should be 1
        And the number of events with action "UPDATE" and datatype "content_item" should be 1

    Scenario: PUT /bundles/{id}/state with valid arguments for 'APPROVED' -> 'PUBLISHED' and the version is already published
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-5"
        When I PUT "/bundles/bundle-5/state"
            """
                {
                    "state": "PUBLISHED"
                }
            """
        Then the HTTP status code should be "200"
        And the response should contain the following JSON response with a dynamic timestamp:
            """
                {
                    "id": "bundle-5",
                    "bundle_type": "SCHEDULED",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "janedoe@example.com"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "PUBLISHED",
                    "title": "bundle-5",
                    "updated_at": "{{DYNAMIC_TIMESTAMP}}",
                    "managed_by": "WAGTAIL"
                }
            """
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        And bundle "bundle-5" should have state "PUBLISHED"
        And bundle "bundle-5" should not have this etag "etag-bundle-5"
        And these content item states should match:
            """
            [
                {
                    "id": "content-item-18",
                    "state": "PUBLISHED"
                }
            ]
            """
        And these dataset versions states should match:
            """
            [
                {
                    "id": "version-7",
                    "state": "published"
                }
            ]
            """
        And the total number of events should be 2
        And the number of events with action "UPDATE" and datatype "bundle" should be 1
        And the number of events with action "UPDATE" and datatype "content_item" should be 1

    Scenario: PUT /bundles/{id}/state with no authentication
        Given I am not authenticated
        When I PUT "/bundles/bundle-1/state"
            """
                {
                    "state": "APPROVED"
                }
            """
        Then the HTTP status code should be "401"
        And the response body should be empty
        And bundle "bundle-1" should have state "IN_REVIEW"
        And bundle "bundle-4" should have this etag "etag-bundle-4"


    Scenario: PUT /bundles/{id}/state with missing etag
        Given I am an admin user
        When I PUT "/bundles/bundle-6/state"
            """
                {
                    "state": "PUBLISHED"
                }
            """
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
                {
                    "errors":[
                        {
                            "code": "BadRequest",
                            "description": "Unable to process request due to missing If-Match header."
                        }
                    ]
                }
            """
        And bundle "bundle-4" should have state "APPROVED"
        And bundle "bundle-4" should have this etag "etag-bundle-4"

    Scenario: PUT /bundles/{id}/state with an invalid state value
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-4"
        When I PUT "/bundles/bundle-4/state"
            """
                {
                    "state": "notavalidstate"
                }
            """
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
                {
                    "errors":[
                        {
                            "code": "BadRequest",
                            "description": "Unable to process request due to a malformed or invalid request body or query parameter."
                        }
                    ]
                }
            """
        And bundle "bundle-4" should have state "APPROVED"
        And bundle "bundle-4" should have this etag "etag-bundle-4"

    Scenario: PUT /bundles/{id}/state with an invalid state transition
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-2"
        When I PUT "/bundles/bundle-2/state"
            """
                {
                    "state": "PUBLISHED"
                }
            """
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
                {
                    "errors":[
                        {
                            "code": "BadRequest",
                            "description": "Unable to process request due to invalid state transition."
                        }
                    ]
                }
            """

    Scenario: PUT /bundles/{id}/state with missing bundle
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-4"
        When I PUT "/bundles/not-a-real-bundle/state"
            """
                {
                    "state": "PUBLISHED"
                }
            """
        Then the HTTP status code should be "404"
        And I should receive the following JSON response:
            """
                {
                    "errors":[
                        {
                            "code": "NotFound",
                            "description": "The requested resource does not exist."
                        }
                    ]
                }
            """
            
    Scenario: PUT /bundles/{id}/state with a bundle with no content items
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-6"
        When I PUT "/bundles/bundle-6/state"
            """
                {
                    "state": "PUBLISHED"
                }
            """
        Then the HTTP status code should be "404"
        And I should receive the following JSON response:
            """
                {
                    "errors":[
                        {
                            "code": "NotFound",
                            "description": "The requested resource does not exist."
                        }
                    ]
                }
            """
        And bundle "bundle-6" should have state "APPROVED"
        And bundle "bundle-6" should have this etag "etag-bundle-6"

    Scenario: PUT /bundles/{id}/state with a bundle with a missing version
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-10"
        When I PUT "/bundles/bundle-10/state"
            """
                {
                    "state": "APPROVED"
                }
            """
        Then the HTTP status code should be "500"
        And bundle "bundle-10" should have state "IN_REVIEW"
        And bundle "bundle-10" should have this etag "etag-bundle-10"


    Scenario: PUT /bundles/{id}/state with valid arguments for 'IN_REVIEW' -> 'APPROVED'
        Given I am an admin user
        And I set the "If-Match" header to "etag-bundle-3"
        When I PUT "/bundles/bundle-3/state"
            """
                {
                    "state": "APPROVED"
                }
            """
        Then the HTTP status code should be "200"
        And the response should contain the following JSON response with a dynamic timestamp:
            """
                {
                    "id": "bundle-3",
                    "bundle_type": "MANUAL",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-05T13:40:00Z",
                    "last_updated_by": {
                        "email": "janedoe@example.com"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "APPROVED",
                    "title": "bundle-3",
                    "updated_at": "{{DYNAMIC_TIMESTAMP}}",
                    "managed_by": "WAGTAIL"
                }
            """
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        And bundle "bundle-3" should have state "APPROVED"
        And bundle "bundle-3" should not have this etag "etag-bundle-3"
        And these content item states should match:
            """
                [
                    {
                        "id": "content-item-12",
                        "state": "APPROVED"
                    },
                    {
                        "id": "content-item-13",
                        "state": "APPROVED"
                    },
                    {
                        "id": "content-item-14",
                        "state": "APPROVED"
                    },
                    {
                        "id": "content-item-15",
                        "state": "APPROVED"
                    },
                    {
                        "id": "content-item-16",
                        "state": "APPROVED"
                    }
                ]
            """
        And these dataset versions states should match:
            """
            [
               {
                    "id": "version-2",
                    "state": "approved" 
                },
                {
                    "id": "version-3",
                    "state": "approved" 
                },
                    {
                    "id": "version-4",
                    "state": "approved" 
                },
                    {
                    "id": "version-5",
                    "state": "approved" 
                },
                {
                    "id": "version-6",
                    "state": "approved" 
                }
            ]
            """
        And the total number of events should be 6
        And the number of events with action "UPDATE" and datatype "bundle" should be 1
        And the number of events with action "UPDATE" and datatype "content_item" should be 5