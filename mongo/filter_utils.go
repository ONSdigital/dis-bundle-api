package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// buildDateTimeFilter builds a bson filter for the datetime with a window of duration around the datetime
func buildDateTimeFilter(date time.Time) bson.M {
	const timeWindow = 2 * time.Second

	startTime := date.Add(timeWindow * -1)
	endTime := date.Add(timeWindow)

	return bson.M{
		"$gte": startTime,
		"$lte": endTime,
	}
}
