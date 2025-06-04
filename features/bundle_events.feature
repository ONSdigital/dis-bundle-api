Feature: List Bundle Events functionality - GET /bundle-events

   Background:
        Given I have these bundle events:
            """
            [
                {
                    "created_at": "2025-05-25T15:40:58.987Z",
                    "requested_by": {
                        "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                        "email": "publisher@ons.gov.uk"
                    },
                    "action": "CREATE",
                    "resource": "/bundles/bundle-10",
                    "bundle": {
                        "bundle_type": "SCHEDULED",
                        "created_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "created_at": "2025-04-04T07:00:00.000Z",
                        "id": "bundle-10",
                        "last_updated_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "preview_teams": [
                            {
                                "id": "1253e849-01fd-4662-bee2-63253538da93"
                            }
                        ],
                        "scheduled_at": "2025-04-05T07:00:00.000Z",
                        "state": "APPROVED",
                        "title": "CPI March 2025",
                        "updated_at": "2025-04-04T10:00:00.000Z",
                        "managed_by": "WAGTAIL"
                    }
                },
                {
                    "created_at": "2025-05-24T10:45:12.321Z",
                    "requested_by": {
                        "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                        "email": "publisher@ons.gov.uk"
                    },
                    "action": "CREATE",
                    "resource": "/bundles/bundle-4",
                    "bundle": {
                        "bundle_type": "MANUAL",
                        "created_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "created_at": "2025-04-04T07:00:00.000Z",
                        "id": "bundle-4",
                        "last_updated_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "preview_teams": [
                            {
                                "id": "1253e849-01fd-4662-bee2-63253538da93"
                            }
                        ],
                        "state": "APPROVED",
                        "title": "CPI March 2025",
                        "updated_at": "2025-04-04T10:00:00.000Z",
                        "managed_by": "WAGTAIL"
                    }
                },
                {
                    "created_at": "2025-05-23T09:30:42.111Z",
                    "requested_by": {
                        "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                        "email": "publisher@ons.gov.uk"
                    },
                    "action": "UPDATE",
                    "resource": "/bundles/bundle-1",
                    "bundle": {
                        "bundle_type": "MANUAL",
                        "created_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "created_at": "2025-04-04T07:00:00.000Z",
                        "id": "bundle-1",
                        "last_updated_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "preview_teams": [
                            {
                                "id": "1253e849-01fd-4662-bee2-63253538da93"
                            }
                        ],
                        "state": "APPROVED",
                        "title": "CPI March 2025",
                        "updated_at": "2025-04-04T10:00:00.000Z",
                        "managed_by": "WAGTAIL"
                    }
                },
                {
                    "created_at": "2025-05-22T08:22:33.444Z",
                    "requested_by": {
                        "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                        "email": "publisher@ons.gov.uk"
                    },
                    "action": "CREATE",
                    "resource": "/bundles/bundle-8",
                    "bundle": {
                        "bundle_type": "MANUAL",
                        "created_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "created_at": "2025-04-04T07:00:00.000Z",
                        "id": "bundle-8",
                        "last_updated_by": {
                            "email": "publisher@ons.gov.uk"
                        },
                        "preview_teams": [
                            {
                                "id": "1253e849-01fd-4662-bee2-63253538da93"
                            }
                        ],
                        "state": "APPROVED",
                        "title": "CPI March 2025",
                        "updated_at": "2025-04-04T10:00:00.000Z",
                        "managed_by": "WAGTAIL"
                    }
                }
            ]
            """
    
    Scenario: GET /bundle-events with default pagination
        Given I am an admin user
        When I GET "/bundle-events"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And the response header "Cache-Control" should be "no-store"
        And the response header "ETag" should be present
        And I should receive the following JSON response:
           """
           {
                "items": [
                    {
                        "created_at": "2025-05-25T15:40:58.987Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-10",
                        "bundle": {
                            "bundle_type": "SCHEDULED",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-10",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "scheduled_at": "2025-04-05T07:00:00Z",
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-24T10:45:12.321Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-4",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-4",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-23T09:30:42.111Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "UPDATE",
                        "resource": "/bundles/bundle-1",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-1",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-22T08:22:33.444Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-8",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-8",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    }
                ],
                "count": 4,
                "limit": 20,
                "offset": 0,
                "total_count": 4
           }
           """

    Scenario: GET /bundle-events with custom pagination
        Given I am an admin user
        When I GET "/bundle-events?limit=2&offset=1"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And I should receive the following JSON response:
           """
           {
                "items": [
                    {
                        "created_at": "2025-05-24T10:45:12.321Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-4",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-4",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-23T09:30:42.111Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "UPDATE",
                        "resource": "/bundles/bundle-1",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-1",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    }
                ],
                "count": 2,
                "limit": 2,
                "offset": 1,
                "total_count": 4
           }
           """

    Scenario: GET /bundle-events filtered by date range
        Given I am an admin user
        When I GET "/bundle-events?after=2025-05-01T00:00:00Z&before=2025-05-25T23:59:59Z"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And I should receive the following JSON response:
           """
           {
                "items": [
                    {
                        "created_at": "2025-05-25T15:40:58.987Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-10",
                        "bundle": {
                            "bundle_type": "SCHEDULED",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-10",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "scheduled_at": "2025-04-05T07:00:00Z",
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-24T10:45:12.321Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-4",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-4",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-23T09:30:42.111Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "UPDATE",
                        "resource": "/bundles/bundle-1",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-1",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    },
                    {
                        "created_at": "2025-05-22T08:22:33.444Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-8",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-8",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    }
                ],
                "count": 4,
                "limit": 20,
                "offset": 0,
                "total_count": 4
           }
           """

    Scenario: GET /bundle-events filtered by bundle ID
        Given I am an admin user
        When I GET "/bundle-events?bundle=bundle-4"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And I should receive the following JSON response:
           """
           {
                "items": [
                    {
                        "created_at": "2025-05-24T10:45:12.321Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-4",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-4",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    }
                ],
                "count": 1,
                "limit": 20,
                "offset": 0,
                "total_count": 1
           }
           """

    Scenario: GET /bundle-events with combined filters
        Given I am an admin user
        When I GET "/bundle-events?bundle=bundle-4&after=2025-05-01T00:00:00Z&limit=10"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json"
        And I should receive the following JSON response:
           """
           {
                "items": [
                    {
                        "created_at": "2025-05-24T10:45:12.321Z",
                        "requested_by": {
                            "id": "0889d599-3f0e-4564-9d6e-9455a6b73da7",
                            "email": "publisher@ons.gov.uk"
                        },
                        "action": "CREATE",
                        "resource": "/bundles/bundle-4",
                        "bundle": {
                            "bundle_type": "MANUAL",
                            "created_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "created_at": "2025-04-04T07:00:00Z",
                            "id": "bundle-4",
                            "last_updated_by": {
                                "email": "publisher@ons.gov.uk"
                            },
                            "preview_teams": [
                                {
                                    "id": "1253e849-01fd-4662-bee2-63253538da93"
                                }
                            ],
                            "state": "APPROVED",
                            "title": "CPI March 2025",
                            "updated_at": "2025-04-04T10:00:00Z",
                            "managed_by": "WAGTAIL"
                        }
                    }
                ],
                "count": 1,
                "limit": 10,
                "offset": 0,
                "total_count": 1
           }
           """

    Scenario: GET /bundle-events without authentication returns 401
        When I GET "/bundle-events"
        Then the HTTP status code should be "401"