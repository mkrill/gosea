package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mkrill/gosea/src/seabackend/domain/entity"
	"github.com/mkrill/gosea/src/seabackend/domain/service"
)

// Cache is an interface containing the caching functions Get and Set
type Cache interface {
	Get(key string, data interface{}) error
	Set(key string, data interface{}) error
}

type (
	SeaBackendServiceAdapter struct {
		Endpoint   string
		Cache      Cache
		HttpClient *http.Client
	}
)

// verify interface
var _ service.SeaBackendLoader = &SeaBackendServiceAdapter{}

// Inject dependencies
func (sba *SeaBackendServiceAdapter) Inject(
	cache *RequestCache,
	cfg *struct {
	SeaEndpoint    string  `inject:"config:seabackend.seaEndpoint"`
	DefaultTimeout float64 `inject:"config:seabackend.defaultTimeout"`
}) *SeaBackendServiceAdapter {
	if cfg != nil {
		sba.Endpoint = cfg.SeaEndpoint
		sba.Cache = cache
		sba.HttpClient = &http.Client{
			Timeout: time.Duration(cfg.DefaultTimeout) * time.Second,
		}
	}
	return sba
}

// LoadPosts loads all posts existing from p.Endpoint
func (sba *SeaBackendServiceAdapter) LoadPosts(ctx context.Context) ([]entity.RemotePost, error) {
	var remotePosts []entity.RemotePost

	err := sba.load(ctx, sba.Endpoint+"/posts", &remotePosts)

	if err != nil {
		return remotePosts, fmt.Errorf("could not load posts: %w", err)
	}

	return remotePosts, nil
}

// LoadUsers loads all existing users from external Endpoint
func (sba *SeaBackendServiceAdapter) LoadUsers(ctx context.Context) ([]entity.RemoteUser, error) {
	var remoteUsers []entity.RemoteUser
	err := sba.load(ctx, sba.Endpoint+"/users", &remoteUsers)
	if err != nil {
		return remoteUsers, fmt.Errorf("could not load users: %w", err)
	}

	return remoteUsers, nil
}

// LoadUser loads user with id from external Endpoint
func (sba *SeaBackendServiceAdapter) LoadUser(ctx context.Context, id string) (entity.RemoteUser, error) {
	var remoteUsers []entity.RemoteUser
	var user entity.RemoteUser

	err := sba.load(ctx, sba.Endpoint+"/users?id="+id, &remoteUsers)

	if err != nil {
		return user, fmt.Errorf("could not load user: %w", err)
	}

	if len(remoteUsers) <= 0 {
		return user, fmt.Errorf("could not load user for id %s", id)
	}

	user = remoteUsers[0]

	return user, nil
}

// load data via get request from requestUrl and write json response into data
func (sba *SeaBackendServiceAdapter) load(ctx context.Context, requestUrl string, data interface{}) (err error) {

	// retrieve request result from Cache
	err = sba.Cache.Get(requestUrl, data)
	// if requestUrl was found in Cache
	if err == nil {
		return nil
	}

	// set timeout context to defaultTimeout (see above)
	ctxTimeout, cancel := context.WithTimeout(ctx, sba.HttpClient.Timeout)
	defer cancel()

	// Create a get request to backend with requestUrl
	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, requestUrl, nil)
	if err != nil {
		// %w for wrapping of the error
		return fmt.Errorf("failed to create request: %w", err)
	}
	// set accept-encoding header attribute to json
	req.Header.Set("accept-encoding", "application/json")

	// execute request
	res, err := sba.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		err = res.Body.Close()
	}()

	if res.StatusCode >= 400 {
		return fmt.Errorf("remote server returned emulatedResponseStatus %d", res.StatusCode)
	}

	// Read body into byte array
	respData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to load body: %w", err)
	}

	err = json.Unmarshal(respData, data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal body: %w", err)
	}

	// refresh Cache item with data just read
	err = sba.Cache.Set(requestUrl, data)
	if err != nil {
		return fmt.Errorf("failed to save data to Cache: %w", err)
	}

	return nil
}
