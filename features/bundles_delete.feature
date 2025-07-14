Feature: Delete a bundle and all its associated content items - DELETE /bundles/{bundle-id}

    Background:
        Given I have these bundles:
        """
        [
            {
                "id": "bundle-with-content-items",
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
                "title": "bundle-with-content-items",
                "updated_at": "2025-01-02T07:00:00Z",
                "managed_by": "WAGTAIL",
                "e_tag": "original-etag"
            },
            {
                "id": "bundle-without-content-items",
                "bundle_type": "SCHEDULED",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "created_at": "2025-01-04T07:00:00Z",
                "last_updated_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "scheduled_at": "2025-01-06T07:00:00Z",
                "state": "DRAFT",
                "title": "bundle-without-content-items",
                "updated_at": "2025-01-06T07:00:00Z",
                "managed_by": "WAGTAIL",
                "e_tag": "original-etag"
            },
            {
                "id": "bundle-published",
                "bundle_type": "SCHEDULED",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "created_at": "2025-01-04T07:00:00Z",
                "last_updated_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "scheduled_at": "2025-01-06T07:00:00Z",
                "state": "PUBLISHED",
                "title": "bundle-published",
                "updated_at": "2025-01-06T07:00:00Z",
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
                "bundle_id": "bundle-with-content-items",
                "content_type": "DATASET",
                "metadata": {
                    "dataset_id": "dataset1",
                    "edition_id": "edition1",
                    "version_id": 1,
                    "title": "Test Dataset 1"
                },
                "links": {
                    "edit": "edit/link",
                    "preview": "preview/link"
                }
            },
            {
                "id": "content-item-2",
                "bundle_id": "bundle-with-content-items",
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

    Scenario: DELETE /bundles/{bundle-id} successfully with a bundle that has contents
        Given I am an admin user
        When I DELETE "/bundles/bundle-with-content-items"
        Then the HTTP status code should be "204"
        And the record with id "content-item-1" should not exist in the "bundle_contents" collection
        And the record with id "content-item-2" should not exist in the "bundle_contents" collection
        And the record with id "bundle-with-content-items" should not exist in the "bundles" collection
    
    Scenario: DELETE /bundles/{bundle-id} successfully with a bundle that has no contents
        Given I am an admin user
        When I DELETE "/bundles/bundle-without-content-items"
        Then the HTTP status code should be "204"
        And the record with id "bundle-without-content-items" should not exist in the "bundles" collection

    Scenario: DELETE /bundles/{bundle-id} with non-existent bundle
        Given I am an admin user
        When I DELETE "/bundles/missing-bundle"
        Then I should receive the following JSON response with status "404":
            """
            {
                "errors": [
                    {
                        "code": "NotFound",
                        "description": "The requested resource does not exist"
                    }
                ]
            }
            """
    
    Scenario: DELETE /bundles/{bundle-id} with a published bundle
        Given I am an admin user
        When I DELETE "/bundles/bundle-published"
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
    
    Scenario: DELETE /bundles/{bundle-id} with no authentication
        Given I am not authenticated
        When I DELETE "/bundles/bundle-with-content-items"
        Then the HTTP status code should be "401"
        And the response body should be empty
