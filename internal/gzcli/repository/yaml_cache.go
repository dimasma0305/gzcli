package repository

import (
	"context"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
)

// YAMLCacheRepository implements CacheRepository using YAML-based caching
type YAMLCacheRepository struct {
	getCache    func(string, interface{}) error
	setCache    func(string, interface{}) error
	deleteCache func(string)
}

// NewYAMLCacheRepository creates a new YAML cache repository
func NewYAMLCacheRepository(
	getCache func(string, interface{}) error,
	setCache func(string, interface{}) error,
	deleteCache func(string),
) CacheRepository {
	return &YAMLCacheRepository{
		getCache:    getCache,
		setCache:    setCache,
		deleteCache: deleteCache,
	}
}

// Get retrieves a value from cache
func (r *YAMLCacheRepository) Get(ctx context.Context, key string, target interface{}) error {
	if r.getCache == nil {
		return errors.Wrap(errors.ErrConfigNotFound, "get cache function not provided")
	}
	
	if err := r.getCache(key, target); err != nil {
		return errors.Wrapf(err, "failed to get cache key: %s", key)
	}
	return nil
}

// Set stores a value in cache
func (r *YAMLCacheRepository) Set(ctx context.Context, key string, value interface{}) error {
	if r.setCache == nil {
		return errors.Wrap(errors.ErrConfigNotFound, "set cache function not provided")
	}
	
	if err := r.setCache(key, value); err != nil {
		return errors.Wrapf(err, "failed to set cache key: %s", key)
	}
	return nil
}

// Delete removes a value from cache
func (r *YAMLCacheRepository) Delete(ctx context.Context, key string) error {
	if r.deleteCache == nil {
		return errors.Wrap(errors.ErrConfigNotFound, "delete cache function not provided")
	}
	
	r.deleteCache(key)
	return nil
}

// Clear clears all cache entries
func (r *YAMLCacheRepository) Clear(ctx context.Context) error {
	// This would need to be implemented based on the cache implementation
	// For now, return an error indicating it's not implemented
	return fmt.Errorf("clear cache not implemented")
}