package posts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	seaEndpoint    = "http://sa-bonn.ddnss.de:3000"
	defaultTimeout = 10 * time.Second
)

// Posts bundles all function to access an endpoint
type Posts struct {
	endpoint   string
	httpClient *http.Client
}

// New() returns a new initialized Post struct for a given endpoint
func New(endpoint string) *Posts {
	return &Posts{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// NewWithSea returns a new intialized Posts struct pointing to SEA json server endpoint
func NewWithSEA() *Posts {
	return New(seaEndpoint)
}

// LoadPosts loads all existing posts from external endpoint
func (p *Posts) LoadPosts() ([]RemotePost, error) {
	var remotePosts []RemotePost
	var err error

	// alternatively http.Get(), but then http header cannot be adjusted
	req, err := http.NewRequest(http.MethodGet, p.endpoint+"/posts", nil)
	if err != nil {
		// %w for wrapping of the error
		return remotePosts, fmt.Errorf("failed to create request: %w", err)
	}
	// here, bearer tokens could also be set
	req.Header.Set("accept-encoding", "application/json")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return remotePosts, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		return remotePosts, fmt.Errorf("remote server returned status %d", res.StatusCode)
	}

	// Read body into byte array
	respData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return remotePosts, fmt.Errorf("failed to load body: %w", err)
	}

	err = json.Unmarshal(respData, &remotePosts)
	if err != nil {
		return remotePosts, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	return remotePosts, nil

}
