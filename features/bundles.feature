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
                    "title": "bundle-1",
                    "updated_at": "2025-04-03T11:25:00Z",
                    "managed_by": "WAGTAIL"
                },
                {
                    "id": "6835899f001ff1868922562c",
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
                    "id": "6835899f001ff1868922562e",
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
                    "state": "DRAFT",
                    "title": "bundle-3",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "WAGTAIL"
                }
            ]
            """

    Scenario: GET /bundles
        Given I am an admin user
        When I GET "/bundles"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        Then I should receive the following JSON response:
            """
            {
                "items": [
                    {
                        "id": "6835899f001ff1868922562e",
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
                        "state": "DRAFT",
                        "title": "bundle-3",
                        "updated_at": "2025-04-05T13:40:00Z",
                        "managed_by": "WAGTAIL"
                    },
                    {
                        "id": "6835899f001ff1868922562c",
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
                        "title": "bundle-1",
                        "updated_at": "2025-04-03T11:25:00Z",
                        "managed_by": "WAGTAIL"
                    }
                ],
                "count": 3,
                "limit": 20,
                "offset": 0,
                "total_count": 3
            }
            """
        And the response header "ETag" should not be empty
        And the response header "Cache-Control" should be "no-store"

    Scenario: GET /bundles?limit=1&offset=1
        Given I am an admin user
        When I GET "/bundles?limit=1"
        Then the HTTP status code should be "200"
        And the response header "ETag" should not be empty
        And I should receive the following JSON response:
            """
            {
                "items": [
                    {
                        "id": "6835899f001ff1868922562e",
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
                        "state": "DRAFT",
                        "title": "bundle-3",
                        "updated_at": "2025-04-05T13:40:00Z",
                        "managed_by": "WAGTAIL"
                    }
                ],
                "count": 1,
                "limit": 1,
                "offset": 0,
                "total_count": 3
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

    Scenario: GET /bundles with no authentication
        Given I am not authenticated
        When I GET "/bundles"
        Then the HTTP status code should be "401"
        And the response body should be empty


