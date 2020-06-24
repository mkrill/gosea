package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mkrill/gosea/seabackend"
)

const workerCount = 3

type seaBackendService interface {
	LoadPosts(ctx context.Context) ([]seabackend.RemotePost, error)
	LoadUser(ctx context.Context, id string) (seabackend.RemoteUser, error)
}

type Api struct {
	seaBackend seaBackendService
	logger     *log.Logger
}

func New(seaBackend seaBackendService, logger *log.Logger) *Api {
	return &Api{
		seaBackend: seaBackend,
		logger:     logger,
	}
}

// Posts returns a json response with filtered remote seabackend
func (a *Api) Posts(w http.ResponseWriter, r *http.Request) {
	var err error

	a.logger.Printf("got request %s %s", r.Method, r.URL.Path)

	// measure runtime
	start := time.Now()
	defer func() {
		a.logger.Printf("request took %s", time.Now().Sub(start))
	}()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ctxValue := context.WithValue(r.Context(), "id", 1)

	remotePosts, err := a.seaBackend.LoadPosts(ctxValue)
	if err != nil {
		a.logger.Printf("error loading seaBackend: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retrieve query parameter 'filterValue' from URL
	filterValue := r.URL.Query().Get("filter")

	responsePosts := make([]Post, 0)
	// Create channel to pass remotePosts to be processed to loadUserFunc
	remotePostsChan := make(chan seabackend.RemotePost)
	// Create channel to pass responsePosts back from loadUserFunc
	responsePostsChan := make(chan Post)

	// create function to enhance remotePosts with user data
	loadUserFunc := func(workerId int, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		for remotePost := range remotePostsChan {
			user, err := a.seaBackend.LoadUser(ctxValue, remotePost.UserID.String())
			if err != nil {
				a.logger.Printf("could not load user %s", remotePost.UserID)
				continue
			}
			post := Post{
				Title:       remotePost.Title,
				Body:        remotePost.Body,
				Username:    user.Username,
				CompanyName: user.Company.Name,
			}

			// pass back post into responsePostChan
			responsePostsChan <- post
		}
		a.logger.Printf("lodUserFunc %d stopped", workerId)
	}

	// create waitGroup wg to keep track of go routines
	wg := &sync.WaitGroup{}

	// create workerCount number of go routines processing loadUserFunc()
	for i := 0; i < workerCount; i++ {
		go loadUserFunc(i, wg)
	}

	// create a signaling channel transfering empty structs to determine, when processing of responsePosts ended
	responsePostProcessingEnded := make(chan struct{})

	// create anonymous go routine to process responsePosts passed back from loadUserFunc()
	go func() {
		for post := range responsePostsChan {
			responsePosts = append(responsePosts, post)
		}
		// put empty struct into responsePostProcessingEnded to indicate that responsePost processing ended
		responsePostProcessingEnded <- struct{}{}
		a.logger.Print("append posts stopped")
	}()

	// start processing remotePosts
	for _, remotePost := range remotePosts {
		// if current remotePost dies not match the filter, skip it
		if !remotePost.Contains(filterValue, seabackend.FieldTitle) {
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
	// wait for empty struct in channel responsePostProcessingEnded indicating that
	// the go routine processing responsePosts ended
	<-responsePostProcessingEnded

	w.Header().Set("content-type", "application/json")

	// encoder enc to convert our responsePosts slice to json
	enc := json.NewEncoder(w)
	err = enc.Encode(responsePosts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
