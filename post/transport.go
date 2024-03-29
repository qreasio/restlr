package post

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-playground/form"
	resthttp "github.com/qreasio/restlr/http"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/toolbox"
	log "github.com/sirupsen/logrus"
)

var decoder *form.Decoder

// MakeHTTPHandler returns http handler that makes a set of endpoints available on predefined paths
func MakeHTTPHandler(s Service) http.Handler {
	r := chi.NewRouter()

	ListPostsHandler := kithttp.NewServer(
		makeListPostsEndpoint(s),
		listPostsRequestDecoder,
		resthttp.EncodeJSONResponse,
	)
	r.Method(http.MethodGet, "/", ListPostsHandler)

	GetPostHandler := kithttp.NewServer(
		makeGetPostEndpoint(s),
		getPostRequestDecoder,
		resthttp.EncodeJSONResponse,
		[]kithttp.ServerOption{
			kithttp.ServerErrorEncoder(resthttp.EncodeError),
		}...,
	)
	r.Method(http.MethodGet, "/{id}", GetPostHandler)

	return r
}

func getPostRequestDecoder(ctx context.Context, r *http.Request) (interface{}, error) {
	var getRequest model.GetItemRequest
	r.ParseForm()
	decoder = form.NewDecoder()
	err := decoder.Decode(&getRequest, r.Form)
	if err != nil {
		log.WithFields(log.Fields{
			"params": r,
			"func":   "decoder.Decode",
		}).Errorf("Failed to decode request: %s", err)
		return nil, err
	}
	id := chi.URLParam(r, "id")
	postID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"params": id,
			"func":   "strconv.ParseUint",
		}).Errorf("Failed to parse uint from string: %s", err)
		//we return err invalid route if the parameter data type is not correct because we assume t doesn't match route if id parameter is not a number
		return nil, model.ErrInvalidRoute
	}
	getRequest.ID = &postID
	if _, ok := r.URL.Query()["_embed"]; ok {
		getRequest.IsEmbed = true
	}

	return getRequest, nil
}

func listPostsRequestDecoder(ctx context.Context, r *http.Request) (interface{}, error) {
	var filter = model.ListFilter{Page: 1, PerPage: 100, Status: toolbox.StringPointer("publish"), Type: "post"}
	var params = model.ListParams{ListFilter: filter}
	var listRequest = model.ListRequest{ListParams: params}
	decoder = form.NewDecoder()
	r.ParseForm()

	err := decoder.Decode(&listRequest, r.Form)
	if err != nil {
		log.WithFields(log.Fields{
			"params": r.Form,
			"func":   "decoder.Decode",
		}).Errorf("Failed to decode request: %s", err)
		return nil, err
	}
	isEmbed := false
	if _, ok := r.URL.Query()["_embed"]; ok {
		isEmbed = true
	}
	listRequest.IsEmbed = isEmbed
	return listRequest, nil
}
