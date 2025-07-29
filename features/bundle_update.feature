Feature: Update a bundle - PUT /bundles/{id}

    Background:
        Given I have these bundles:
            """
            [
                {
                    "id": "bundle-1",
                    "bundle_type": "MANUAL",
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
                    "state": "DRAFT",
                    "title": "Original Bundle Title",
                    "updated_at": "2025-04-03T11:25:00Z",
                    "managed_by": "DATA-ADMIN"
                },
                {
                    "id": "bundle-2",
                    "bundle_type": "SCHEDULED",
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
                    "scheduled_at": "2025-05-05T08:00:00Z",
                    "state": "IN_REVIEW",
                    "title": "Scheduled Bundle",
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
                    "state": "APPROVED",
                    "title": "Approved Bundle",
                    "updated_at": "2025-04-05T13:40:00Z",
                    "managed_by": "DATA-ADMIN"
                }
            ]
            """

Scenario: PUT /bundles/{id} successfully updates a bundle
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "title": "Updated Bundle Title",
                "managed_by": "DATA-ADMIN"
            }
            """
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should not be empty
        And the response should contain the following JSON response with a dynamic timestamp:
            """
            {
                "bundle_type": "MANUAL",
                "created_at": "2025-04-03T11:25:00Z",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "id": "bundle-1",
                "last_updated_by": {
                    "email": "janedoe@example.com"
                },
                "managed_by": "DATA-ADMIN",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "title": "Updated Bundle Title",
                "updated_at": "{{DYNAMIC_TIMESTAMP}}"
            }
            """
        And the total number of events should be 1
        And the number of events with action "UPDATE" and datatype "bundle" should be 1

    Scenario: PUT /bundles/{id} with state transition from DRAFT to IN_REVIEW
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "IN_REVIEW",
                "title": "Bundle Moving to Review",
                "managed_by": "DATA-ADMIN"
            }
            """
        Then the HTTP status code should be "200"
        And the response should contain the following JSON response with a dynamic timestamp:
            """
            {
                "bundle_type": "MANUAL",
                "created_at": "2025-04-03T11:25:00Z",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "id": "bundle-1",
                "last_updated_by": {
                    "email": "janedoe@example.com"
                },
                "managed_by": "DATA-ADMIN",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "IN_REVIEW",
                "title": "Bundle Moving to Review",
                "updated_at": "{{DYNAMIC_TIMESTAMP}}"
            }
            """
        And the total number of events should be 1
        And the number of events with action "UPDATE" and datatype "bundle" should be 1

    Scenario: PUT /bundles/{id} with state transition from APPROVED to PUBLISHED
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-3"
        When I PUT "/bundles/bundle-3"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "567j908h-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "PUBLISHED",
                "title": "Published Bundle",
                "managed_by": "DATA-ADMIN"
            }
            """
        Then the HTTP status code should be "200"
        And the response should contain the following JSON response with a dynamic timestamp:
            """
            {
                "bundle_type": "MANUAL",
                "created_at": "2025-04-05T13:40:00Z",
                "created_by": {
                    "email": "publisher@ons.gov.uk"
                },
                "id": "bundle-3",
                "last_updated_by": {
                    "email": "janedoe@example.com"
                },
                "managed_by": "DATA-ADMIN",
                "preview_teams": [
                    {
                        "id": "567j908h-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "PUBLISHED",
                "title": "Published Bundle",
                "updated_at": "{{DYNAMIC_TIMESTAMP}}"
            }
            """
        And the total number of events should be 1
        And the number of events with action "UPDATE" and datatype "bundle" should be 1

    Scenario: PUT /bundles/{id} without If-Match header returns 400
        Given I am an admin user
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "title": "Updated Title",
                "managed_by": "DATA-ADMIN"
            }
            """
        Then I should receive the following JSON response with status "400":
            """
            {
                "errors": [
                    {
                        "code": "MissingParameters",
                        "description": "Unable to process request due to missing If-Match header."
                    }
                ]
            }
            """

    Scenario: PUT /bundles/{id} with mismatched ETag returns 409
        Given I am an admin user
        And I set the header "If-Match" to "wrong-etag"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "title": "Updated Title",
                "managed_by": "DATA-ADMIN"
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

    Scenario: PUT /bundles/{id} with invalid bundle data returns 400
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "INVALID_TYPE",
                "preview_teams": [],
                "state": "INVALID_STATE",
                "title": "",
                "managed_by": "INVALID_MANAGER"
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
                        "code": "MissingParameters",
                        "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                        "source": {
                            "field": "/preview_teams"
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
                        "code": "MissingParameters",
                        "description": "Unable to process request due to missing required parameters in the request body or query parameters.",
                        "source": {
                            "field": "/title"
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

    Scenario: PUT /bundles/{id} with duplicate title returns 400
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "title": "Scheduled Bundle",
                "managed_by": "DATA-ADMIN"
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
                            "field": "/title"
                        }
                    }
                ]
            }
            """


  Scenario: PUT /bundles/{id} with SCHEDULED type missing scheduled_at returns 400
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "SCHEDULED",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "title": "Missing Scheduled Date",
                "managed_by": "DATA-ADMIN"
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
                            "field": "/scheduled_at"
                        }
                    }
                ]
            }
            """

   Scenario: PUT /bundles/{id} with MANUAL type and scheduled_at returns 400
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "scheduled_at": "2025-06-01T10:00:00Z",
                "title": "Manual with Scheduled Date",
                "managed_by": "DATA-ADMIN"
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
                            "field": "/scheduled_at"
                        }
                    }
                ]
            }
            """

   Scenario: PUT /bundles/{id} with past scheduled_at returns 400
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "SCHEDULED",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "DRAFT",
                "scheduled_at": "2020-01-01T10:00:00Z",
                "title": "Past Scheduled Date",
                "managed_by": "DATA-ADMIN"
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
                            "field": "/scheduled_at"
                        }
                    }
                ]
            }
            """

    Scenario: PUT /bundles/{id} with invalid state transition returns 400
        Given I am an admin user
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "state": "PUBLISHED",
                "title": "Invalid Transition",
                "managed_by": "DATA-ADMIN"
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
                            "field": "/state"
                        }
                    }
                ]
            }
            """

    Scenario: PUT /bundles/{id} for non-existent bundle returns 404
        Given I am an admin user
        And I set the header "If-Match" to "some-etag"
        When I PUT "/bundles/bundle-missing"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "title": "Missing Bundle",
                "managed_by": "DATA-ADMIN"
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

    Scenario: PUT /bundles/{id} without authentication returns 401
        Given I am not authenticated
        And I set the header "If-Match" to "etag-bundle-1"
        When I PUT "/bundles/bundle-1"
            """
            {
                "bundle_type": "MANUAL",
                "preview_teams": [
                    {
                        "id": "890m231k-98df-11ec-b909-0242ac120002"
                    }
                ],
                "title": "Updated Title",
                "managed_by": "DATA-ADMIN"
            }
            """
        Then the HTTP status code should be "401"
        And the response body should be empty