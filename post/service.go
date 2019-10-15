package post

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/shared"
	"github.com/qreasio/restlr/term"
	"github.com/qreasio/restlr/toolbox"
	"github.com/qreasio/restlr/user"
	log "github.com/sirupsen/logrus"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
	"strconv"
	"strings"
)

// Service handles async log of audit event
type Service interface {
	GetPost(ctx context.Context, req model.GetItemRequest) (interface{}, error)
	ListPosts(ctx context.Context, params model.ListRequest) (interface{}, error)
}

type service struct {
	post   Repository
	term   term.Repository
	shared shared.Repository
	user   user.Repository
}

// NewService is a simple helper function to create a service instance
func NewService(postRepo Repository, termRepo term.Repository, sharedRepo shared.Repository, userRepo user.Repository) Service {
	return &service{
		post:   postRepo,
		term:   termRepo,
		shared: sharedRepo,
		user:   userRepo,
	}
}

// Get featured media id from the post meta map
func GetFeaturedMedia(postID uint64, metas map[uint64]map[string]string) uint64 {
	if featuredMedia, ok := metas[postID]["_thumbnail_id"]; ok {
		if mediaID, err := strconv.ParseUint(featuredMedia, 10, 64); err == nil {
			return mediaID
		}
	}
	return 0
}

// PullRawPostData pull and store post metas, user, taxonomies term taxonomies and format for post
func (s *service) PullRawPostData(ctx context.Context, idList []uint64, authors []uint64, embed bool) (*model.RawPost, error) {
	metas, err := s.shared.PostMetasByPostIDs(ctx, idList)
	if err != nil {
		log.WithFields(log.Fields{
			"params": idList,
			"func":   "s.shared.PostMetasByPostIDs",
		}).Errorf("Failed to post meta: %s", err)
		return nil, err
	}

	rawPost := &model.RawPost{
		Metas:         metas,
		FeaturedMedia: map[uint64]uint64{},
		User:          map[uint64]*model.UserDetail{},
	}

	rawPost.StickyPostIDs, err = s.GetStickyPostID(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"params": ctx,
			"func":   "s.GetStickyPostID",
		}).Errorf("Failed to get sticky post id: %s", err)
		return nil, err
	}

	usersDict, err := s.user.GetUserByIDList(ctx, authors)

	if err != nil {
		log.WithFields(log.Fields{
			"params": ctx,
			"func":   "s.GetUserByIDList",
		}).Errorf("Failed to get user by id list: %s", err)
		return nil, err
	}

	for idx, ID := range idList {
		postID := toolbox.UInt64ToStr(ID)

		rawPost.FeaturedMedia[ID] = GetFeaturedMedia(ID, metas)

		rawPost.User[authors[idx]] = usersDict[authors[idx]]

		if embed {
			rawPost.TermTaxonomies, rawPost.Taxonomies, rawPost.FormatMap, err = s.term.GetPostTaxonomyAndFormat(ctx, []string{postID})
			if err != nil {
				log.WithFields(log.Fields{
					"params": fmt.Sprintf("context: %v, postIDList: %v", ctx, []string{postID}),
					"func":   "s.term.GetPostTaxonomyAndFormat",
				}).Errorf("Failed to get GetPostTaxonomyAndFormat: %s", err)
				return nil, err
			}
		}
	}

	return rawPost, nil
}

