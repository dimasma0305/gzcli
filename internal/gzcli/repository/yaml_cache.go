package repository

import (
	"fmt"
)

// YAMLCacheRepository implements CacheRepository using YAML files
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
func (r *YAMLCacheRepository) Get(key string, dest interface{}) error {
	return r.getCache(key, dest)
}

// Set stores a value in cache
func (r *YAMLCacheRepository) Set(key string, value interface{}) error {
	return r.setCache(key, value)
}

// Delete removes a value from cache
func (r *YAMLCacheRepository) Delete(key string) error {
	if r.deleteCache == nil {
		return fmt.Errorf("delete cache function not provided")
	}
	return r.deleteCache(key)
}

// Exists checks if a key exists in cache
func (r *YAMLCacheRepository) Exists(key string) (bool, error) {
	var dest interface{}
	err := r.getCache(key, dest)
	return err == nil, nil
}