Feature: Get bundle as preview user - GET /bundles/{bundle-id}

    Background:
        Given I have these bundles:
            """
            [
                {
                    "id": "test-bundle-1",
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
                            "id": "team-preview-1"
                        }
                    ],
                    "state": "DRAFT",
                    "title": "test-bundle-1",
                    "updated_at": "2025-04-03T11:25:00Z",
                    "managed_by": "WAGTAIL"
                }
            ]
            """
        And I have these content items:
            """
            [
                {
                    "id": "test-content-1",
                    "bundle_id": "test-bundle-1",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "test-static-dataset-1",
                        "edition_id": "time-series",
                        "version_id": 1
                    },
                    "links": {
                        "edit": "/datasets/test-static-dataset-1/editions/time-series/versions/1",
                        "preview": "/datasets/test-static-dataset-1/editions/time-series/versions/1"
                    }
                },
                {
                    "id": "content-2",
                    "bundle_id": "test-bundle-1",
                    "content_type": "DATASET",
                    "metadata": {
                        "dataset_id": "test-static-dataset-2",
                        "edition_id": "2026",
                        "version_id": 1
                    },
                    "links": {
                        "edit": "/datasets/test-static-dataset-2/editions/2026/versions/1",
                        "preview": "/datasets/test-static-dataset-2/editions/2026/versions/1"
                    }
                }
            ]
            """

    Scenario: Preview user with permission to read bundle receives 200
        Given I am a preview user
        And I have preview access to these dataset editions:
            """
            [
                "test-static-dataset-1/time-series",
                "test-static-dataset-2/2026"
            ]
            """
        When I GET "/bundles/test-bundle-1"
        Then the HTTP status code should be "200"
        When I GET "/bundles/test-bundle-1/contents"
        Then the HTTP status code should be "200"

    Scenario: Preview user without permission to read bundle receives 403
        Given I am a preview user
        When I GET "/bundles/test-bundle-1"
        Then the HTTP status code should be "403"

