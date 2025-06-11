Feature: List Single Bundle functionality - GET /Bundles/{bundle-id}

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
                    "state": "DRAFT",
                    "title": "bundle-1",
                    "updated_at": "2025-04-03T11:25:00Z",
                    "managed_by": "WAGTAIL"
                }
            ]
            """

    Scenario: GET /bundles/{bundle-id}
        Given I am an admin user
        When I GET "/bundles/bundle-1"
        Then the HTTP status code should be "200"
        Then I should receive the following JSON response:
            """
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
                "state": "DRAFT",
                "title": "bundle-1",
                "updated_at": "2025-04-03T11:25:00Z",
                "managed_by": "WAGTAIL"
            }
            """
        And the response header "ETag" should not be empty
        And the response header "ETag" should be "etag-bundle-1"

        And the response header "Cache-Control" should be "no-store"

    Scenario: GET /bundles/{bundle-id} with an invalid ID
        Given I am an admin user
        When I GET "/bundles/invalid-id"
        Then the HTTP status code should be "404"
        And I should receive the following JSON response:
            """
            {
                "errors": [
                    {
                        "code": "not_found",
                        "description": "The requested resource does not exist"
                    }
                ]
            }
            """

    Scenario: GET /bundles/{bundle-id} with no authentication
        Given I am not authenticated
        When I GET "/bundles/bundle-1"
        Then the HTTP status code should be "401"
        And the response body should be empty
