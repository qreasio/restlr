package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/toolbox"
	log "github.com/sirupsen/logrus"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
	"strconv"
	"strings"
)

type Repository interface {
	PostByID(ctx context.Context, postID uint64, postType string) (*model.Post, error)
	QueryPosts(ctx context.Context, listRequest model.ListFilter) ([]uint64, error)
	PostsByIDs(ctx context.Context, postType string, idList []uint64) ([]*model.Post, []uint64, error)
	ParseStickyPostID(option string) map[int]bool
	CommentsByPostIDs(commentPostIDStr []string) ([]*model.Comment, error)
	GetPredecessorVersion(ctx context.Context, idList []uint64) (map[uint64]map[int]uint64, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

// getQueryColumns returns slice of table columns
func getQueryColumns(postType string, siteURL string, permalinkStructure string, alias string) []string {
	dottedAlias := alias + "."
	permalink := `CONCAT('` +
		siteURL +
		`', REPLACE( REPLACE( REPLACE( REPLACE( REPLACE( '` +
		permalinkStructure + `', '%year%', DATE_FORMAT( ` + dottedAlias + `post_date, '%Y' ) ) ,'%monthnum%', 
			DATE_FORMAT( ` + dottedAlias + `post_date, '%m' ) ) , '%day%', 
			DATE_FORMAT( ` + dottedAlias + `post_date, '%d' ) ) , 
			'%postname%', ` + dottedAlias + `post_name ) , '%category%', ` + dottedAlias + `post_type ) ) AS permalink `

	fields := []string{
		dottedAlias + "ID",
		dottedAlias + "post_author",
		dottedAlias + "post_date",
		dottedAlias + "post_date_gmt",
		dottedAlias + "post_content",
		dottedAlias + "post_title",
		dottedAlias + "post_excerpt",
		dottedAlias + "post_status",
		dottedAlias + "comment_status",
		dottedAlias + "ping_status",
		dottedAlias + "post_password",
		dottedAlias + "post_name",
		dottedAlias + "post_modified",
		dottedAlias + "post_modified_gmt",
		dottedAlias + "guid",
		dottedAlias + "post_type",
		permalink,
	}

	switch postType {

	case "page":
		fields = append(fields, alias+"."+"menu_order")
		fields = append(fields, alias+"."+"post_parent")

	case "media":
		fields = append(fields, alias+"."+"post_mime_type")
	}

	return fields
}

// getQueryProperties return slice of struct fields/properties that will be scanned
func getQueryProperties(post *model.Post, postType string) []interface{} {
	fields := []interface{}{&post.ID, &post.Author, &post.Date, &post.DateGmt, &post.Content.Rendered, &post.Title.Rendered, &post.Excerpt.Rendered, &post.Status,
		&post.CommentStatus, &post.PingStatus, &post.Password, &post.Slug, &post.Modified, &post.ModifiedGmt, &post.GUID.Rendered, &post.Type, &post.Link}

	if postType == "page" {
		fields = append(fields, &post.MenuOrder)
		fields = append(fields, &post.Parent)
	}

	if postType == "media" {
		fields = append(fields, &post.MimeType)
	}

	return fields
}

// NewPost return new initialized post
func NewPost() model.Post {
	post := model.Post{}
	post.Content = &model.ContentRendered{}
	post.Title = &model.Rendered{}
	post.Excerpt = &model.ContentRendered{}
	post.GUID = &model.Rendered{}
	post.Meta = []map[string]string{}
	return post
}

// getPostByIDSQL return string sql to get post by id
func getPostByIDSQL(ctx context.Context, postType string) string {
	config := ctx.Value(model.APICONFIGKEY).(model.APIModel)
	tableName := config.TablePrefix + "posts"
	permalinkStructure := "/%postname%/"
	alias := "postp"

	columnsList := strings.Join(getQueryColumns(postType, config.SiteURL, permalinkStructure, alias), ",")

	sqlQuery := fmt.Sprintf(`SELECT %s `+
		` FROM %s %s`+
		` WHERE %s.ID = ?`,
		columnsList,
		tableName,
		alias,
		alias,
	)

	return sqlQuery
}

// getSQLFilterAndArgs return sql query and arguments from filter
func getSQLFilterAndArgs(tablePrefix string, params model.ListFilter) (string, []interface{}, string, string, error) {
	var args []interface{}
	sqlFilter := ""

	if params.Search != nil {
		searchSQL := fmt.Sprintf(`AND ((post_title LIKE ?) OR (post_excerpt LIKE ?) OR (post_content LIKE ?)) ` +
			`AND (post_password = '')`)
		sqlFilter += searchSQL
		searchKeyword := fmt.Sprintf("%%%s%%", *params.Search)
		args = append(args, searchKeyword)
		args = append(args, searchKeyword)
		args = append(args, searchKeyword)
	}

	if params.Before != nil && params.After == nil {
		//if before is exists but after is not
		sqlFilter += " AND (post_date < ?)"
		args = append(args, *params.Before)
	} else if params.Before == nil && params.After != nil {
		//if after param is exists
		sqlFilter += " AND (post_date > ?)"
		args = append(args, *params.After)
	} else if params.Before != nil && params.After != nil {
		//if both after and before param are exists
		sqlFilter += " AND (post_date < ? AND post_date > ?)"
		args = append(args, *params.Before)
		args = append(args, *params.After)
	}

	if len(params.Include) > 0 {
		sqlFilter += " AND ID IN (" + toolbox.UInt64SliceToCSV(params.Include) + ")"
	}

	if len(params.Exclude) > 0 {
		sqlFilter += " AND ID NOT IN (" + toolbox.UInt64SliceToCSV(params.Exclude) + ")"
	}

	if params.Slug != nil {
		slugArr := strings.Split(*params.Slug, ",")
		var slugArrWithQuote []string
		for _, slug := range slugArr {
			slugArrWithQuote = append(slugArrWithQuote, "'"+slug+"'")
		}
		slugString := strings.Join(slugArrWithQuote, ",")
		sqlFilter += " AND post_name IN (" + slugString + ")"
	}

	if len(params.Author) > 0 {
		sqlFilter += " AND post_author  IN (" + toolbox.UInt64SliceToCSV(params.Author) + ")"
	}

	if len(params.AuthorExclude) > 0 {
		sqlFilter += " AND post_author NOT IN (" + toolbox.UInt64SliceToCSV(params.AuthorExclude) + ")"
	}

	if len(params.StickyIDs) > 0 {
		var postIDList []string

		for postID, _ := range params.StickyIDs {
			postIDString := fmt.Sprintf("%v", postID)
			postIDList = append(postIDList, postIDString)
		}

		postIdCSV := strings.Join(postIDList, ",")

		if *params.Sticky {
			sqlFilter += " AND ID IN (" + postIdCSV + ")"
		}

		if !*params.Sticky {
			sqlFilter += " AND ID NOT IN (" + postIdCSV + ")"
		}
	}

	var taxonomyIDs = make([]string, 0)

	if len(params.TermTaxonomies[model.TAG_TYPE]) > 0 {
		for _, tt := range params.TermTaxonomies[model.TAG_TYPE] {
			taxonomyIDs = append(taxonomyIDs, strconv.FormatUint(tt.TermTaxonomyID, 10))
		}
	}

	if len(params.TermTaxonomies[model.CATEGORY_TYPE]) > 0 {
		for _, tt := range params.TermTaxonomies[model.CATEGORY_TYPE] {
			taxonomyIDs = append(taxonomyIDs, strconv.FormatUint(tt.TermTaxonomyID, 10))
		}
	}

	if len(taxonomyIDs) > 0 {
		sqlFilter += "AND (  term_relationship.term_taxonomy_id IN ( " + strings.Join(taxonomyIDs, ",") + " ) )"
	}

	var taxonomyIDsExclude = make([]string, 0)

	if len(params.TermTaxonomiesExclude[model.TAG_TYPE]) > 0 {
		for _, tt := range params.TermTaxonomiesExclude[model.TAG_TYPE] {
			taxonomyIDsExclude = append(taxonomyIDs, strconv.FormatUint(tt.TermTaxonomyID, 10))
		}
	}

	if len(params.TermTaxonomiesExclude[model.CATEGORY_TYPE]) > 0 {
		for _, tt := range params.TermTaxonomiesExclude[model.CATEGORY_TYPE] {
			taxonomyIDsExclude = append(taxonomyIDs, strconv.FormatUint(tt.TermTaxonomyID, 10))
		}
	}

	if len(taxonomyIDsExclude) > 0 {
		sqlFilter += "AND (  ID NOT IN ( SELECT object_id FROM " + tablePrefix + "term_relationship WHERE term_taxonomy_id IN (" + strings.Join(taxonomyIDsExclude, ",") + " ) ) )"
	}

	if params.Type == "page" {

		if params.MenuOrder != nil {
			sqlFilter += " AND menu_order = ?"
			args = append(args, *params.MenuOrder)
		}

		if params.Parent != nil {
			sqlFilter += " AND post_parent IN (" + *params.Parent + ")"
		}

		if params.ParentExclude != nil {
			sqlFilter += " AND post_parent NOT IN (" + *params.ParentExclude + ")"
		}

		args = append(args, "page")

	} else if params.Type == "attachment" {

		if params.MediaType != nil {

			var mediaTypeSQL []string
			for _, postMimeType := range model.MimeTypes {
				if strings.Contains(postMimeType, *params.MediaType) {
					mediaTypeSQL = append(mediaTypeSQL, "wpp.post_mime_type = '"+postMimeType+"'")
				}
			}

			sqlFilter += strings.Join(mediaTypeSQL, " OR ")
			sqlFilter += ") "
		}

		if params.MimeType != nil {
			sqlFilter += " AND post_mime_type = ?"
			args = append(args, *params.MimeType)
		}

		if params.Parent != nil {
			sqlFilter += " AND post_parent = ?"
			args = append(args, *params.Parent)
		}

		args = append(args, "attachment")

	} else {
		args = append(args, "post")

	}

	if params.Status != nil {
		args = append(args, *params.Status)
	} else {
		args = append(args, "publish")
	}

	orderBy := ""
	sortOrder := "desc"
	search := ""

	if params.Search != nil {
		search = *params.Search
	}

	orderFieldMap := map[string]string{"title": "post_title",
		"author":    "post_author",
		"date":      "post_date",
		"id":        "post_id",
		"modified":  "post_modified",
		"parent":    "post_parent",
		"slug":      "post_name",
		"include":   "FIELD(ID, " + toolbox.UInt64SliceToCSV(params.Include) + ")",
		"relevance": "post_title LIKE '%" + search + "%'"}

	if params.OrderBy != nil {

		if len(params.Include) == 0 && *params.OrderBy == "include" {
			return "", nil, "", "", errors.New("you need to define an include parameter to order by include")
		}

		//if order by parameter is 'include', we ignore the ascending and descending order
		if len(params.Include) > 0 && *params.OrderBy == "include" {
			sortOrder = ""
		}

		if params.Search == nil && *params.OrderBy == "relevance" {
			return "", nil, "", "", errors.New("you need to define a search term to order by relevance")
		}
		orderBy = orderFieldMap[*params.OrderBy]

	} else {
		orderBy = "post_date"
	}

	offset := (params.Page - 1) * params.PerPage
	args = append(args, offset)
	args = append(args, params.PerPage)

	return sqlFilter, args, orderBy, sortOrder, nil
}

// getSQLFilterAndArgs return sql query string and argument slice to filter posts
func getSQLQuery(ctx context.Context, params model.ListFilter) (string, []interface{}, error) {
	config := ctx.Value(model.APICONFIGKEY).(model.APIModel)
	tableName := config.TablePrefix + "posts"
	var args []interface{}
	sqlFilter, args, orderBy, sortDirection, err := getSQLFilterAndArgs(config.TablePrefix, params)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("params: %v, tablePrefix: %v", params, config.TablePrefix),
			"func":   "getSQLFilterAndArgsgetSQLQuery",
		}).Errorf("Failed to run getSQLFilterAndArgs: %s", err)
		return "", nil, err
	}

	sqlQuery := `SELECT wpp.ID ` +
		`FROM ` + tableName + ` wpp ` +
		`LEFT JOIN ` + config.TablePrefix + `term_relationships term_relationship ON (wpp.ID = term_relationship.object_id) ` +
		`WHERE 1=1 ` +
		sqlFilter +
		` AND post_type = ?` +
		` AND post_status = ?` +
		` GROUP BY wpp.ID ` +
		` ORDER BY ` + orderBy + ` ` + sortDirection +
		` LIMIT ?, ?`

	return sqlQuery, args, nil
}

