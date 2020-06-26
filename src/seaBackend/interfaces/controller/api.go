package controller

import (
	"context"
	"flamingo.me/flamingo/v3/framework/web"
	"github.com/mkrill/gosea/src/seaBackend/domain/Entity"
	"github.com/mkrill/gosea/src/seaBackend/domain/Service"
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

type (
	ApiController struct {
		seaBackend  Service.ISeaBackendService
		responder   *web.Responder
		workerCount int
	}
)

func (a *ApiController) Inject(
	sba Service.ISeaBackendService,
	responder *web.Responder,
	cfg *struct {
	workerCount float64 `inject:"config:api.workerCount"`
},
) *ApiController {
	if cfg != nil {
		a.workerCount = int(cfg.workerCount)
		a.responder = responder
		a.seaBackend = sba
	}
	return a
}

// showPostsWithUsers returns a json response with filtered remote seabackend
func (a *ApiController) ShowPostsWithUsers(ctx context.Context, req *web.Request) web.Result {
	var err error

	if req.Request().Method != http.MethodGet {
		return a.responder.ServerError(errors.Errorf("Controller method needs to be called by GET request"))
	}

	ctxValue := context.WithValue(req.Request().Context(), "id", 1)

	remotePosts, err := a.seaBackend.LoadPosts(ctxValue)
	if err != nil {
		//a.logger.Printf("error loading seaBackend: %s", err)
		return a.responder.ServerError(err)
	}

	// retrieve query parameter 'filterValue' from URL
	filterValue := req.Request().URL.Query().Get("filter")

	responsePosts := make([]Entity.Post, 0)
	// Create channel to pass remotePosts to be processed to loadUserFunc
	remotePostsChan := make(chan Entity.RemotePost)
	// Create channel to pass responsePosts back from loadUserFunc
	responsePostsChan := make(chan Entity.Post)

	// create function to enhance remotePosts with user data
	loadUserFunc := func(workerId int, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		for remotePost := range remotePostsChan {
			user, err := a.seaBackend.LoadUser(ctxValue, remotePost.UserID.String())
			if err != nil {
				continue
			}
			post := Entity.Post{
				Title:       remotePost.Title,
				Body:        remotePost.Body,
				Username:    user.Username,
				CompanyName: user.Company.Name,
			}

			// pass back post into responsePostChan
			responsePostsChan <- post
		}
		//a.logger.Printf("lodUserFunc %d stopped", workerId)
	}

	// create waitGroup wg to keep track of go routines
	wg := &sync.WaitGroup{}

	// create workerCount number of go routines processing loadUserFunc()
	for i := 0; i < a.workerCount; i++ {
		go loadUserFunc(i, wg)
	}

	// create a signaling channel transfering empty structs to determine, when processing of responsePosts ended
	responsePostProcessingEndedChan := make(chan struct{})

	// create anonymous go routine to process responsePosts passed back from loadUserFunc()
	go func() {
		for post := range responsePostsChan {
			responsePosts = append(responsePosts, post)
		}
		// put empty struct into responsePostProcessingEndedChan to indicate that responsePost processing ended
		responsePostProcessingEndedChan <- struct{}{}
		//a.logger.Print("append posts stopped")
	}()

	// start processing remotePosts
	for _, remotePost := range remotePosts {
		// if current remotePost dies not match the filter, skip it
		if !remotePost.Contains(filterValue, Entity.FieldTitle) {
			continue
		}
		// put remotePost into remotePostChan as input for loadUserFunc()
		remotePostsChan <- remotePost
	}
	// close remotePostsChan to trigger stop of for loop in loadUserFunc() go routines
	close(remotePostsChan)

	// wait for all go routines belonging to wg to end
	wg.Wait()

	// close responsePostsChan after all loadUserFunc() go routines have stopped
	close(responsePostsChan)
	// wait for empty struct in channel responsePostProcessingEndedChan indicating that
	// the go routine processing responsePosts ended
	<-responsePostProcessingEndedChan

	return a.responder.Data(responsePosts)
}
