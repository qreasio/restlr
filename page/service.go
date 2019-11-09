package page

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/post"
	"github.com/qreasio/restlr/shared"
	"github.com/qreasio/restlr/term"
	"github.com/qreasio/restlr/user"
	log "github.com/sirupsen/logrus"
)

// Service handles async log of audit event
type Service interface {
	GetPage(ctx context.Context, req model.GetItemRequest) (interface{}, error)
	ListPages(ctx context.Context, params model.ListRequest) (interface{}, error)
}

// service is struct that will implement Service interface and store related repositories
type service struct {
	page   post.Repository
	term   term.Repository
	shared shared.Repository
	user   user.Repository
}

// NewService is a simple helper function to create a service instance
func NewService(postRepo post.Repository, sharedRepo shared.Repository, userRepo user.Repository) Service {
	return &service{
		page:   postRepo,
		shared: sharedRepo,
		user:   userRepo,
	}
}

func (s *service) GetPage(ctx context.Context, params model.GetItemRequest) (interface{}, error) {
	p, err := s.page.PostByID(ctx, *params.ID, "page")
	if err == sql.ErrNoRows {
		return nil, model.ErrInvalidPostID
	}
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.ID,
			"func":   "s.page.PostByID",
		}).Errorf("Failed to get post by id: %s", err)
		return nil, err
	}

	p.Meta = []map[string]string{}
	metas, err := s.shared.PostMetasByPostIDs(ctx, []uint64{p.ID})
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.ID,
			"func":   "s.shared.PostMetasByPostIDs",
		}).Errorf("Failed to get post meta by id: %s", err)
		return nil, err
	}

	p.SetFeaturedMediaID(metas)
	p.SetLinks(ctx)
	// pull required related data to construct complete post response
	postData, err := s.PullRawPostData(ctx, []uint64{p.ID}, []uint64{p.Author}, params.IsEmbed)

	if err = s.SetPredecessorVersion(ctx, p); err != nil {
		log.WithFields(log.Fields{
			"params": params.ID,
			"func":   "s.SetPredecessorVersion",
		}).Errorf("Failed to set predecessor version: %s", err)
		return nil, err
	}

	if params.IsEmbed {
		err := s.SetEmbedded(ctx, p, postData.User[p.Author])
		if err != nil {
			log.WithFields(log.Fields{
				"params": params.ID,
				"func":   "s.SetEmbedded",
			}).Errorf("Failed to set embedded: %s", err)
			return nil, err
		}
	}

	if params.Context == "embed" {
		return &model.ContentBase{Base: p.Base, SharedContent: p.SharedContent, Embedded: p.Embedded}, err
	}

	p.SetViewAttributes(metas, map[uint64]map[string][]uint64{}, map[uint64]string{})
	return p, err
}

// ListPages returns list of post data base on list posts request parameter
func (s *service) ListPages(ctx context.Context, params model.ListRequest) (interface{}, error) {
	log.WithFields(log.Fields{
		"params": params,
	}).Debug("service.ListPages")

	postIDList, err := s.page.QueryPosts(ctx, params.ListParams.ListFilter)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.ListParams.ListFilter,
			"func":   "s.term.QueryPosts",
		}).Errorf("Failed to queryPosts: %s", err)

		return nil, err
	}

	if len(postIDList) == 0 {
		return []model.Post{}, nil
	}

	posts, authorIDList, err := s.page.PostsByIDs(ctx, params.Type, postIDList)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("postIDList: %v, type: %v", postIDList, params.Type),
			"func":   "s.page.PostsByIDs",
		}).Errorf("Failed to get posts by ids: %s", err)
		return nil, err
	}
	// pull required related data to construct complete post response
	postData, err := s.PullRawPostData(ctx, postIDList, authorIDList, params.IsEmbed)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("postIDList: %v, authors :%v, is_embed: %t", postIDList, authorIDList, params.IsEmbed),
			"func":   "s.PullRawPostData",
		}).Errorf("Failed to pull raw post data: %s", err)
		return nil, err
	}

	predecessors, err := s.page.GetPredecessorVersion(ctx, postIDList)
	if err != nil {
		log.WithFields(log.Fields{
			"params": postIDList,
			"func":   "s.page.GetPredecessorVersion",
		}).Errorf("Failed to get predecessor version: %s", err)
		return nil, err
	}

	var basePosts = make([]*model.ContentBase, 0)

	for _, p := range posts {
		p.FeaturedMedia = postData.FeaturedMedia[p.ID]
		p.SetLinks(ctx)
		p.SetPredecessorVersion(model.GetBaseURL(ctx), predecessors[p.ID][0])

		// if post is embed ( _embed is on query string ) we need to pull all embedded attributes like author, term, replies, and featured media
		if params.IsEmbed {
			err = s.SetEmbedded(ctx, p, postData.User[p.Author])
			if err != nil {
				log.WithFields(log.Fields{
					"params": fmt.Sprintf("user : %v", postData.User[p.ID]),
					"func":   "s.SetPostEmbedded",
				}).Errorf("Failed to set post embedded: %s", err)
				return nil, err
			}
		}

		if params.Context != nil && *params.Context == "embed" {
			basePosts = append(basePosts, &model.ContentBase{Base: p.Base, SharedContent: p.SharedContent, Embedded: p.Embedded})
			continue
		}

	}

	if len(basePosts) > 0 {
		return basePosts, nil
	}
	return posts, nil
}

// PullRawPostData pull and store post metas, user, taxonomies term taxonomies and format for post
func (s *service) PullRawPostData(ctx context.Context, idList []uint64, authors []uint64, embed bool) (*model.RawPost, error) {
	rawPost := &model.RawPost{
		User: map[uint64]*model.UserDetail{},
	}

	usersDict, err := s.user.GetUserByIDList(ctx, authors)
	if err != nil {
		log.WithFields(log.Fields{
			"params": ctx,
			"func":   "s.GetUserByIDList",
		}).Errorf("Failed to get user by id list: %s", err)
		return nil, err
	}

	for idx, _ := range idList {
		rawPost.User[authors[idx]] = usersDict[authors[idx]]
	}

	return rawPost, nil
}

func (s *service) SetEmbedded(ctx context.Context, p *model.Post, user *model.UserDetail) error {
	config := ctx.Value(model.APICONFIGKEY).(model.APIConfig)
	p.Embedded = &model.Embedded{}

	p.Embedded.Author = s.user.UserDetailAsUserSlice(config.APIBaseURL, config.APIHost, user)
	return nil
}

func (s *service) SetPredecessorVersion(ctx context.Context, page *model.Post) (err error) {
	predecessorVersions, err := s.page.GetPredecessorVersion(ctx, []uint64{page.ID})
	href := fmt.Sprintf("%s/pages/%d/revisions/%d", model.GetBaseURL(ctx), page.ID, predecessorVersions[page.ID][0])
	versionLink := model.VersionLink{ID: predecessorVersions[page.ID][0], Href: href}
	page.Links.PredecessorVersion = append(page.Links.PredecessorVersion, versionLink)
	return err
}
