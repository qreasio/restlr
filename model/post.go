package model

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/qreasio/restlr/toolbox"
)

// Embedded represents _embedded json response of post
type Embedded struct {
	Author        []*User      `json:"author,omitempty"`
	FeaturedMedia []*BaseMedia `json:"wp:featuredmedia,omitempty"`
	Term          []*Term      `json:"wp:term,omitempty"`
	Replies       []*Comment   `json:"wp:replies,omitempty"`
}

// PostMeta represents a row from 'postmeta' table
type PostMeta struct {
	MetaID    uint64         `json:"meta_id"`    // meta_id
	PostID    uint64         `json:"post_id"`    // post_id
	MetaKey   sql.NullString `json:"meta_key"`   // meta_key
	MetaValue sql.NullString `json:"meta_value"` // meta_value
}

// SharedContent is struct that represent shared attributes for content type like post, page
type SharedContent struct {
	// excerpt
	Excerpt *ContentRendered `json:"excerpt,omitempty"`

	// The id of the featured media for the object.
	FeaturedMedia uint64 `json:"featured_media"`

	Links RestLink `json:"_links"`

	// The A password to protect access to the post. This only appears on view=edit
	Password *string `json:"password,omitempty"`
}

// ContentBase is the struct that that represent shared attributes of post and page for context = embed
type ContentBase struct {
	Base

	SharedContent

	Embedded *Embedded `json:"_embedded,omitempty"`
}

// ContentView is struct to represent shared attributes for context = view of every content type like post, page, media
type ContentView struct {
	Base

	// Whether or not comments are open on the object
	// Enum: [open closed]
	CommentStatus string `json:"comment_status,omitempty"`

	// content
	Content *ContentRendered `json:"content,omitempty"`

	// The date the object was published, as GMT.
	// Format: date-time
	DateGmt *strfmt.DateTime `json:"date_gmt,omitempty"`

	// meta
	Meta []map[string]string `json:"meta"`

	// Whether or not the object can be pinged.
	// Enum: [open closed]
	PingStatus string `json:"ping_status,omitempty"`

	// A named status for the object.
	// Enum: [publish future draft pending private]
	Status string `json:"status,omitempty"`

	// The theme file to use to display the object.
	Template string `json:"template,omitempty"`

	// The globally unique identifier for the object.
	GUID *Rendered `json:"guid,omitempty"`

	// The date the object was last modified, in the site's timezone.
	// Format: date-time
	Modified *strfmt.DateTime `json:"modified,omitempty"`

	// The date the object was last modified, as GMT.
	// Format: date-time
	ModifiedGmt *strfmt.DateTime `json:"modified_gmt,omitempty"`
}

// Post represent generic post and page data that will return to client
type Post struct {
	SharedContent

	ContentView

	// The format for the object.
	// Enum: [standard aside chat gallery link image quote status video audio]
	Format string `json:"format,omitempty"`

	// Whether or not the object should be treated as sticky.
	Sticky *bool `json:"sticky,omitempty"`

	Categories []uint64 `json:"categories,omitempty"`

	Tags []uint64 `json:"tags,omitempty"`

	// Only page
	MenuOrder *int `json:"menu_order,omitempty"`

	// Only for Page , the post id for the parent of the object.
	Parent *uint64 `json:"parent,omitempty"`

	// Only for media
	MimeType  *string `json:"mime_type,omitempty"`
	MediaType *string `json:"media_type,omitempty"`

	Embedded *Embedded `json:"_embedded,omitempty"`
}

// RawPost store unprocessed required raw data to construct post
type RawPost struct {
	Metas          map[uint64]map[string]string
	TermTaxonomies map[uint64][]*TermWithPostTaxonomy
	Taxonomies     map[uint64]map[string][]uint64
	FormatMap      map[uint64]string
	User           map[uint64]*UserDetail
	FeaturedMedia  map[uint64]uint64
	StickyPostIDs  map[int]bool
}

// SetViewAttributes will set post attributes that will be required on request with context = view
func (p *Post) SetViewAttributes(
	metas map[uint64]map[string]string,
	taxonomies map[uint64]map[string][]uint64,
	formatMap map[uint64]string) {

	// set Template
	template, ok := metas[p.ID]["_wp_page_template"]
	if ok {
		p.Template = template
	}

	if p.Type == PostType {
		// set Tags
		tags, ok := taxonomies[p.ID][TagType]
		if ok {
			p.Tags = tags
		}
		// set Categories
		categories, ok := taxonomies[p.ID][CategoryType]
		if ok {
			p.Categories = categories
		}
		// set Format
		format, ok := formatMap[p.ID]
		if ok {
			p.Format = strings.Replace(format, "post-format-", "", -1)
		}
	}

}

// SetFeaturedMediaID set featured media of post from map
func (p *Post) SetFeaturedMediaID(metas map[uint64]map[string]string) {
	if featuredMedia, ok := metas[p.ID]["_thumbnail_id"]; ok {
		if mediaID, err := strconv.ParseUint(featuredMedia, 10, 64); err == nil {
			p.FeaturedMedia = mediaID
		}
	}
}

// SetLinks will construct link metadata base on type
func (p *Post) SetLinks(ctx context.Context) {
	baseURL := ctx.Value(APIConfigKey).(APIConfig).APIBaseURL
	links := RestLink{}

	idStr := strconv.FormatUint(p.ID, 10)

	url := NewLinkURL(baseURL, p.Type)

	links.Collection = append(links.Collection, HrefMap(url.Collection()))

	links.SelfLink = append(links.SelfLink, HrefMap(url.Self(idStr)))

	links.About = append(links.About, HrefMap(url.About()))

	links.Author = append(links.Author, GetEmbeddableLink(url.Author(p.Author)))

	links.Replies = append(links.Replies, GetEmbeddableLink(url.Replies(idStr)))

	if p.FeaturedMedia != 0 {
		links.FeaturedMedia = append(links.FeaturedMedia, GetEmbeddableLink(url.FeaturedMedia(p.FeaturedMedia)))
	}

	links.VersionHistory = append(links.VersionHistory, HrefMap(url.Revisions(idStr)))

	links.Attachment = append(links.Attachment, HrefMap(url.Attachment(idStr)))

	links.Curies = append(links.Curies, &Curie{Name: "wp", Href: url.Curies(), Templated: true})

	if p.Type == PostType {
		links.Term = append(links.Term, TermPost{Href: url.Categories(idStr), Embeddable: true, Taxonomy: TagType})
		links.Term = append(links.Term, TermPost{Href: url.Tags(idStr), Embeddable: true, Taxonomy: CategoryType})
	}

	p.Links = links
}

// SetSticky set sticky value of post
func (p *Post) SetSticky(stickyPostIDs map[int]bool) {
	if stickyPostIDs[int(p.ID)] {
		p.Sticky = toolbox.BoolPointer(true)
	}
}

// SetPredecessorVersion set predecessor version of post
func (p *Post) SetPredecessorVersion(baseURL string, predecessor uint64) {
	url := PredecessorVersion(baseURL, p.ID, predecessor)
	versionLink := VersionLink{ID: predecessor, Href: url}
	p.Links.PredecessorVersion = append(p.Links.PredecessorVersion, versionLink)
}
