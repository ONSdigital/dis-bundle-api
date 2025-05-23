package models

import (
	"encoding/json"
	"io"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

type Bundle struct {
	ID              string          `bson:"_id" json:"id"`
	BundleType      string          `bson:"bundle_type" json:"bundle_type"`
	Contents        []BundleContent `bson:"contents" json:"contents"`
	CreatedDate     time.Time       `bson:"created_date" json:"created_date"`
	LastUpdatedBy   User            `bson:"last_updated_by" json:"last_updated_by"`
	PreviewTeams    []PreviewTeam   `bson:"preview_teams" json:"preview_teams"`
	PublishDateTime time.Time       `bson:"publish_date_time" json:"publish_date_time"`
	State           string          `bson:"state" json:"state"`
	Title           string          `bson:"title" json:"title"`
	UpdatedDate     time.Time       `bson:"updated_at" json:"updated_at"`
	WagtailManaged  bool            `bson:"wagtail_managed" json:"wagtail_managed"`
	ETag            string          `bson:"e_tag"                           json:"-"`
}

type BundleContent struct {
	DatasetID string `bson:"dataset_id" json:"dataset_id"`
	EditionID string `bson:"edition_id" json:"edition_id"`
	ItemID    string `bson:"item_id" json:"item_id"`
	State     string `bson:"state" json:"state"`
	Title     string `bson:"title" json:"title"`
	URLPath   string `bson:"url_path" json:"url_path"`
}

type PaginationFields struct {
	Count      int `json:"count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
}

type User struct {
	Email string `bson:"email,omitempty" json:"email,omitempty"`
}

type PreviewTeam struct {
	ID string `bson:"id" json:"id"`
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