// ListPosts returns list of post data base on list posts request parameter
func (s *service) ListPosts(ctx context.Context, params model.ListRequest) (interface{}, error) {
	log.WithFields(log.Fields{
		"params": params,
	}).Debug("service.ListPosts")

	//set sticky IDs
	if params.Sticky != nil {
		stickyIDs, err := s.GetStickyPostID(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"params": ctx,
				"func":   "s.GetStickyPostID",
			}).Errorf("Failed to get sticky post id: %s", err)
			return nil, err
		}
		params.ListParams.StickyIDs = stickyIDs
	}

	//set tag and category term taxonomies that will be used to filter posts
	tags, err := s.term.TermTaxonomyByTermIDListTaxonomy(params.Tags, model.TAG_TYPE)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.Tags,
			"func":   "s.term.TermTaxonomyByTermIDListTaxonomy",
		}).Errorf("Failed to get term taxonomy for tags: %s", err)
		return nil, err
	}
	categories, err := s.term.TermTaxonomyByTermIDListTaxonomy(params.Categories, model.CATEGORY_TYPE)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.Categories,
			"func":   "s.term.TermTaxonomyByTermIDListTaxonomy",
		}).Errorf("Failed to get term taxonomy for categories: %s", err)
		return nil, err
	}
	termTaxonomiesMap := map[string][]*model.TermTaxonomy{model.TAG_TYPE: tags, model.CATEGORY_TYPE: categories}
	params.ListParams.TermTaxonomies = termTaxonomiesMap

	//set excluded tag and category term taxonomies that will be used to filter posts
	tagsExclude, err := s.term.TermTaxonomyByTermIDListTaxonomy(params.TagsExclude, model.TAG_TYPE)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.TagsExclude,
			"func":   "s.term.TermTaxonomyByTermIDListTaxonomy",
		}).Errorf("Failed to get term taxonomy for tags exclude: %s", err)
		return nil, err
	}
	categoriesExclude, err := s.term.TermTaxonomyByTermIDListTaxonomy(params.CategoriesExclude, model.CATEGORY_TYPE)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.CategoriesExclude,
			"func":   "s.term.TermTaxonomyByTermIDListTaxonomy",
		}).Errorf("Failed to get term taxonomy for categories exclude: %s", err)
		return nil, err
	}
	termTaxonomiesMapExclude := map[string][]*model.TermTaxonomy{model.TAG_TYPE: tagsExclude, model.CATEGORY_TYPE: categoriesExclude}
	params.ListParams.TermTaxonomiesExclude = termTaxonomiesMapExclude
	postIDList, err := s.post.QueryPosts(ctx, params.ListParams.ListFilter)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params.ListParams.ListFilter,
			"func":   "s.term.QueryPosts",
		}).Errorf("Failed to queryPosts: %s", err)

		return nil, err
	}

	log.WithFields(log.Fields{
		"postIDList": postIDList,
	}).Debug("service.ListPosts")

	if len(postIDList) == 0 {
		return []model.Post{}, nil
	}

	posts, authorIDList, err := s.post.PostsByIDs(ctx, params.Type, postIDList)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("postIDList: %v, type: %v", postIDList, params.Type),
			"func":   "s.post.PostsByIDs",
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

	predecessors, err := s.post.GetPredecessorVersion(ctx, postIDList)
	if err != nil {
		log.WithFields(log.Fields{
			"params": postIDList,
			"func":   "s.post.GetPredecessorVersion",
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
			err = s.SetPostEmbedded(ctx, p, postData.TermTaxonomies[p.ID], postData.FormatMap, postData.User[p.Author])
			if err != nil {
				log.WithFields(log.Fields{
					"params": fmt.Sprintf("termtaxonomies :%v, formatmap: %v, user : %v", postData.TermTaxonomies[p.ID], postData.FormatMap, postData.User[p.ID]),
					"func":   "s.SetPostEmbedded",
				}).Errorf("Failed to set post embedded: %s", err)
				return nil, err
			}
		}

		if params.Context != nil && *params.Context == "embed" {
			basePosts = append(basePosts, &model.ContentBase{Base: p.Base, SharedContent: p.SharedContent, Embedded: p.Embedded})
			continue
		}

		p.SetViewAttributes(postData.Metas, postData.Taxonomies, postData.FormatMap)
		p.SetSticky(postData.StickyPostIDs)
	}

	if len(basePosts) > 0 {
		return basePosts, nil
	}
	return posts, nil
}

