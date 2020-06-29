package infrastructure

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
	"time"

	"flamingo.me/flamingo/v3/framework/flamingo"
)

var (
	errorNotFound    = errors.New("Cacher item not found")
	errorTTLExceeded = errors.New("Cacher item outdated")
)

type cacheItem struct {
	data      []byte
	createdAt time.Time
}

type RequestCache struct {
	maxTTL       time.Duration
	cache        map[string]cacheItem
	protectCache sync.RWMutex
	logger       flamingo.Logger
}

// Inject dependencies
func (rc *RequestCache) Inject(
	logger flamingo.Logger,
	cfg *struct {
		DefaultCacheTTL float64 `inject:"config:seabackend.defaultCacheTTL"`
	},
) {
	if cfg != nil {
		rc.maxTTL = time.Duration(cfg.DefaultCacheTTL) * time.Second
	}
	rc.cache = make(map[string]cacheItem)
	rc.logger = logger

}

// Set writes data with key into the request Cacher
func (rc *RequestCache) Set(key string, data interface{}) error {
	// convert data to []byte
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	// lock Cacher
	rc.protectCache.Lock()
	defer rc.protectCache.Unlock()

	// write data into Cacher
	rc.cache[key] = cacheItem{
		data:      buf.Bytes(),
		createdAt: time.Now(),
	}

	return nil
}

// Get reads data with key from request Cacher into data, if TTL not exceeded
func (rc *RequestCache) Get(key string, data interface{}) error {

	// loc Cacher from reading which setting new value for key
	rc.protectCache.RLock()
	defer rc.protectCache.RUnlock()

	item, found := rc.cache[key]

	// if item not found in Cacher
	if !found {
		rc.logger.Warn(fmt.Sprintf("item for key %s not found in Cacher", key))
		return errorNotFound
	}

	// if item is outdated
	if time.Now().Sub(item.createdAt) > rc.maxTTL {
		rc.logger.Warn(fmt.Sprintf("item for key %s in Cacher, but outdated (from %v)", key, item.createdAt))
		return errorTTLExceeded
	}

	// decode cacheItem and store it in data
	buf := bytes.NewBuffer(item.data)
	enc := gob.NewDecoder(buf)
	err := enc.Decode(data)
	if err != nil {
		return fmt.Errorf("could not decode data: %w", err)
	}

	return nil
}