// QueryPosts will query posts base on filter parameters
func (repo *repository) QueryPosts(ctx context.Context, params model.ListFilter) ([]uint64, error) {
	sqlQuery, args, err := getSQLQuery(ctx, params)
	if err != nil {
		log.WithFields(log.Fields{
			"params": params,
			"func":   "repository.getSQLQuery",
		}).Errorf("Failed to get sql query: %s", err)
		return nil, err
	}

	log.WithFields(log.Fields{
		"params": args,
		"func":   "repository.QueryPosts",
	}).Errorf("QueryPosts SQL : %s", sqlQuery)

	queryRes, err := repo.db.Query(sqlQuery, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"params": args,
			"func":   "repo.db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}

	var postID uint64
	var postIDs []uint64

	for queryRes.Next() {
		err = queryRes.Scan(&postID)
		if err != nil {
			log.WithFields(log.Fields{
				"params": postID,
				"func":   "repo.Scan",
			}).Errorf("Failed to scan db query: %s", err)
			return nil, err
		}
		postIDs = append(postIDs, postID)
	}

	return postIDs, err
}

// getPostsByIDsSQL return sql query string to get post from some post IDs in csv format (comma separated string)
func (repo *repository) getPostsByIDsSQL(ctx context.Context, postType string, postIDList string) string {
	config := ctx.Value(model.APICONFIGKEY).(model.APIModel)
	tableName := config.TablePrefix + "posts"
	permalinkStructure := "/%postname%/"
	alias := "postp"

	columnsList := strings.Join(getQueryColumns(postType, config.SiteURL, permalinkStructure, alias), ",")

	sqlQuery := fmt.Sprintf(`SELECT %s `+
		` FROM %s %s`+
		` WHERE %s.ID IN (%s)`,
		columnsList,
		tableName,
		alias,
		alias,
		postIDList,
	)
	return sqlQuery
}

// PostsByIDs get Post by array of post ID string
func (repo *repository) PostsByIDs(ctx context.Context, postType string, idList []uint64) ([]*model.Post, []uint64, error) {
	var userIDArr []uint64

	postIDList := toolbox.UInt64SliceToCSV(idList)

	sqlQuery := repo.getPostsByIDsSQL(ctx, postType, postIDList)

	q, err := repo.db.Query(sqlQuery)

	if err != nil {
		log.WithFields(log.Fields{
			"params": sqlQuery,
			"func":   "repo.db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, nil, err
	}

	var posts = make([]*model.Post, 0)

	for q.Next() {
		p := NewPost()
		if p.Type == "post" {
			p.Format = "standard"
		}
		fields := getQueryProperties(&p, postType) // get list of target fields to be scanned
		// scan
		err = q.Scan(fields...)
		if err != nil {
			log.WithFields(log.Fields{
				"params": fields,
				"func":   "q.Scan",
			}).Errorf("Failed to run db scan: %s", err)
			return nil, nil, err
		}

		if p.Excerpt.Rendered == "" {
			p.Excerpt.Rendered = model.GenerateExcerpt(p.Content.Rendered)
		}
		posts = append(posts, &p)
		userIDArr = append(userIDArr, p.Author)
	}

	return posts, userIDArr, nil
}

// PostByID retrieves a row from prefix+'_posts' as a Post.
func (repo *repository) PostByID(ctx context.Context, id uint64, postType string) (*model.Post, error) {
	var err error

	post := NewPost()
	if postType == "post" {
		post.Format = "standard"
	}

	sqlQuery := getPostByIDSQL(ctx, postType)     //generate SQL
	fields := getQueryProperties(&post, postType) // get list of target fields to be scanned
	err = repo.db.QueryRow(sqlQuery, id).Scan(fields...)

	if err == sql.ErrNoRows {
		log.WithFields(log.Fields{
			"params": id,
		}).Infof("%s", err)
		return nil, err
	}

	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"params": fields,
			"func":   "repo.QueryRow.Scan",
		}).Errorf("Failed to scan db query row scan: %s", err)
		return nil, err
	}

	// if media and mime type contains image, we set media type to image
	if postType == "media" && strings.Contains(*post.MimeType, "image") {
		post.MediaType = toolbox.StringPointer("image")
	}

	// if excerpt on database is empty, populate it from content
	if post.Excerpt.Rendered == "" {
		post.Excerpt.Rendered = model.GenerateExcerpt(post.Content.Rendered)
	}

	return &post, nil
}