// GetPost returns post data base on get post request parameter
func (s *service) GetPost(ctx context.Context, params model.GetItemRequest) (interface{}, error) {
	p, err := s.post.PostByID(ctx, *params.ID, "post")
	if err == sql.ErrNoRows {
		return nil, model.ErrInvalidPostID
	}
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("ID: %d, Type: %s", params.ID, "post"),
			"func":   "s.post.PostByID",
		}).Errorf("Failed to get post by id: %s", err)
		return nil, err
	}

	// pull required related data to construct complete post response
	postData, err := s.PullRawPostData(ctx, []uint64{p.ID}, []uint64{p.Author}, params.IsEmbed)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("ID: %v, Authors: %s, IsEmbed: %t", []uint64{p.ID}, []uint64{p.Author}, params.IsEmbed),
			"func":   "s..PullRawPostData",
		}).Errorf("Failed to get raw post: %s", err)
		return nil, err
	}

	p.FeaturedMedia = postData.FeaturedMedia[p.ID]
	p.SetLinks(ctx)
	predecessorVersions, err := s.post.GetPredecessorVersion(ctx, []uint64{p.ID})
	p.SetPredecessorVersion(model.GetBaseURL(ctx), predecessorVersions[p.ID][0])

	// if post is embed ( _embed is on query string ) we need to pull all embedded attributes like author, term, replies, and featured media
	if params.IsEmbed {
		err = s.SetPostEmbedded(ctx, p, postData.TermTaxonomies[p.ID], postData.FormatMap, postData.User[p.ID])
		if err != nil {
			log.WithFields(log.Fields{
				"params": fmt.Sprintf("termtaxonomies :%v, formatmap: %v, user : %v", postData.TermTaxonomies[p.ID], postData.FormatMap, postData.User[p.ID]),
				"func":   "s.SetPostEmbedded",
			}).Errorf("Failed to set post embedded: %s", err)
			return nil, err
		}
	}

	// if context = embed, we only return core attributes of post
	if params.Context == "embed" {
		return &model.ContentBase{Base: p.Base, SharedContent: p.SharedContent, Embedded: p.Embedded}, err
	}

	p.SetViewAttributes(postData.Metas, postData.Taxonomies, postData.FormatMap)
	p.SetSticky(postData.StickyPostIDs)

	return p, err
}

// SetPostEmbedded set required attributes of post for _embed
func (s *service) SetPostEmbedded(ctx context.Context, p *model.Post, taxonomies []*model.TermWithPostTaxonomy, formatMap map[uint64]string, user *model.UserDetail) error {
	apiConfig := ctx.Value(model.APICONFIGKEY).(model.APIModel)

	p.Embedded = &model.Embedded{}
	// set author
	p.Embedded.Author = s.user.UserDetailAsUserSlice(apiConfig.APIBaseURL, apiConfig.APIHost, user)
	// set comments
	ids := toolbox.UInt64ToStrSlice(p.ID)
	comments, err := s.post.CommentsByPostIDs(ids)
	if err != nil {
		log.WithFields(log.Fields{
			"params": ids,
			"func":   "s.post.CommentsByPostIDs",
		}).Errorf("Failed to get comments by post id: %s", err)
		return err
	}
	if embeddedComments, err := model.CommentsAsEmbeddedComments(apiConfig.APIBaseURL, p.Link, comments); err == nil {
		p.Embedded.Replies = embeddedComments
	} else {
		return err
	}

	// set term
	p.Embedded.Term = s.TermPostTaxonomiesAsEmbeddedTerms(apiConfig.APIBaseURL, taxonomies)
	// set featured media
	featuredMedia, err := s.GetEmbeddedFeaturedMedia(ctx, p)
	if err != nil {
		log.WithFields(log.Fields{
			"params": p,
			"func":   "s.GetEmbeddedFeaturedMedia",
		}).Errorf("Failed to get embedded featured media: %s", err)
		return err
	}
	p.Embedded.FeaturedMedia = featuredMedia
	return nil
}

// GetStickyPostID return slice of post id that is set as sticky
func (s *service) GetStickyPostID(ctx context.Context) (map[int]bool, error) {
	stickyPostOption, err := s.shared.LoadOption(ctx, "sticky_posts")
	return s.post.ParseStickyPostID(stickyPostOption.OptionValue), err
}

