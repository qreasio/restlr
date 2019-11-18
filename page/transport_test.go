package page

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	resthttp "github.com/qreasio/restlr/http"
	"github.com/qreasio/restlr/model"
	"github.com/stretchr/testify/assert"
)

func TestGetPageHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := NewMockService(ctrl)
	handler := MakeHTTPHandler(s)
	r := chi.NewRouter()
	r.Mount("/pages", handler)

	srv := httptest.NewServer(r)

	method := "GET"
	url := srv.URL + "/pages/1"
	var err error
	// test for valid post id
	req, _ := http.NewRequest(method, url, nil)
	resp, _ := http.DefaultClient.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)
	post := &model.Post{}
	err = json.Unmarshal(body, post)

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	assert.Equal(t, post.ID, uint64(1))

	// test for invalid post id
	url2 := srv.URL + "/pages/99999"
	req2, _ := http.NewRequest(method, url2, nil)
	resp2, _ := http.DefaultClient.Do(req2)

	body2, _ := ioutil.ReadAll(resp2.Body)
	post2 := &resthttp.APIResponse{}
	err2 := json.Unmarshal(body2, post2)

	assert.Nil(t, err2)
	assert.Equal(t, resp2.StatusCode, http.StatusNotFound)
	assert.Equal(t, post2.Code, resthttp.RestInvalidIDCode)
}