// ParseStickyPostID return list of post id that have sticky option value equals to true
func (repo *repository) ParseStickyPostID(optionValue string) map[int]bool {
	var stickyPosts = make(map[int]bool)
	decoder := php_serialize.NewUnSerializer(optionValue)
	val, err := decoder.Decode()
	if err != nil {
		log.WithFields(log.Fields{
			"params": optionValue,
			"func":   "decoder.Decode",
		}).Errorf("Error while decoding %v\n", err)

	} else {
		valArr, isArray := val.(php_serialize.PhpArray)
		if isArray {
			for _, postID := range valArr {
				stickyPosts[postID.(int)] = true
			}
		}
	}

	return stickyPosts
}

// CommentsByPostIDs query comments from array of string post id
func (repo *repository) CommentsByPostIDs(commentPostIDStr []string) ([]*model.Comment, error) {
	var err error
	postIDParameters := strings.Join(commentPostIDStr, ",")

	sqlQuery := `SELECT ` +
		`comment_ID, user_id, comment_author, comment_author_url, comment_date, comment_content, comment_parent, comment_post_id ` +
		`FROM wp_comments ` +
		`WHERE comment_post_ID IN (` + postIDParameters + `) and comment_approved = '1' and comment_type in ('')`

	q, err := repo.db.Query(sqlQuery)
	if err != nil {
		log.WithFields(log.Fields{
			"params": sqlQuery,
			"func":   "db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	var res []*model.Comment
	for q.Next() {
		wc := model.Comment{}
		wc.Content = &model.ContentRendered{}
		err = q.Scan(&wc.ID, &wc.Author, &wc.AuthorName, &wc.AuthorAvatarURL, &wc.Date, &wc.Content.Rendered, &wc.Parent, &wc.PostID)
		if err != nil {
			log.WithFields(log.Fields{
				"params": commentPostIDStr,
				"func":   "q.Scan",
			}).Errorf("Failed to run db scan: %s", err)
			return nil, err
		}

		res = append(res, &wc)
	}

	return res, nil
}

// GetPredecessorVersion is function to get post ID of previous version from [prefix]posts table
func (repo *repository) GetPredecessorVersion(ctx context.Context, idList []uint64) (map[uint64]map[int]uint64, error) {
	var err error
	apiConfig := ctx.Value(model.APICONFIGKEY).(model.APIModel)
	tableName := apiConfig.TablePrefix + "posts"
	idParameters := toolbox.UInt64SliceToCSV(idList)

	sqlQuery := fmt.Sprintf(`SELECT post_parent, ID FROM %s WHERE 1=1 AND post_parent IN ( %s )  
								AND post_type = 'revision' 
								AND ((post_status = 'inherit'))  
								ORDER BY post_date DESC, wp_posts.ID DESC`,
		tableName,
		idParameters)

	q, err := repo.db.Query(sqlQuery)
	if err != nil {
		log.WithFields(log.Fields{
			"params": sqlQuery,
			"func":   "db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	// load results and put in map of parent ID and the sorted revision ID, sort from the latest first
	res := map[uint64]map[int]uint64{}
	for q.Next() {
		var parentID uint64
		var postID uint64

		err = q.Scan(&parentID, &postID)
		if err != nil {
			log.WithFields(log.Fields{
				"params": fmt.Sprintf("parentID : %d, postID: %d", parentID, postID),
				"func":   "q.Scan",
			}).Errorf("Failed to run query scan: %s", err)
			return nil, err
		}

		_, ok := res[parentID]

		if !ok {
			res[parentID] = map[int]uint64{0: postID}
		} else {
			idx := len(res[parentID])
			res[parentID][idx] = postID
		}

	}

	return res, nil
}
