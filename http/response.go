package http

import (
	"context"
	"encoding/json"
	"net/http"

	httpkit "github.com/go-kit/kit/transport/http"
	"github.com/qreasio/restlr/model"
)

const (
	// RestInvalidParamCode is string response code for invalid parameter error
	RestInvalidParamCode = "rest_invalid_param"
	// RestNoRouteCode is string response code for invalid route (404) error
	RestNoRouteCode = "rest_no_route"
	// RestInvalidIDCode is string response code for invalid id (404) if post/item not found
	RestInvalidIDCode = "rest_post_invalid_id"
	// NoRouteMessage is json response message for no route error
	NoRouteMessage = "No route was found matching the URL and request method"
	// RestInvalidPostIDMessage is json response message for invalid post id
	RestInvalidPostIDMessage = "Invalid post ID"
)

// APIResponse represent api response mainly on non 200 http status response
type APIResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    ResponseData `json:"data"`
}

// ResponseData is child struct of APIResponse for Data field
type ResponseData struct {
	Status int               `json:"status"`
	Params map[string]string `json:"replies,omitempty"`
}

// NewRouteNotFoundResponse is used to generate custom route not found api response
func NewRouteNotFoundResponse() APIResponse {
	return APIResponse{
		Code:    RestNoRouteCode,
		Message: NoRouteMessage,
		Data: ResponseData{
			Status: http.StatusNotFound,
		},
	}
}

// NewInvalidPostResponse is used to generate invalid post api response
func NewInvalidPostResponse() APIResponse {
	return APIResponse{
		Code:    RestInvalidIDCode,
		Message: RestInvalidPostIDMessage,
		Data: ResponseData{
			Status: http.StatusNotFound,
		},
	}
}

// NewInvalidParam is used to generate custom invalid parameter api response
func NewInvalidParam(invalidParameter string, invalidMessage string) APIResponse {
	response := APIResponse{
		Code:    RestInvalidParamCode,
		Message: "Invalid parameter(s): " + invalidParameter,
		Data: ResponseData{
			Status: http.StatusBadRequest,
		},
	}

	if invalidParameter != "" {
		response.Data.Params = map[string]string{invalidParameter: invalidMessage}
	}
	return response
}

/**
Example HTTP Error Response
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
	if resp, ok := response.(APIResponse); ok {
		code = resp.Data.Status
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}

	return json.NewEncoder(w).Encode(response)
}

// EncodeError to generate custom api response on decode error in transport layer
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err == model.ErrInvalidRoute {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(NewRouteNotFoundResponse())
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NewInvalidParam("", ""))
	}

}
