package model

import (
	"fmt"
	"time"

	"github.com/qreasio/restlr/toolbox"
)

type Comment struct {
	ID              uint64           `json:"id"`
	Parent          uint64           `json:"parent"`
	Author          uint64           `json:"author"`
	AuthorName      string           `json:"author_name"`
	AuthorAvatarURL string           `json:"author_url"`
	Date            time.Time        `json:"date""`
	Content         *ContentRendered `json:"content"`
	Link            string           `json:"link"`
	Type            string           `json:"type"`
	Links           *CommentLink     `json:"_links"`
	PostID          *uint64          `json:"post_id,omitempty"`
}

//CommentLink is to represents _links in EmbedPostComment
type CommentLink struct {
	SelfLink   []map[string]string      `json:"self"`
	Collection []map[string]string      `json:"collection"`
	Author     []*EmbeddableLink        `json:"author,omitempty"`
	Up         []*EmbeddableCommentLink `json:"up"`
	InReplyTo  []*EmbeddableLink        `json:"in-reply-to,omitempty"`
	Children   []*map[string]string     `json:"children,omitempty"`
}

//EmbeddableCommentLink struct for CommentLink in EmbedPostComment that have embeddable property
type EmbeddableCommentLink struct {
	EmbeddableLink
	PostType string `json:"post_type"`
}

// CommentsAsEmbeddedComments process Comment to have attributes that can be used as Embedded Comment
func CommentsAsEmbeddedComments(baseURL string, postLink string, comments []*Comment) ([]*Comment, error) {

	//we collect all the comment ID that is specified in parent id column
	//this means the comment with the ID has child comment
	var parentCommentIDs []uint64
	for _, comment := range comments {
		if comment.Parent != 0 {
			parentCommentIDs = append(parentCommentIDs, comment.Parent)
		}
	}

	for _, comment := range comments {
		comment.Type = "comment"
		comment.Links = &CommentLink{}

		if comment.Parent != 0 {
			parentEmbeddedLink := &EmbeddableLink{Embeddable: true, Href: baseURL + "/comments/" + fmt.Sprint(comment.Parent)}
			comment.Links.InReplyTo = append(comment.Links.InReplyTo, parentEmbeddedLink)
		}

		if IsParent(comment.ID, parentCommentIDs) {
			comment.Links.Children = []*map[string]string{}
			childrenLink := baseURL + "/comments?parent=" + fmt.Sprint(comment.ID)
			comment.Links.Children = append(comment.Links.Children, &map[string]string{"href": childrenLink})
		}

		idStr := toolbox.UInt64ToStr(comment.ID)

		selfLink := baseURL + "/comments/" + idStr
		comment.Links.SelfLink = append(comment.Links.SelfLink, HrefMap(selfLink))

		collectionLink := baseURL + "/comments"
		comment.Links.Collection = append(comment.Links.Collection, HrefMap(collectionLink))

		if comment.Author != 0 {
			authorIDStr := toolbox.UInt64ToStr(comment.Author)
			authorLink := baseURL + "/users/" + authorIDStr
			authorEmbeddedLink := GetEmbeddableLink(authorLink)
			comment.Links.Author = append(comment.Links.Author, &authorEmbeddedLink)
		}

		comment.Link = postLink + "#comment-" + fmt.Sprint(comment.ID)

		postIDStr := toolbox.UInt64ToStr(*comment.PostID)
		embeddedLink := EmbeddableLink{Embeddable: true, Href: baseURL + "/posts/" + postIDStr}

		up := &EmbeddableCommentLink{EmbeddableLink: embeddedLink, PostType: "post"}
		comment.Links.Up = append(comment.Links.Up, up)

		//set post id to nil so it will not appear in json
		comment.PostID = nil
	}

	return comments, nil
}
