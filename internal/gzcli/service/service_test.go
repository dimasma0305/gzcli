package service

import (
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestContainer(t *testing.T) {
	// Create a mock config
	conf := &config.Config{
		Event: gzapi.Game{
			Id:    1,
			Title: "Test Game",
		},
	}

	// Create container
	container := NewContainer(ContainerConfig{
		Config:      conf,
		API:         nil, // Mock API
		Game:        &conf.Event,
		GetCache:    func(string, interface{}) error { return nil },
		SetCache:    func(string, interface{}) error { return nil },
		DeleteCache: func(string) error { return nil },
	})

	// Test that services are created
	challengeService := container.ChallengeService()
	if challengeService == nil {
		t.Error("Expected challenge service to be created")
	}

	attachmentService := container.AttachmentService()
	if attachmentService == nil {
		t.Error("Expected attachment service to be created")
	}

	flagService := container.FlagService()
	if flagService == nil {
		t.Error("Expected flag service to be created")
	}

	gameService := container.GameService()
	if gameService == nil {
		t.Error("Expected game service to be created")
	}

	// Test that repositories are created
	challengeRepoFromContainer := container.ChallengeRepository()
	if challengeRepoFromContainer == nil {
		t.Error("Expected challenge repository to be created")
	}

	attachmentRepoFromContainer := container.AttachmentRepository()
	if attachmentRepoFromContainer == nil {
		t.Error("Expected attachment repository to be created")
	}

	flagRepoFromContainer := container.FlagRepository()
	if flagRepoFromContainer == nil {
		t.Error("Expected flag repository to be created")
	}

	gameRepoFromContainer := container.GameRepository()
	if gameRepoFromContainer == nil {
		t.Error("Expected game repository to be created")
	}

	cacheRepoFromContainer := container.CacheRepository()
	if cacheRepoFromContainer == nil {
		t.Error("Expected cache repository to be created")
	}
}

func TestRetryHandler(t *testing.T) {
	retryHandler := NewRetryHandler(3, 0)
	
	// Test successful execution
	callCount := 0
	err := retryHandler.Execute(func() error {
		callCount++
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestErrorHandler(t *testing.T) {
	errorHandler := NewErrorHandler()
	
	// Test error wrapping
	originalErr := &testError{message: "original error"}
	wrappedErr := errorHandler.Wrap(originalErr, "context")
	
	if wrappedErr == nil {
		t.Error("Expected wrapped error to not be nil")
	}
	
	// Test nil error
	nilErr := errorHandler.Wrap(nil, "context")
	if nilErr != nil {
		t.Errorf("Expected nil error to remain nil, got %v", nilErr)
	}
}

func TestValidator(t *testing.T) {
	validator := NewValidator()
	
	// Test ValidateNotEmpty
	err := validator.ValidateNotEmpty("", "field")
	if err == nil {
		t.Error("Expected error for empty string")
	}
	
	err = validator.ValidateNotEmpty("value", "field")
	if err != nil {
		t.Errorf("Expected no error for non-empty string, got %v", err)
	}
	
	// Test ValidatePositive
	err = validator.ValidatePositive(0, "field")
	if err == nil {
		t.Error("Expected error for zero value")
	}
	
	err = validator.ValidatePositive(-1, "field")
	if err == nil {
		t.Error("Expected error for negative value")
	}
	
	err = validator.ValidatePositive(1, "field")
	if err != nil {
		t.Errorf("Expected no error for positive value, got %v", err)
	}
}

// Helper type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}