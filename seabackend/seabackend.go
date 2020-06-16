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
	seaEndpoint    = "http://sa-bonn.ddnss.de:3000"
	defaultTimeout = 10 * time.Second
)

// SeaBackend bundles all function to access external json endpoint
type SeaBackend struct {
	endpoint   string
	httpClient *http.Client
}

// New returns a new initialized Posts struct for given endpoint
func New(endpoint string) *SeaBackend {
	return &SeaBackend{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// NewWithSEA returns a new initialized Posts struct pointing
// to SEA json server endpoint
func NewWithSEA() *SeaBackend {
	return New(seaEndpoint)
}

// LoadPosts loads all posts existing from p.endpoint
func (p *SeaBackend) LoadPosts(ctx context.Context) ([]RemotePost, error) {
	var RemotePosts []RemotePost

	err := p.load(ctx, p.endpoint+"/posts", &RemotePosts)

	if err != nil {
		return RemotePosts, err
	}

	return RemotePosts, nil
}

// LoadUsers loads all users existing from p.endpoint
func (p *SeaBackend) LoadUsers(ctx context.Context) ([]RemoteUser, error) {
	var RemoteUsers []RemoteUser

	err := p.load(ctx, p.endpoint+"/users", &RemoteUsers)

	if err != nil {
		return RemoteUsers, err
	}

	return RemoteUsers, nil
}

// load data via get request from requestUrl and write json response into data
func (p *SeaBackend) load(ctx context.Context, requestUrl string, data interface{}) error {

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
		err := res.Body.Close()
		if err != nil {
			// ToDo How do I get access to the logger from main?
			log.Fatal(err)
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

	return nil

}
