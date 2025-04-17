package models

import (
	"encoding/json"
	"io"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

type Bundle struct {
	ID              string          `bson:"id" json:"id"`
	BundleType      string          `bson:"bundle_type" json:"bundle_type"`
	Contents        []BundleContent `bson:"contents" json:"contents"`
	Creator         string          `bson:"creator" json:"creator"`
	CreatedDate     time.Time       `bson:"created_date" json:"created_date"`
	LastUpdatedBy   string          `bson:"last_updated_by" json:"last_updated_by"`
	PreviewTeams    []string        `bson:"preview_teams" json:"preview_teams"`
	PublishDateTime time.Time       `bson:"publish_date_time" json:"publish_date_time"`
	State           string          `bson:"state" json:"state"`
	Title           string          `bson:"title" json:"title"`
	UpdatedDate     time.Time       `bson:"updated_date" json:"updated_date"`
	WagtailManaged  bool            `bson:"wagtail_managed" json:"wagtail_managed"`
}

type BundleContent struct {
	DatasetID string `bson:"dataset_id" json:"dataset_id"`
	EditionID string `bson:"edition_id" json:"edition_id"`
	ItemID    string `bson:"item_id" json:"item_id"`
	State     string `bson:"state" json:"state"`
	Title     string `bson:"title" json:"title"`
	URLPath   string `bson:"url_path" json:"url_path"`
}

func CreateBundle(reader io.Reader) (*Bundle, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var bundle Bundle

	err = json.Unmarshal(b, &bundle)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	return &bundle, nil
}
