package page

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/post"
	mockpost "github.com/qreasio/restlr/post/mock"
	mockshared "github.com/qreasio/restlr/shared/mock"
	mockuser "github.com/qreasio/restlr/user/mock"
	"github.com/stretchr/testify/assert"
)

var (
	ctx       = context.Background()
	apiConfig = model.APIConfig{}
)

func TestService_GetPage(t *testing.T) {
	ctx = context.WithValue(ctx, model.APIConfigKey, apiConfig)
	ctrl := gomock.NewController(t)
	postRepoMock := mockpost.NewMockRepository(ctrl)

	id := uint64(1)
	invalidID := uint64(100000)
	idList := []uint64{id}

	page1 := post.NewPost()
	page1.ID = id
	page1.Type = model.PageType
	page1.Author = id

	predecessorVersion := map[uint64]map[int]uint64{id: map[int]uint64{0: id}}

	postRepoMock.EXPECT().PostByID(ctx, id, "page").Return(&page1, nil)
	postRepoMock.EXPECT().PostByID(ctx, invalidID, "page").Return(nil, model.ErrInvalidPostID)
	postRepoMock.EXPECT().GetPredecessorVersion(ctx, idList).Return(predecessorVersion, nil)
	postRepo := postRepoMock

	thumbnailID := map[string]string{"_thumbnail_id": "10"}
	metas := map[uint64]map[string]string{id: thumbnailID}

	sharedRepoMock := mockshared.NewMockRepository(ctrl)
	sharedRepoMock.EXPECT().PostMetasByPostIDs(ctx, idList).Return(metas, nil)
	sharedRepo := sharedRepoMock

	user := model.UserDetail{User: model.User{ID: 1}}
	userMap := map[uint64]*model.UserDetail{id: &user}
	userRepoMock := mockuser.NewMockRepository(ctrl)

	userRepoMock.EXPECT().GetUserByIDList(ctx, idList).Return(userMap, nil)
	userRepo := userRepoMock

	s := NewService(postRepo, sharedRepo, userRepo)

	page := &model.Post{}
	param := model.GetItemRequest{ID: &id}
	p, err := s.GetPage(ctx, param)

	page = p.(*model.Post)

	assert.Nil(t, err)
	assert.Equal(t, page.ID, id)

	invalidParam := model.GetItemRequest{ID: &invalidID}
	_, err = s.GetPage(ctx, invalidParam)

	assert.NotNil(t, err)

}
