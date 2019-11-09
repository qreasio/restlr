package page

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/qreasio/restlr/http"
	"github.com/qreasio/restlr/model"
)

func makeGetPageEndpoint(s Service) endpoint.Endpoint {
	endpoint := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.GetItemRequest)

		res, err := s.GetPage(ctx, req)
		if err == model.ErrInvalidPostID {
			return http.NewInvalidPostResponse(), nil
		}
		return res, nil
	}
	return endpoint
}

func makeListPagesEndpoint(s Service) endpoint.Endpoint {
	endpoint := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.ListRequest)
		return s.ListPages(ctx, req)
	}
	return endpoint
}
