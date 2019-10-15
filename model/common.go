package model

import "github.com/go-openapi/strfmt"

const (
	APICONFIGKEY  = "APICONFIG"
	TAG_TYPE      = "post_tag"
	CATEGORY_TYPE = "category"
)

// Base is struct that represent base of post, page, media data that also usually used inside _embed
type Base struct {
	// Unique identifier for the object.
	ID uint64 `json:"id,omitempty"`

	// The date the object was published, in the site's timezone.
	// Format: date-time
	Date strfmt.DateTime `json:"date,omitempty"`

	// An alphanumeric identifier for the object unique to its type.
	Slug string `json:"slug,omitempty"`

	// Type of Post for the object.
	Type string `json:"type,omitempty"`

	// URL to the object.
	Link string `json:"link,omitempty"`

	// title
	Title *Rendered `json:"title,omitempty"`

	// The id for the author of the object.
	Author uint64 `json:"author,omitempty"`
}

type BaseLink struct {
	SelfLink []map[string]string `json:"self"`

	Collection []map[string]string `json:"collection"`

	About []map[string]string `json:"about"`

	Author []EmbeddableLink `json:"author"`

	Replies []EmbeddableLink `json:"replies"`
}

// APIModel to store shared config
type APIModel struct {
	APIHost            string
	TablePrefix        string
	SiteURL            string
	PermalinkStructure string
	APIPath            string
	Version            string
	UploadPath         string
	APIBaseURL         string
}

// Content represents content in post json response
type ContentRendered struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected"`
}

// Rendered represents dictionary in json response with 'rendered' key
type Rendered struct {
	Rendered *string `json:"rendered"`
}

var PluralContentTypeMap = map[string]string{
	"post":     "posts",
	"page":     "pages",
	"category": "categories",
	"post_tag": "tags",
	"comment":  "comments",
}

func Plural(contentType string) string {
	return PluralContentTypeMap[contentType]
}

// IsParent check whether ID is part of slice of parent IDs
func IsParent(ID uint64, parentIDs []uint64) bool {
	for _, parentCommentID := range parentIDs {
		if ID == parentCommentID {
			return true
		}
	}
	return false
}
