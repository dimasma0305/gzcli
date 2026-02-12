package challenge

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func resetRegistryLoginStateForTest() {
	registryLoginMu.Lock()
	defer registryLoginMu.Unlock()
	registryLoginCache = make(map[string]registryLoginState)
}

func TestDockerLoginOnce_CachesSuccess(t *testing.T) {
	resetRegistryLoginStateForTest()

	origRunDocker := runDockerCommand
	defer func() { runDockerCommand = origRunDocker }()

	var calls int32
	runDockerCommand = func(_ context.Context, _ string, _ []string, _ string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	ctx := context.Background()
	if err := dockerLoginOnce(ctx, "registry.example.com", "user", "pass"); err != nil {
		t.Fatalf("first dockerLoginOnce() failed: %v", err)
	}
	if err := dockerLoginOnce(ctx, "registry.example.com", "user", "pass"); err != nil {
		t.Fatalf("second dockerLoginOnce() failed: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("docker login calls = %d, want 1", got)
	}
}

func TestDockerLoginOnce_DoesNotCacheFailure(t *testing.T) {
	resetRegistryLoginStateForTest()

	origRunDocker := runDockerCommand
	defer func() { runDockerCommand = origRunDocker }()

	var calls int32
	runDockerCommand = func(_ context.Context, _ string, _ []string, _ string) error {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return errors.New("temporary registry failure")
		}
		return nil
	}

	ctx := context.Background()
	if err := dockerLoginOnce(ctx, "registry.example.com", "user", "pass"); err == nil {
		t.Fatal("first dockerLoginOnce() should fail")
	}
	if err := dockerLoginOnce(ctx, "registry.example.com", "user", "pass"); err != nil {
		t.Fatalf("second dockerLoginOnce() should retry and succeed, got: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("docker login calls = %d, want 2", got)
	}
}
