package http

import (
	"context"
	"encoding/json"
	httpkit "github.com/go-kit/kit/transport/http"
	"net/http"
	"github.com/qreasio/restlr/model"
)

const (
	RestInvalidParamCode     = "rest_invalid_param"
	RestNoRouteCode          = "rest_no_route"
	RestInvalidIDCode        = "rest_post_invalid_id"
	NoRouteMessage           = "No route was found matching the URL and request method"
	RestInvalidPostIDMessage = "Invalid post ID"
)

type HttpAPIResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    ResponseData `json:"data"`
}

type ResponseData struct {
	Status int               `json:"status"`
	Params map[string]string `json:"replies,omitempty"`
}

func NewRouteNotFoundResponse() HttpAPIResponse {
	return HttpAPIResponse{
		Code:    RestNoRouteCode,
		Message: NoRouteMessage,
		Data: ResponseData{
			Status: http.StatusNotFound,
		},
	}
}

func NewInvalidPostResponse() HttpAPIResponse {
	return HttpAPIResponse{
		Code:    RestInvalidIDCode,
		Message: RestInvalidPostIDMessage,
		Data: ResponseData{
			Status: http.StatusNotFound,
		},
	}
}

func NewInvalidParam(invalidParameter string, invalidMessage string) HttpAPIResponse {
	response := HttpAPIResponse{
		Code:    RestInvalidParamCode,
		Message: "Invalid parameter(s): " + invalidParameter,
		Data: ResponseData{
			Status: http.StatusBadRequest,
		},
	}

	if invalidParameter != "" {
		response.Data.Params = map[string]string{ invalidParameter: invalidMessage }
	}
	return response
}

/**
404
{
  "code": "rest_post_invalid_id",
  "message": "Invalid post ID.",
  "data": {
    "status": 404
  }
}

404
{
  "code": "rest_no_route",
  "message": "No route was found matching the URL and request method",
  "data": {
    "status": 404
  }
}

400
{
  "code": "rest_invalid_param",
  "message": "Invalid parameter(s): include",
  "data": {
    "status": 400,
    "params": {
      "include": "include[0] is not of type integer."
    }
  }
}
*/
// EncodeJSONResponse is a Custom EncodeResponseFunc that serializes the response as a
// JSON object to the ResponseWriter using json library package to serialize json.
// Many JSON-over-HTTP services can use it as
// a sensible default. If the response implements Headerer, the provided headers
// will be applied to the response. If the response implements StatusCoder, the
// provided StatusCode will be used instead of 200.
func EncodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if header, ok := response.(httpkit.Headerer); ok {
		for k, values := range header.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	code := http.StatusOK
	if sc, ok := response.(httpkit.StatusCoder); ok {
		code = sc.StatusCode()
	}
	if resp, ok := response.(HttpAPIResponse); ok {
		code = resp.Data.Status
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}

	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err == model.ErrInvalidRoute {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(NewRouteNotFoundResponse())
	}else{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NewInvalidParam("",""))
	}



}