func (s *service) GetEmbeddedFeaturedMedia(ctx context.Context, p *model.Post) ([]*model.BaseMedia, error) {
	if p.FeaturedMedia == 0 {
		return nil, nil
	}

	m, err := s.post.PostByID(ctx, p.FeaturedMedia, "media")
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("ID: %d, Type: %s", p.ID, "media"),
			"func":   "s.post.PostByID",
		}).Errorf("Failed to get post by id: %s", err)
		return nil, err
	}

	media := model.BaseMedia{Base: m.Base}
	media.MimeType = *m.MimeType
	media.MediaType = *m.MediaType

	mediaMetas, err := s.shared.PostMetasByPostIDs(ctx, []uint64{p.FeaturedMedia})
	if err != nil {
		log.WithFields(log.Fields{
			"params": []uint64{p.FeaturedMedia},
			"func":   "s.shared.PostMetasByPostIDs",
		}).Errorf("Failed to get post meta by post id: %s", err)
		return nil, err
	}

	altMeta, altTextOk := mediaMetas[m.ID]["_wp_attachment_image_alt"]
	if altTextOk {
		media.AltText = altMeta
	}

	metaValue, metadataOk := mediaMetas[m.ID]["_wp_attachment_metadata"]
	valueString := ""
	if metadataOk {
		if err == nil {
			valueString = metaValue
		}
	}

	apiConfig := ctx.Value(model.APICONFIGKEY).(model.APIModel)
	mediaDetail := &model.MediaDetails{}

	if metadataOk {
		decoder := php_serialize.NewUnSerializer(valueString)
		val, err := decoder.Decode()

		if err != nil {
			log.WithFields(log.Fields{
				"params": valueString,
				"func":   "decoder.Decode",
			}).Errorf("Failed to decode/unserialize php value: %s", err)
		} else {
			valArr, isArray := val.(php_serialize.PhpArray)
			if isArray {

				imageMetadata, isPhpArray := valArr[php_serialize.PhpValue("image_meta")].(php_serialize.PhpArray)
				width := valArr[php_serialize.PhpValue("width")].(php_serialize.PhpValue)
				height := valArr[php_serialize.PhpValue("height")].(php_serialize.PhpValue)
				file := valArr[php_serialize.PhpValue("file")].(php_serialize.PhpValue)

				theWidth, success := width.(int)
				if !success {
					widthString := width.(string)
					mediaDetail.Width, err = strconv.Atoi(widthString)
					if err != nil {
						log.WithFields(log.Fields{
							"params": widthString,
							"func":   " strconv.Atoi",
						}).Errorf("Error while convert width %v\n", err)
					}

				} else {
					mediaDetail.Width = theWidth
				}

				theHeight, success := height.(int)
				if !success {
					heightString := height.(string)
					mediaDetail.Height, err = strconv.Atoi(heightString)
					if err != nil {
						log.WithFields(log.Fields{
							"params": heightString,
							"func":   " strconv.Atoi",
						}).Errorf("Error while convert theHeight %v\n", err)
					}

				} else {
					mediaDetail.Height = theHeight
				}

				mediaDetail.File = file.(string)
				mediaDetail.ImageMeta = &model.ImageMeta{}

				filePathArr := strings.Split(mediaDetail.File, "/")

				if isPhpArray {

					aperture, ok := imageMetadata[php_serialize.PhpValue("aperture")].(string)
					if ok {
						mediaDetail.ImageMeta.Aperture = aperture
					}
					credit, ok := imageMetadata[php_serialize.PhpValue("credit")].(string)
					if ok {
						mediaDetail.ImageMeta.Credit = credit
					}
					camera, ok := imageMetadata[php_serialize.PhpValue("camera")].(string)
					if ok {
						mediaDetail.ImageMeta.Camera = camera
					}
					caption, ok := imageMetadata[php_serialize.PhpValue("caption")].(string)
					if ok {
						mediaDetail.ImageMeta.Caption = caption
					}
					createdTimestamp, ok := imageMetadata[php_serialize.PhpValue("created_timestamp")].(string)
					if ok {
						mediaDetail.ImageMeta.CreatedTimestamp = createdTimestamp
					}
					copyright, ok := imageMetadata[php_serialize.PhpValue("copyright")].(string)
					if ok {
						mediaDetail.ImageMeta.Copyright = copyright
					}
					focalLength, ok := imageMetadata[php_serialize.PhpValue("focal_length")].(string)
					if ok {
						mediaDetail.ImageMeta.FocalLength = focalLength
					}
					iso, ok := imageMetadata[php_serialize.PhpValue("iso")].(string)
					if ok {
						mediaDetail.ImageMeta.Iso = iso
					}
					shutterSpeed, ok := imageMetadata[php_serialize.PhpValue("shutter_speed")].(string)
					if ok {
						mediaDetail.ImageMeta.ShutterSpeed = shutterSpeed
					}
					title, ok := imageMetadata[php_serialize.PhpValue("title")].(string)
					if ok {
						mediaDetail.ImageMeta.Title = title
					}
					orientation, ok := imageMetadata[php_serialize.PhpValue("orientation")].(string)
					if ok {
						mediaDetail.ImageMeta.Orientation = orientation
					}
				}

				imageSizes, isSizeArray := valArr[php_serialize.PhpValue("sizes")].(php_serialize.PhpArray)
				imageSizeMap := map[string]*model.ImageSize{}
				if isSizeArray {
					for k := range imageSizes {
						keyString, _ := k.(string)
						sizeMapValue, sizeOk := imageSizes[k].(php_serialize.PhpArray)
						imgSize := &model.ImageSize{}
						if sizeOk {
							if height, heightOk := sizeMapValue[php_serialize.PhpValue("height")]; heightOk {

								theHeight, ok := height.(int)
								if !ok {
									heightString := height.(string)
									theHeight, err := strconv.Atoi(heightString)
									if err != nil {
										log.WithFields(log.Fields{
											"params": heightString,
											"func":   " strconv.Atoi",
										}).Errorf("Error while convert theHeight %v\n", err)
									}
									imgSize.Height = theHeight
								} else {
									imgSize.Height = theHeight
								}

							}
							if width, widthOk := sizeMapValue[php_serialize.PhpValue("height")]; widthOk {
								theWidth, ok := width.(int)
								if !ok {
									widthString := width.(string)
									theWidth, err := strconv.Atoi(widthString)
									if err != nil {
										log.WithFields(log.Fields{
											"params": widthString,
											"func":   " strconv.Atoi",
										}).Errorf("Error while convert theWidth %v\n", err)
									}
									imgSize.Height = theWidth
								} else {
									imgSize.Height = theWidth
								}

							}
							if mimeType, mimeTypeOk := sizeMapValue[php_serialize.PhpValue("mime-type")].(string); mimeTypeOk {
								imgSize.MimeType = mimeType
							}
							if file, fileOk := sizeMapValue[php_serialize.PhpValue("file")].(string); fileOk {
								imgSize.File = file
							}

						}

						sourceURL := apiConfig.SiteURL + "/" + apiConfig.UploadPath + "/" + filePathArr[0] + "/" + filePathArr[1] + "/" + imgSize.File
						imgSize.SourceURL = sourceURL
						imageSizeMap[keyString] = imgSize
					}
				}

				fullImgSize := &model.ImageSize{File: filePathArr[2], Height: mediaDetail.Height, Width: mediaDetail.Width, MimeType: media.MimeType, SourceURL: *m.GUID.Rendered}
				imageSizeMap["full"] = fullImgSize
				mediaDetail.Sizes = imageSizeMap
			}

		}
	}

	media.SourceURL = apiConfig.SiteURL + "/" + apiConfig.UploadPath + "/" + mediaDetail.File
	media.MediaDetails = mediaDetail
	return []*model.BaseMedia{&media}, nil
}

func (s *service) TermPostTaxonomiesAsEmbeddedTerms(APIBaseURL string, taxonomies []*model.TermWithPostTaxonomy) []*model.Term {
	var terms []*model.Term
	for _, t := range taxonomies {
		term := &t.Term

		if t.Taxonomy == model.CATEGORY_TYPE {
			term.Link = model.CategoryLink(APIBaseURL, t.Slug)

		} else if t.Taxonomy == model.TAG_TYPE {
			term.Link = model.TagLink(APIBaseURL, t.Slug)

		}

		id := strconv.FormatUint(t.TermID, 10)
		term.Links = model.GetTermLinks("post", APIBaseURL, id)
		terms = append(terms, term)
	}
	return terms
}
