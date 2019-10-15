package model

//TermTaxonomy struct represents table term_taxonomy
type TermTaxonomy struct {
	TermTaxonomyID uint64 `json:"term_taxonomy_id"` // term_taxonomy_id
	TermID         uint64 `json:"id"`               // term_id
	Taxonomy       string `json:"taxonomy"`         // taxonomy
	Description    string `json:"description"`      // description
	Parent         uint64 `json:"parent"`           // parent
	Count          int64  `json:"count"`            // count
}

// Term struct represents term table
type Term struct {
	TermID   uint64    `json:"id"`       // term_id
	Link     string    `json:"link"`     // name
	Name     string    `json:"name"`     // name
	Slug     string    `json:"slug"`     // slug
	Taxonomy string    `json:"taxonomy"` // taxonomy
	Links    *TermLink `json:"_links"`   // name
}

type TermTaxonomyJoin struct {
	Term
	TermGroup      int64  `json:"term_group"` // term_group
	TermTaxonomyID uint64 `json:"term_taxonomy_id"`
	Description    string `json:"description"` // description
	Parent         uint64 `json:"parent"`      // parent
	Count          int64  `json:"count"`       // count
}

type TermWithPostTaxonomy struct {
	TermTaxonomyJoin
	ObjectID uint64 `json:"object_id"` // object_id
}

// Curie represents _curie element as child of _link element of RestLink struct
type Curie struct {
	Name      string `json:"name"`
	Href      string `json:"href"`
	Templated bool   `json:"templated"`
}

type TermLink struct {
	SelfLink   []map[string]string `json:"self"`
	Collection []map[string]string `json:"collection"`
	About      []map[string]string `json:"about"`
	PostType   []map[string]string `json:"wp:post_type"`
	Curies     []*Curie            `json:"curies"`
}

// TermPost represents specific post taxonomy term
type TermPost struct {
	Taxonomy   string `json:"taxonomy, omitempty"`
	Embeddable bool   `json:"embeddable, omitempty"`
	Href       string `json:"href, omitempty"`
}
