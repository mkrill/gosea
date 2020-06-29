package infrastructure

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	errorNotFound    = errors.New("Cache item not found")
	errorTTLExceeded = errors.New("Cache item outdated")
)

type cacheItem struct {
	data      []byte
	createdAt time.Time
}

type RequestCache struct {
	maxTTL       time.Duration
	cache        map[string]cacheItem
	protectCache sync.RWMutex
	logger       *log.Logger
}

// Inject dependencies
func (rc *RequestCache) Inject(
	cfg *struct {
	DefaultCacheTTL float64 `inject:"config:seabackend.defaultCacheTTL"`
}) *RequestCache {
	if cfg != nil {
		rc.maxTTL = time.Duration(cfg.DefaultCacheTTL) * time.Second
	}
	rc.cache = make(map[string]cacheItem)
	rc.logger = log.New(os.Stdout, "gosea", log.LstdFlags)

	return rc
}

//func NewRequestCache(ttl time.Duration, logger *log.Logger) *RequestCache {
//	return &RequestCache{
//		maxTTL: ttl,
//		cache:  make(map[string]cacheItem),
//		logger: logger,
//	}
//}

// Set writes data with key into the request Cache
func (rc *RequestCache) Set(key string, data interface{}) error {
	// convert data to []byte
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	// lock Cache
	rc.protectCache.Lock()
	defer rc.protectCache.Unlock()

	// write data into Cache
	rc.cache[key] = cacheItem{
		data:      buf.Bytes(),
		createdAt: time.Now(),
	}

	return nil
}

// Get reads data with key from request Cache into data, if TTL not exceeded
func (rc *RequestCache) Get(key string, data interface{}) error {

	// loc Cache from reading which setting new value for key
	rc.protectCache.RLock()
	defer rc.protectCache.RUnlock()

	item, found := rc.cache[key]

	// if item not found in Cache
	if !found {
		rc.logger.Printf("item for key %s not found in Cache", key)
		return errorNotFound
	}

	// if item is outdated
	if time.Now().Sub(item.createdAt) > rc.maxTTL {
		rc.logger.Printf("item for key %s in Cache, but outdated (from %v)", key, item.createdAt)
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
