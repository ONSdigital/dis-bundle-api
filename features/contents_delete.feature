Feature: Delete a content item from a bundle - POST /bundles/{bundle-id}/contents/{content-id}

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
                "title": "bundle-1",
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
                "bundle_id": "bundle-1",
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
                "bundle_id": "bundle-2",
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
                },
                "state": "PUBLISHED"
            }
        ]
        """

    Scenario: DELETE /bundles/{bundle-id}/contents/{content-id} successfully
        Given I am an admin user
        When I DELETE "/bundles/bundle-1/contents/content-item-1"
        Then the HTTP status code should be "204"
        And the record with id "content-item-1" should not exist in the "bundle_contents" collection
    
    Scenario: DELETE /bundles/{bundle-id}/contents/{content-id} with non-existent bundle
        Given I am an admin user
        When I DELETE "/bundles/bundle-3/contents/content-item-1"
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

    Scenario: DELETE /bundles/{bundle-id}/contents/{content-id} with non-existent content item
        Given I am an admin user
        When I DELETE "/bundles/bundle-1/contents/content-item-3"
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

    Scenario: DELETE /bundles/{bundle-id}/contents/{content-id} with a content item that is published
        Given I am an admin user
        When I DELETE "/bundles/bundle-2/contents/content-item-2"
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

    Scenario: DELETE /bundles/{bundle-id}/contents/{content-id} with no authentication
        Given I am not authenticated
        When I DELETE "/bundles/bundle-1/contents/content-item-1"
        Then the HTTP status code should be "401"
        And the response body should be empty

    
