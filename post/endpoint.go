package post

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/qreasio/restlr/http"
	"github.com/qreasio/restlr/model"
)

func makeGetPostEndpoint(s Service) endpoint.Endpoint {
	endpoint := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.GetItemRequest)
		res, err := s.GetPost(ctx, req)
		if err == model.ErrInvalidPostID {
			return http.NewInvalidPostResponse(), nil
		}
		return res, nil
	}
	return endpoint
}

func makeListPostsEndpoint(s Service) endpoint.Endpoint {
	endpoint := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.ListRequest)
		return s.ListPosts(ctx, req)
	}
	return endpoint
}
