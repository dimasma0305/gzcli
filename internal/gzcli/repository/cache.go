package repository

import (
	"fmt"
)

// YAMLCacheRepository implements CacheRepository using YAML cache
type YAMLCacheRepository struct {
	getCache    func(string, interface{}) error
	setCache    func(string, interface{}) error
	deleteCache func(string) error
}

// NewYAMLCacheRepository creates a new YAML cache repository
func NewYAMLCacheRepository(
	getCache func(string, interface{}) error,
	setCache func(string, interface{}) error,
	deleteCache func(string) error,
) *YAMLCacheRepository {
	return &YAMLCacheRepository{
		getCache:    getCache,
		setCache:    setCache,
		deleteCache: deleteCache,
	}
}

// Get retrieves a value from cache
func (r *YAMLCacheRepository) Get(key string, value interface{}) error {
	if r.getCache == nil {
		return fmt.Errorf("getCache function not provided")
	}
	return r.getCache(key, value)
}

// Set stores a value in cache
func (r *YAMLCacheRepository) Set(key string, value interface{}) error {
	if r.setCache == nil {
		return fmt.Errorf("setCache function not provided")
	}
	return r.setCache(key, value)
}

// Delete removes a value from cache
func (r *YAMLCacheRepository) Delete(key string) error {
	if r.deleteCache == nil {
		return fmt.Errorf("deleteCache function not provided")
	}
	return r.deleteCache(key)
}