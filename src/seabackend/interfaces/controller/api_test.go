package controller

import (
	"context"
	"errors"
	"testing"

	"flamingo.me/flamingo/v3/framework/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mkrill/gosea/src/seabackend/domain/entity"
	"github.com/mkrill/gosea/src/seabackend/interfaces/controller/mocks"
)

func TestApi_ShowPostsWithUsers(t *testing.T) {

	// Initialize testcases with in and output parameters for
	// ShowPostsWithUsers(ctx context.Context, req *web.Request) web.Result
	testcases := []struct {
		name                string
		filter              string
		mockedResponsePosts []entity.Post
		mockedResponseError error
		wantedHttpStatus    uint
		wantedResponsePosts []entity.Post
	}{
		{
			name:   "Testfall 1",
			filter: "",
			mockedResponsePosts: []entity.Post{
				{
					Username:    "Moriah.Stanton",
					CompanyName: "Hoeger LLC",
					Name:        "",
					Title:       "beatae soluta recusandae",
					Body:        "dolorem quibusdam ducimus consequuntur dicta aut quo laboriosam\nvoluptatem quis enim recusandae ut sed sunt\nnostrum est odit totam\nsit error sed sunt eveniet provident qui nulla",
				},
			},
			mockedResponseError: nil,
			wantedHttpStatus:    200,
			wantedResponsePosts: []entity.Post{
				{
					Username:    "Moriah.Stanton",
					CompanyName: "Hoeger LLC",
					Name:        "",
					Title:       "beatae soluta recusandae",
					Body:        "dolorem quibusdam ducimus consequuntur dicta aut quo laboriosam\nvoluptatem quis enim recusandae ut sed sunt\nnostrum est odit totam\nsit error sed sunt eveniet provident qui nulla",
				},
			},
		},
		{
			name:                "Testfall 2",
			filter:              "",
			mockedResponsePosts: []entity.Post{},
			mockedResponseError: errors.New("arbitrary error"),
			wantedHttpStatus:    500,
			wantedResponsePosts: nil,
		},
	}

	// loop through the testcases
	for _, testcase := range testcases {
		// run testcase
		t.Run(testcase.name, func(t *testing.T) {
			// initialize mock for PostsWithUsersLoader
			mockedPwul := mocks.PostsWithUsersLoader{}

			// Initialize results depending on input parameters for mocked methods
			mockedPwul.On("RetrievePostsWithUsersFromBackend", mock.Anything, testcase.filter).
				Return(testcase.mockedResponsePosts, testcase.mockedResponseError).
				Once()

			// declare and initialize testAPIController
			testApiController := &ApiController{}

			// inject attributes to initialize the controller
			// Inject(pwul PostsWithUsersLoader, responder *web.Responder)
			// ToDo: where do I get the responder from?
			webResponder := &web.Responder{}

			testApiController.Inject(&mockedPwul, webResponder)

			// create request and session
			session := web.EmptySession()
			testRequest := web.CreateRequest(nil, session)
			testRequest.Params = map[string]string{
				"filter": testcase.filter,
			}

			// call testApiController.ShowPostsWithUsers(ctx context.Context, req *web.Request)
			// with parameters from testcase
			testcaseResult := testApiController.ShowPostsWithUsers(context.TODO(), testRequest)

			if testcase.wantedHttpStatus == 200 {
				response, ok := testcaseResult.(*web.DataResponse)
				// assure that response is of type web.DataResponse
				assert.True(t, ok)

				// assure correct response type
				assert.Equal(t, testcase.wantedHttpStatus, response.Response.Status)

				// assure that the response is correct
				assert.Equal(t, testcase.wantedResponsePosts, response.Data)
			} else {
				response, ok := testcaseResult.(*web.ServerErrorResponse)
				// assure that response is of type web.ServerErrorResponse
				assert.True(t, ok)

				// assure correct response type
				assert.Equal(t, testcase.wantedHttpStatus, response.Response.Status)
			}

		})
	}

}
