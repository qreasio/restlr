package http

import (
	"context"
	"github.com/qreasio/restlr/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncodeError(t *testing.T) {
	ctx := context.Background()
	response1 := httptest.NewRecorder()

	//should return 404 status code if error is ErrInvalidRoute
	EncodeError(ctx, model.ErrInvalidRoute, response1)
	assert.Equal(t, response1.Code, http.StatusNotFound)

	//should return 400 status code for any error other than ErrInvalidRoute
	response2 := httptest.NewRecorder()
	EncodeError(ctx, model.ErrInvalidParameter, response2)
	assert.Equal(t, response2.Code, http.StatusBadRequest)
}

type customResponse struct {
	message string
}

func (r customResponse) StatusCode() int {
	return http.StatusAccepted
}

func (r customResponse) Headers() http.Header {
	val := []string{r.message}
	headerMap := map[string][]string{"message": val}
	return headerMap
}

func TestEncodeJSONResponse(t *testing.T) {
	ctx := context.Background()
	resp1 := httptest.NewRecorder()

	//if response is APIResponse
	notFound := NewRouteNotFoundResponse()
	EncodeJSONResponse(ctx, resp1, notFound)
	assert.Equal(t, resp1.Code, http.StatusNotFound)

	//if response implements Headerer and StatusCoder interface
	resp2 := httptest.NewRecorder()
	custResp := customResponse{message: "Golang is amazing"}
	EncodeJSONResponse(ctx, resp2, custResp)

	assert.Equal(t, resp2.Code, http.StatusAccepted)
	assert.Equal(t, resp2.Header().Get("message"), "Golang is amazing")

}
