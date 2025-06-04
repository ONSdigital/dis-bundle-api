package mongo

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildDateTimeFilter(t *testing.T) {
	t.Parallel()

	Convey("When we call buildDateTimeFilter", t, func() {
		dateTime := time.Date(2025, 01, 01, 01, 01, 01, 01, time.UTC)

		filter := buildDateTimeFilter(dateTime)
		Convey("Then it should return a bson.M object with a gte and lte around that time", func() {
			duration := 2 * time.Second
			expectedResult := bson.M{
				"$lte": dateTime.Add(duration),
				"$gte": dateTime.Add(duration * -1),
			}

			So(filter, ShouldResemble, expectedResult)
		})
	})
}
