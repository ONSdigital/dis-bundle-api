Feature: Bundles API

  Background: we have some bundles in Mongo
    Given the bundles collection contains:
      """
      [
        {
          "id": "bundle-1",
          "bundle_type": "MANUAL",
          "created_at": "2025-05-01T10:00:00Z",
          "updated_at": "2025-05-10T12:00:00Z",
          "last_updated_by": { "email": "alice@example.com" },
          "preview_teams": [ { "id": "team-A" } ],
          "publish_date_time": "2025-05-15T09:00:00Z",
          "state": "ACTIVE",
          "title": "First bundle",
          "managed_by": "ADMIN"
        },
        {
          "id": "bundle-2",
          "bundle_type": "SCHEDULED",
          "created_at": "2025-05-02T11:00:00Z",
          "updated_at": "2025-05-11T13:00:00Z",
          "last_updated_by": { "email": "bob@example.com" },
          "preview_teams": [ { "id": "team-B" }, { "id": "team-C" } ],
          "publish_date_time": "2025-05-16T10:00:00Z",
          "state": "DRAFT",
          "title": "Second bundle",
          "managed_by": "USER"
        },
        {
        "id": "bundle-3",
        "bundle_type": "MANUAL",
        "created_at": "2025-05-03T08:00:00Z",
        "updated_at": "2025-05-12T14:00:00Z",
        "last_updated_by": { "email": "carol@example.com" },
        "preview_teams": [ { "id": "team-D" } ],
        "publish_date_time": "2025-05-17T11:00:00Z",
        "state": "PUBLISHED",
        "title": "Third bundle",
        "managed_by": "ADMIN"
        },
        {
        "id": "bundle-4",
        "bundle_type": "SCHEDULED",
        "created_at": "2025-05-04T09:30:00Z",
        "updated_at": "2025-05-13T15:15:00Z",
        "last_updated_by": { "email": "dave@example.com" },
        "preview_teams": [ { "id": "team-E" }, { "id": "team-F" } ],
        "publish_date_time": "2025-05-18T12:00:00Z",
        "state": "DRAFT",
        "title": "Fourth bundle",
        "managed_by": "USER"
        },
        {
        "id": "bundle-5",
        "bundle_type": "MANUAL",
        "created_at": "2025-05-05T10:45:00Z",
        "updated_at": "2025-05-14T16:30:00Z",
        "last_updated_by": { "email": "eve@example.com" },
        "preview_teams": [ { "id": "team-G" } ],
        "publish_date_time": "2025-05-19T13:00:00Z",
        "state": "ACTIVE",
        "title": "Fifth bundle",
        "managed_by": "ADMIN"
        },
        {
        "id": "bundle-6",
        "bundle_type": "SCHEDULED",
        "created_at": "2025-05-06T11:15:00Z",
        "updated_at": "2025-05-15T17:45:00Z",
        "last_updated_by": { "email": "frank@example.com" },
        "preview_teams": [ { "id": "team-H" }, { "id": "team-I" } ],
        "publish_date_time": "2025-05-20T14:00:00Z",
        "state": "DRAFT",
        "title": "Sixth bundle",
        "managed_by": "USER"
        },
        {
        "id": "bundle-7",
        "bundle_type": "MANUAL",
        "created_at": "2025-05-07T12:00:00Z",
        "updated_at": "2025-05-16T18:30:00Z",
        "last_updated_by": { "email": "grace@example.com" },
        "preview_teams": [ { "id": "team-J" } ],
        "publish_date_time": "2025-05-21T15:00:00Z",
        "state": "PUBLISHED",
        "title": "Seventh bundle",
        "managed_by": "ADMIN"
        }
      ]
    """
  Scenario: GET /bundles with no parameters returns default page
    When I GET "/bundles"
    Then the response status code should be 200
    And the response body should be a valid JSON page with:
      | count      | 20   |
      | offset     | 0    |
      | limit      | 20   |
      | total_count| 20  |
    And the JSON “items” array should contain exactly 20 bundles, sorted by updated_at descending
    And the response header "Cache-Control" should be "no-store"
    And the response header "ETag" should be present

  Scenario: GET /bundles with pagination parameters
    When I GET "/bundles?limit=5&offset=10"
    Then the response status code should be 200
    And the JSON page metadata should be:
      | count       | 5    |
      | offset      | 10   |
      | limit       | 5    |
      | total_count | 20  |
    And the “items” array should contain bundles 11–15 in updated_at DESC order

  Scenario: GET /bundles with invalid pagination parameters
    When I GET "/bundles?limit=-1"
    Then the response status code should be 400
    And the response body should be:
      """
      {
        "error_code": "ErrInvalidParameters",
        "description": "Unable to process request due to a malformed or invalid request body or query parameter."
      }
      """

  Scenario: GET /bundles without authentication
    Given I do not send an Authorization header
    When I GET "/bundles"
    Then the response status code should be 401
    And the response body should be:
      """
      {
        "error_code": "Unauthorised",
        "description": "Access denied."
      }
      """

  Scenario: GET /bundles when the service fails
    Given the bundle-store will return an error
    When I GET "/bundles"
    Then the response status code should be 500
    And the response body should be:
      """
      {
        "error_code": "InternalError",
        "description": "An internal error occurred"
      }
      """
