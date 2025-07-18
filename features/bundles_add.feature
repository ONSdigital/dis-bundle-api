Feature: Create bundle - POST /Bundles

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

    Scenario: POST /bundles successfully
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-07-05T07:00:00.000Z",
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then the HTTP status code should be "201"
        And the response header "Content-Type" should be "application/json"
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        And the response header "Location" should not be empty

    Scenario: POST /bundles invalid body (missing double quotes)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED,
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-07-05T07:00:00.000Z",
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "BadRequest",
                            "description": "Unable to process request due to a malformed or invalid request body or query parameter."
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (invalid scheduled_at format)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "invalid-date-format",
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "InvalidParameters",
                            "description": "Invalid time format in request body.",
                            "source": {
                                "field": "scheduled_at"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (scheduled_at set for manual bundle)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "MANUAL",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-07-05T07:00:00.000Z",
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "InvalidParameters",
                            "description": "scheduled_at should not be set for manual bundles.",
                            "source": {
                                "field": "/scheduled_at"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (scheduled_at not set for scheduled bundle)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "InvalidParameters",
                            "description": "scheduled_at is required for scheduled bundles.",
                            "source": {
                                "field": "/scheduled_at"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (scheduled_at is set in the past for scheduled bundle)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2025-01-01T07:00:00Z",
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "InvalidParameters",
                            "description": "scheduled_at cannot be in the past.",
                            "source": {
                                "field": "/scheduled_at"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (missing required fields)
        Given I am an admin user
        When I POST "/bundles"
            """
                {}
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "MissingParameters",
                            "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                            "source": {
                                "field": "/bundle_type"
                            }
                        },
                        {
                            "code": "MissingParameters",
                            "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                            "source": {
                                "field": "/preview_teams"
                            }
                        },
                        {
                            "code": "MissingParameters",
                            "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                            "source": {
                                "field": "/state"
                            }
                        },
                        {
                            "code": "MissingParameters",
                            "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                            "source": {
                                "field": "/title"
                            }
                        },
                        {
                            "code": "MissingParameters",
                            "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                            "source": {
                                "field": "/managed_by"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (invalid fields)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "INVALID_TYPE",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-07-05T07:00:00.000Z",
                    "state": "INVALID_STATE",
                    "title": "Title of the Bundle",
                    "managed_by": "INVALID_MANAGED_BY"
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
                                "field": "/bundle_type"
                            }
                        },
                        {
                            "code": "InvalidParameters",
                            "description": "Unable to process request due to a malformed or invalid request body or query parameter.",
                            "source": {
                                "field": "/state"
                            }
                        },
                        {
                            "code": "InvalidParameters",
                            "description": "Unable to process request due to a malformed or invalid request body or query parameter.",
                            "source": {
                                "field": "/managed_by"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles bundle with the same title already exists
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-01-01T07:00:00Z",
                    "state": "DRAFT",
                    "title": "bundle-1",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "409":
            """
                {
                    "errors": [
                        {
                            "code": "Conflict",
                            "description": "A bundle with the same title already exists.",
                            "source": {
                                "field": "/title"
                            }
                        }
                    ]
                }
            """

    Scenario: POST /bundles invalid body (invalid state)
        Given I am an admin user
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-01-01T07:00:00Z",
                    "state": "APPROVED",
                    "title": "bundle-1",
                    "managed_by": "WAGTAIL"
                }
            """
        Then I should receive the following JSON response with status "400":
            """
                {
                    "errors": [
                        {
                            "code": "BadRequest",
                            "description": "state not allowed to transition."
                        }
                    ]
                }
            """
    Scenario: POST /bundles with no authentication
        Given I am not authenticated
        When I POST "/bundles"
            """
                {
                    "bundle_type": "SCHEDULED",
                    "preview_teams": [
                        {
                            "id": "team1"
                        },
                        {
                            "id": "team2"
                        }
                    ],
                    "scheduled_at": "2125-07-05T07:00:00.000Z",
                    "state": "DRAFT",
                    "title": "Title of the Bundle",
                    "managed_by": "WAGTAIL"
                }
            """
        Then the HTTP status code should be "401"
        And the response body should be empty