package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mkrill/gosea/src/seabackend/domain/entity"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// postsMock simulates a result from LoadPosts()
type postsMock struct {
	remotePosts []entity.RemotePost
	err         error
}

func (pm *postsMock) LoadPosts(ctx context.Context) ([]entity.RemotePost, error) {
	return pm.remotePosts, pm.err
}

func (pm *postsMock) LoadUser(ctx context.Context, id string) (entity.RemoteUser, error) {
	// ToDo Returning empty user???
	return entity.RemoteUser{}, nil
}

func TestApi_Posts(t *testing.T) {
	logBuf := &bytes.Buffer{}

	testApi := &ApiController{
		seaBackend: &postsMock{
			remotePosts: []entity.RemotePost{
				{
					UserID: json.Number("1"),
					ID:     json.Number("1"),
					Title:  "Title1",
					Body:   "Body1",
				},
				{
					UserID: json.Number("2"),
					ID:     json.Number("2"),
					Title:  "Title2",
					Body:   "Body2",
				},
			},
			err: nil,
		},
		logger: log.New(logBuf, "test", log.LstdFlags),
	}

	r := httptest.NewRequest(http.MethodGet, "http://localhost/seabackend", nil)
	w := httptest.NewRecorder()

	testApi.showPostsWithUsers(w, r)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("content-type"))

	var responsePosts []entity.Post
	err := json.NewDecoder(w.Body).Decode(&responsePosts)
	if err != nil {
		log.Fatal(err)
	}

	expected := []entity.Post{
		{
			Title: "Title1",
			Body:  "Body1",
		},
		{
			Title: "Title2",
			Body:  "Body2",
		},
	}

	assert.Equal(t, expected, responsePosts)
}
