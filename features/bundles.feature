Feature: List Bundles functionality - GET /Bundles

    Background:
        Given I have these bundles:
            """
            [
                {
                    "id": "6835899f001ff18689225631",
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
                    "title": "bundle-10",
                    "updated_at": "2025-04-20T15:30:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "6835899f001ff1868922562c",
                    "bundle_type": "MANUAL",
                    "created_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "created_at": "2025-04-08T13:40:00Z",
                    "last_updated_by": {
                        "email": "publisher@ons.gov.uk"
                    },
                    "preview_teams": [
                        {
                            "id": "567j908h-98df-11ec-b909-0242ac120002"
                        }
                    ],
                    "state": "DRAFT",
                    "title": "bundle-8",
                    "updated_at": "2025-04-12T10:30:00Z",
                    "managed_by": "WAGTAIL"
                }
            ]
            """

    Scenario: GET /bundles
        Given I am an admin user

        When I GET "/bundles"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And I should receive the following JSON response:
            """
            {
                "items": [
                    {
                        "id": "6835899f001ff18689225631",
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
                        "title": "bundle-10",
                        "updated_at": "2025-04-20T15:30:00Z",
                        "managed_by": "WAGTAIL"
                    },
                    {
                        "id": "6835899f001ff1868922562c",
                        "bundle_type": "MANUAL",
                        "created_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "created_at": "2025-04-08T13:40:00Z",
                        "last_updated_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "preview_teams": [
                            {
                                "id": "567j908h-98df-11ec-b909-0242ac120002"
                            }
                        ],
                        "state": "DRAFT",
                        "title": "bundle-8",
                        "updated_at": "2025-04-12T10:30:00Z",
                        "managed_by": "WAGTAIL"
                    }
                ],
                "count": 2,
                "limit": 20,
                "offset": 0,
                "total_count": 2
            }
            """

    Scenario: GET /bundles?limit=1
        Given I am an admin user

        When I GET "/bundles?limit=1"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And I should receive the following JSON response:
            """
            {
                "items": [
                    {
                        "id": "6835899f001ff18689225631",
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
                        "title": "bundle-10",
                        "updated_at": "2025-04-20T15:30:00Z",
                        "managed_by": "WAGTAIL"
                    }
                ],
                "count": 1,
                "limit": 1,
                "offset": 0,
                "total_count": 2
            }
            """

    Scenario: GET /bundles with invalid offset
        Given I am an admin user

        When I GET "/bundles?offset=invalid"
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
            {
                "code": "bad_request",
                "description": "Unable to process request due to a malformed or invalid request body or query parameter",
                "source": {
                    "parameter": " offset"
                }
            }
            """
            
    Scenario: GET /bundles with invalid limit
        Given I am an admin user

        When I GET "/bundles?limit=invalid"
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
            {
                "code": "bad_request",
                "description": "Unable to process request due to a malformed or invalid request body or query parameter",
                "source": {
                    "parameter": " limit"
                }
            }
            """