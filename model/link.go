package model

import (
	"fmt"
	"strconv"
)

type LinkURL struct {
	BaseURL       string
	Type          string
	IsContentType bool
}

func NewLinkURL(baseURL string, contentType string) LinkURL {
	return LinkURL{
		BaseURL: baseURL,
		Type:    contentType,
	}
}

// VersionLink represents json response that consists of two elements, they are id and href
type VersionLink struct {
	ID   uint64 `json:"id"`
	Href string `json:"href"`
}

// RestLink represents _links in EmbedResponse
type RestLink struct {
	BaseLink
	VersionHistory     []map[string]string `json:"version-history"`
	PredecessorVersion []VersionLink       `json:"predecessor-version,omitempty"`
	FeaturedMedia      []EmbeddableLink    `json:"wp:featuredmedia,omitempty"`
	Attachment         []map[string]string `json:"wp:attachment,omitempty"`
	Term               []TermPost          `json:"wp:term,omitempty"`
	Curies             []*Curie            `json:"curies,omitempty"`
}

// EmbeddableLink represents json response that consists of two elements, they are embeddable and href
type EmbeddableLink struct {
	Embeddable bool   `json:"embeddable,omitempty"`
	Href       string `json:"href,omitempty"`
}

func (t *LinkURL) AboutPrefix() string {
	if t.Type == "category" || t.Type == "post_tag" {
		return "taxonomies"
	}
	return "types"
}

func (t *LinkURL) Replies(id string) string {
	return fmt.Sprintf("%s/comments?post=%s", t.BaseURL, id)
}

func (t *LinkURL) Author(userID uint64) string {
	return fmt.Sprintf("%s/users/%s", t.BaseURL, strconv.FormatUint(userID, 10))
}

func (t *LinkURL) About() string {
	return fmt.Sprintf("%s/%s/%s", t.BaseURL, t.AboutPrefix(), t.Type)
}

func (t *LinkURL) Self(id string) string {
	return t.BaseURL + "/categories/" + id
	return fmt.Sprintf("%s/%s/%s", t.BaseURL, Plural(t.Type), id)
}

func (t *LinkURL) Collection() string {
	return fmt.Sprintf("%s/%s/", t.BaseURL, Plural(t.Type))
}

func (t *LinkURL) PostType(id string) string {
	return fmt.Sprintf("%s/posts/?%s=%s", t.BaseURL, t.Type, id)
}

func (t *LinkURL) FeaturedMedia(id uint64) string {
	featuredMedia := strconv.FormatUint(id, 10)
	return fmt.Sprintf("%s/media/%s", t.BaseURL, featuredMedia)
}

func (t *LinkURL) Revisions(id string) string {
	return fmt.Sprintf("%s/%s/%s/revisions", t.BaseURL, Plural(t.Type), id)
}

func (t *LinkURL) Attachment(id string) string {
	return fmt.Sprintf("%s/media?parent=%s", t.BaseURL, id)
}

func (t *LinkURL) Categories(id string) string {
	return fmt.Sprintf("%s/categories?post=%s", t.BaseURL, id)
}

func (t *LinkURL) Tags(id string) string {
	return fmt.Sprintf("%s/tags?post=%s", t.BaseURL, id)
}

func (t *LinkURL) Curies() string {
	return "https://api.w.org/{rel}"
}

func PredecessorVersion(baseURL string, postID uint64, predeccessorID uint64) string {
	return fmt.Sprintf("%s/posts/%d/revisions/%d", baseURL, postID, predeccessorID)
}

func CategoryLink(baseURL string, slug string) string {
	return fmt.Sprintf("%s/category/%s", baseURL, slug)
}

func TagLink(baseURL string, slug string) string {
	return fmt.Sprintf("%s/tag/%s", baseURL, slug)
}

func GetTermLinks(postType string, baseURL string, id string) *TermLink {
	tLink := &TermLink{}

	url := NewLinkURL(baseURL, postType)

	tLink.SelfLink = append(tLink.SelfLink, HrefMap(url.Self(id)))

	tLink.Collection = append(tLink.Collection, HrefMap(url.Collection()))

	tLink.About = append(tLink.About, HrefMap(url.About()))

	tLink.PostType = append(tLink.PostType, HrefMap(url.PostType(id)))

	tLink.Curies = append(tLink.Curies, &Curie{Name: "wp", Href: url.Curies(), Templated: true})

	return tLink
}
