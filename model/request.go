package model

// GetItemRequest is struct to represents common HTTP URL request values to get specific post or item in API
type GetItemRequest struct {
	ID       *uint64
	Context  string  `form:"context"`
	Password *string `form:"password"`
	Embed    *string `form:"_embed"`
	IsEmbed  bool
}

// ListRequest represents query string to browse/list post/page with context and embed parameter
// This model is used inside service
type ListRequest struct {
	ListParams
	Context *string `form:"context"`
	Embed   *string
	IsEmbed bool `form:"embed"`
}

// ListParams represents URL query string to browse/list post/page without context and embed parameter
// This model is used inside service
type ListParams struct {
	ListFilter
	Categories        []uint64 `form:"categories"`
	CategoriesExclude []uint64 `form:"categories_exclude"`
	Tags              []uint64 `form:"tags"`
	TagsExclude       []uint64 `form:"tags_exclude"`
}

// ListFilter represents parameters to call QueryPosts function in repository to get list of posts that
// match the filter/parameters criteria
type ListFilter struct {
	Page                  int      `form:"page"`
	PerPage               int      `form:"per_page"`
	Search                *string  `form:"search"`
	After                 *string  `form:"after"`
	Author                []uint64 `form:"author"`
	AuthorExclude         []uint64 `form:"author_exclude"`
	Before                *string  `form:"before"`
	Exclude               []uint64 `form:"exclude"`
	Include               []uint64 `form:"include"`
	MimeType              *string  `form:"mime_type"`
	Offset                int      `form:"offset"`
	Order                 *string  `form:"order"`
	OrderBy               *string  `form:"order_by"`
	Parent                *string  `form:"parent"`
	ParentExclude         *string  `form:"parent_exclude"`
	Slug                  *string  `form:"slug"`
	Status                *string  `form:"status"`
	Sticky                *bool    `form:"sticky"`
	MenuOrder             *string  `form:"menu_order"`
	MediaType             *string  `form:"media_type"`
	Type                  string
	StickyIDs             map[int]bool
	TermTaxonomies        map[string][]*TermTaxonomy
	TermTaxonomiesExclude map[string][]*TermTaxonomy
}
