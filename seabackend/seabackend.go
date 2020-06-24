package seabackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	seaEndpoint     = "http://sa-bonn.ddnss.de:3000"
	defaultTimeout  = 10 * time.Second
	defaultCacheTTL = 2 * time.Minute
)

// Cache is an interface containing the caching functions Get and Set
type Cache interface {
	Get(key string, data interface{}) error
	Set(key string, data interface{}) error
}

// SeaBackend bundles all function to access external json endpoint
type SeaBackend struct {
	endpoint   string
	cache      Cache
	httpClient *http.Client
	logger *log.Logger
}

// New returns a new initialized SeaBackend struct for given endpoint
func New(endpoint string, logger *log.Logger) *SeaBackend {
	return &SeaBackend{
		endpoint: endpoint,
		cache:	NewRequestCache(defaultCacheTTL, logger),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: logger,
	}
}

// NewWithSEA returns a new initialized Posts struct pointing
// to SEA json server endpoint
func NewWithSEA(logger *log.Logger) *SeaBackend {
	return New(seaEndpoint, logger)
}

// LoadPosts loads all posts existing from p.endpoint
func (p *SeaBackend) LoadPosts(ctx context.Context) ([]RemotePost, error) {
	var remotePosts []RemotePost

	err := p.load(ctx, p.endpoint+"/posts", &remotePosts)

	if err != nil {
		return remotePosts, fmt.Errorf("could not load posts: %w", err)
	}

	return remotePosts, nil
}

// LoadUsers loads all existing users from external endpoint
func (p *SeaBackend) LoadUsers(ctx context.Context) ([]RemoteUser, error) {
	var remoteUsers []RemoteUser
	err := p.load(ctx, p.endpoint+"/users", &remoteUsers)
	if err != nil {
		return remoteUsers, fmt.Errorf("could not load users: %w", err)
	}

	return remoteUsers, nil
}

// LoadUser loads user with id from external endpoint
func (p *SeaBackend) LoadUser(ctx context.Context, id string) (RemoteUser, error) {
	var remoteUsers []RemoteUser
	var user RemoteUser

	err := p.load(ctx, p.endpoint+"/users?id="+id, &remoteUsers)

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
func (p *SeaBackend) load(ctx context.Context, requestUrl string, data interface{}) (err error) {

	// retrieve request result from cache
	err = p.cache.Get(requestUrl, data)
	// if requestUrl was found in cache
	if err == nil {
		p.logger.Printf("Retrieved response to %s from cache", requestUrl)
		return nil
	}

	p.logger.Printf("Retrieving response to %s from backend", requestUrl)

	// set timeout context to defaultTimeout (see above)
	ctxTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
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
	res, err := p.httpClient.Do(req)
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

	// refresh cache item with data just read
	err = p.cache.Set(requestUrl, data)
	if err != nil {
		return fmt.Errorf("failed to save data to cache: %w", err)
	}

	return nil

}
