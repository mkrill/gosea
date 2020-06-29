package application

import (
	"context"
	"sync"

	"github.com/mkrill/gosea/src/seabackend/domain/entity"
	"github.com/mkrill/gosea/src/seabackend/domain/service"
)

type PostsWithUsers struct {
	seaBackendAdapter service.SeaBackendLoader
	workerCount       int
}

func (pwu *PostsWithUsers) Inject(seaBackendAdapter service.SeaBackendLoader,
	cfg *struct {
	WorkerCount float64 `inject:"config:api.workerCount"`
},
) *PostsWithUsers {
	pwu.seaBackendAdapter = seaBackendAdapter
	pwu.workerCount = int(cfg.WorkerCount)
	return pwu
}

func (pwu *PostsWithUsers) RetrievePostsWithUsersFromBackend(ctx context.Context, filter string) ([]entity.Post, error) {

	responsePosts := make([]entity.Post, 0)

	remotePosts, err := pwu.seaBackendAdapter.LoadPosts(ctx)
	if err != nil {
		//a.logger.Printf("error loading seabackend: %s", err)
		return responsePosts, err
	}

	// Create channel to pass remotePosts to be processed to loadUserFunc
	remotePostsChan := make(chan entity.RemotePost)
	// Create channel to pass responsePosts back from loadUserFunc
	responsePostsChan := make(chan entity.Post)

	// create function to enhance remotePosts with user data
	loadUserFunc := func(workerId int, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		for remotePost := range remotePostsChan {
			user, err := pwu.seaBackendAdapter.LoadUser(ctx, remotePost.UserID.String())
			if err != nil {
				continue
			}
			post := entity.Post{
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
	for i := 0; i < pwu.workerCount; i++ {
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
		if !remotePost.Contains(filter, entity.FieldTitle) {
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

	return responsePosts, nil

}
