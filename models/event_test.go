package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateEvent(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the event has all required fields", func() {
			now := time.Now()
			requestedBy := RequestedBy{
				ID:    "user123",
				Email: "user@example.com",
			}
			testEvent := Event{
				CreatedAt:   now,
				RequestedBy: requestedBy,
				Action:      ActionCreate,
				Resource:    "/bundles/123/contents/456",
				Data: &Data{
					DatasetID: "cpih",
					EditionID: "march-2025",
					ItemID:    "de3bc0b6-d6c4-4e20-917e-95d7ea8c91dc",
					State:     "published",
					URLPath:   "/datasets/cpih/editions/march-2025/versions/1",
					Title:     "cpih",
				},
			}
			b, err := json.Marshal(testEvent)
			if err != nil {
				t.Logf("failed to marshal test data into bytes, error: %v", err)
				t.FailNow()
			}
			reader := bytes.NewReader(b)
			event, err := CreateEvent(reader)
			So(err, ShouldBeNil)
			So(event, ShouldNotBeNil)
			So(event.CreatedAt, ShouldEqual, now)
			So(event.RequestedBy, ShouldResemble, requestedBy)
			So(event.Action, ShouldEqual, ActionCreate)
			So(event.Resource, ShouldEqual, "/bundles/123/contents/456")
			So(event.Data, ShouldResemble, &Data{
				DatasetID: "cpih",
				EditionID: "march-2025",
				ItemID:    "de3bc0b6-d6c4-4e20-917e-95d7ea8c91dc",
				State:     "published",
				URLPath:   "/datasets/cpih/editions/march-2025/versions/1",
				Title:     "cpih",
			})
		})

		Convey("when the event has minimal required fields", func() {
			// Test with minimal fields (no CreatedAt or Data)
			jsonStr := `{
				"requested_by": {"id": "user456"},
				"action": "READ",
				"resource": "/bundles/789"
			}`
			reader := bytes.NewReader([]byte(jsonStr))
			event, err := CreateEvent(reader)
			So(err, ShouldBeNil)
			So(event, ShouldNotBeNil)
			So(event.RequestedBy.ID, ShouldEqual, "user456")
			So(event.Action, ShouldEqual, ActionRead)
			So(event.Resource, ShouldEqual, "/bundles/789")
			So(event.Data, ShouldBeNil)
		})
	})

	Convey("Return error when unable to parse json", t, func() {
		Convey("when the json is invalid syntax", func() {
			b := `{"action":"invalid-body}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateEvent(reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
		})

		Convey("when the json contains invalid data", func() {
			b := `{"requested_by":{"id":"user123"},"action":"INVALID_ACTION","resource":"/bundles/123"}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateEvent(reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
		})
	})
}

func TestActionMarshalJSON(t *testing.T) {
	Convey("Given an event", t, func() {
		now := time.Now()
		requestedBy := RequestedBy{
			ID:    "user123",
			Email: "user@example.com",
		}
		testEvent := Event{
			CreatedAt:   now,
			RequestedBy: requestedBy,
			Action:      ActionCreate,
			Resource:    "/bundles/123/contents/456",
			Data: &Data{
				DatasetID: "cpih",
				EditionID: "march-2025",
				ItemID:    "de3bc0b6-d6c4-4e20-917e-95d7ea8c91dc",
				State:     "published",
				URLPath:   "/datasets/cpih/editions/march-2025/versions/1",
				Title:     "cpih",
			},
		}

		Convey("when the action is valid", func() {
			Convey("then it should marshal without error", func() {
				_, err := json.Marshal(testEvent)
				So(err, ShouldBeNil)
			})
		})

		Convey("when the action is invalid", func() {
			testEvent.Action = "INVALID"
			Convey("then it should return an error", func() {
				_, err := json.Marshal(testEvent)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid Action")
			})
		})

		Convey("all valid actions should marshal and unmarshal correctly", func() {
			actions := []Action{ActionCreate, ActionRead, ActionUpdate, ActionDelete}

			for _, action := range actions {
				testEvent.Action = action
				marshaled, err := json.Marshal(testEvent)
				So(err, ShouldBeNil)

				var unmarshaled Event
				err = json.Unmarshal(marshaled, &unmarshaled)
				So(err, ShouldBeNil)
				So(unmarshaled.Action, ShouldEqual, action)
			}
		})
	})
}

func TestActionUnmarshalJSON(t *testing.T) {
	Convey("Given invalid JSON for Action", t, func() {
		invalidJSON := []byte(`123`)
		var action Action

		Convey("When UnmarshalJSON is called", func() {
			err := action.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given invalid value for Action", t, func() {
		invalidJSON := []byte(`"INVALID"`)
		var action Action

		Convey("When UnmarshalJSON is called", func() {
			err := action.UnmarshalJSON(invalidJSON)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid Action: INVALID")
			})
		})
	})

	Convey("Given valid value for Action", t, func() {
		validJSON := []byte(`"CREATE"`)
		var action Action

		Convey("When UnmarshalJSON is called", func() {
			err := action.UnmarshalJSON(validJSON)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
				So(action, ShouldEqual, ActionCreate)
			})
		})
	})
}

