package infrastructure

//go:generate mockery --all

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

// Cacher is an interface containing the caching functions Get and Set
type Cacher interface {
	Get(key string, data interface{}) error
	Set(key string, data interface{}) error
}

type (
	SeaBackendServiceAdapter struct {
		endpoint   string
		cache      Cacher
		httpClient *http.Client
	}
)

// verify interface
var _ service.SeaBackendLoader = &SeaBackendServiceAdapter{}

// Inject dependencies
func (sba *SeaBackendServiceAdapter) Inject(
	cache Cacher,
	cfg *struct {
		SeaEndpoint    string  `inject:"config:seabackend.endpoint"`
		DefaultTimeout float64 `inject:"config:seabackend.defaultTimeout"`
	},
) {
	if cfg != nil {
		sba.endpoint = cfg.SeaEndpoint
		sba.httpClient = &http.Client{
			Timeout: time.Duration(cfg.DefaultTimeout) * time.Second,
		}
	}
	sba.cache = cache
}

// LoadPosts loads all posts existing from p.Endpoint
func (sba *SeaBackendServiceAdapter) LoadPosts(ctx context.Context) ([]entity.RemotePost, error) {
	var remotePosts []entity.RemotePost

	err := sba.load(ctx, sba.endpoint+"/posts", &remotePosts)

	if err != nil {
		return remotePosts, fmt.Errorf("could not load posts: %w", err)
	}

	return remotePosts, nil
}

// LoadUsers loads all existing users from external Endpoint
func (sba *SeaBackendServiceAdapter) LoadUsers(ctx context.Context) ([]entity.RemoteUser, error) {
	var remoteUsers []entity.RemoteUser
	err := sba.load(ctx, sba.endpoint+"/users", &remoteUsers)
	if err != nil {
		return remoteUsers, fmt.Errorf("could not load users: %w", err)
	}

	return remoteUsers, nil
}

// LoadUser loads user with id from external Endpoint
func (sba *SeaBackendServiceAdapter) LoadUser(ctx context.Context, id string) (entity.RemoteUser, error) {
	var remoteUsers []entity.RemoteUser
	var user entity.RemoteUser

	err := sba.load(ctx, sba.endpoint+"/users?id="+id, &remoteUsers)

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

	// retrieve request result from Cacher
	err = sba.cache.Get(requestUrl, data)
	// if requestUrl was found in Cacher
	if err == nil {
		return nil
	}

	// set timeout context to defaultTimeout (see above)
	ctxTimeout, cancel := context.WithTimeout(ctx, sba.httpClient.Timeout)
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
	res, err := sba.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		bodyCloseErr := res.Body.Close()
		// return bodyCloseErr only, if err == nil so far, otherwise err!=nil so far might get overwritten
		if err == nil {
			err = bodyCloseErr
		}
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

	// refresh Cacher item with data just read
	err = sba.cache.Set(requestUrl, data)
	if err != nil {
		return fmt.Errorf("failed to save data to Cacher: %w", err)
	}

	return nil
}