func TestValidateEvent(t *testing.T) {
	Convey("Given a valid event", t, func() {
		now := time.Now()
		requestedBy := RequestedBy{
			ID:    "user123",
			Email: "user@example.com",
		}
		testEvent := Event{
			CreatedAt:   now,
			RequestedBy: requestedBy,
			Action:      ActionCreate,
			Resource:    "/bundles/123/contents/456",
			Data: &Data{
				DatasetID: "cpih",
				EditionID: "march-2025",
			},
		}

		Convey("When ValidateEvent is called with a valid event", func() {
			err := ValidateEvent(&testEvent)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When ValidateEvent is called and RequestedBy.ID is empty", func() {
			testEvent.RequestedBy.ID = ""
			Convey("Then it should return an error", func() {
				err := ValidateEvent(&testEvent)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"requested_by.id"}))
			})
		})

		Convey("When ValidateEvent is called and Action is empty", func() {
			testEvent.Action = ""
			Convey("Then it should return an error", func() {
				err := ValidateEvent(&testEvent)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"action"}))
			})
		})

		Convey("When ValidateEvent is called and Resource is empty", func() {
			testEvent.Resource = ""
			Convey("Then it should return an error", func() {
				err := ValidateEvent(&testEvent)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("missing mandatory fields: %v", []string{"resource"}))
			})
		})

		Convey("When ValidateEvent is called and all mandatory fields are empty", func() {
			testEvent.RequestedBy.ID = ""
			testEvent.Action = ""
			testEvent.Resource = ""
			Convey("Then it should return an error", func() {
				err := ValidateEvent(&testEvent)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "missing mandatory fields")
				So(err.Error(), ShouldContainSubstring, "requested_by.id")
				So(err.Error(), ShouldContainSubstring, "action")
				So(err.Error(), ShouldContainSubstring, "resource")
			})
		})
	})
}

func TestEventsList(t *testing.T) {
	Convey("EventsList properly embeds PaginationFields", t, func() {
		now := time.Now()
		event := Event{
			CreatedAt: now,
			RequestedBy: RequestedBy{
				ID:    "user123",
				Email: "user@example.com",
			},
			Action:   ActionCreate,
			Resource: "/bundles/123/contents/456",
			Data: &Data{
				DatasetID: "cpih",
				EditionID: "march-2025",
				ItemID:    "de3bc0b6-d6c4-4e20-917e-95d7ea8c91dc",
				State:     "published",
				URLPath:   "/datasets/cpih/editions/march-2025/versions/1",
				Title:     "cpih",
			},
		}

		events := []Event{event}
		eventsList := EventsList{
			PaginationFields: PaginationFields{
				Count:      1,
				Limit:      20,
				Offset:     0,
				TotalCount: 1,
			},
			Items: events,
		}

		Convey("it should have the correct pagination fields", func() {
			So(eventsList.Count, ShouldEqual, 1)
			So(eventsList.Limit, ShouldEqual, 20)
			So(eventsList.Offset, ShouldEqual, 0)
			So(eventsList.TotalCount, ShouldEqual, 1)
		})

		Convey("it should have the correct items", func() {
			So(len(eventsList.Items), ShouldEqual, 1)
			So(eventsList.Items[0], ShouldResemble, event)
		})

		Convey("it should marshal to JSON correctly", func() {
			bytes, err := json.Marshal(eventsList)
			So(err, ShouldBeNil)

			var unmarshaled map[string]interface{}
			err = json.Unmarshal(bytes, &unmarshaled)
			So(err, ShouldBeNil)

			So(unmarshaled["count"], ShouldEqual, float64(1))
			So(unmarshaled["limit"], ShouldEqual, float64(20))
			So(unmarshaled["offset"], ShouldEqual, float64(0))
			So(unmarshaled["total_count"], ShouldEqual, float64(1))
			So(unmarshaled["items"], ShouldNotBeNil)
		})
	})
}

func TestAction_IsValid(t *testing.T) {
	Convey("IsValid should return true for valid Action values", t, func() {
		So(ActionCreate.IsValid(), ShouldBeTrue)
		So(ActionRead.IsValid(), ShouldBeTrue)
		So(ActionUpdate.IsValid(), ShouldBeTrue)
		So(ActionDelete.IsValid(), ShouldBeTrue)
	})

	Convey("IsValid should return false for invalid Action values", t, func() {
		So(Action("INVALID").IsValid(), ShouldBeFalse)
	})
}

func TestAction_String(t *testing.T) {
	Convey("String should return the string representation of Action", t, func() {
		So(ActionCreate.String(), ShouldEqual, "CREATE")
		So(ActionRead.String(), ShouldEqual, "READ")
		So(ActionUpdate.String(), ShouldEqual, "UPDATE")
		So(ActionDelete.String(), ShouldEqual, "DELETE")
	})
}
