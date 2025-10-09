package service

import (
	"context"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
)

// mockChallengeRepository is a mock implementation of ChallengeRepository
type mockChallengeRepository struct {
	challenges []gzapi.Challenge
	err        error
}

func (m *mockChallengeRepository) GetChallenges(ctx context.Context, gameID int) ([]gzapi.Challenge, error) {
	return m.challenges, m.err
}

func (m *mockChallengeRepository) GetChallenge(ctx context.Context, gameID int, challengeID int) (*gzapi.Challenge, error) {
	for _, c := range m.challenges {
		if c.Id == challengeID {
			return &c, nil
		}
	}
	return nil, m.err
}

func (m *mockChallengeRepository) CreateChallenge(ctx context.Context, gameID int, challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	return challenge, m.err
}

func (m *mockChallengeRepository) UpdateChallenge(ctx context.Context, gameID int, challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	return challenge, m.err
}

func (m *mockChallengeRepository) DeleteChallenge(ctx context.Context, gameID int, challengeID int) error {
	return m.err
}

// mockCacheRepository is a mock implementation of CacheRepository
type mockCacheRepository struct {
	data map[string]interface{}
	err  error
}

func (m *mockCacheRepository) Get(ctx context.Context, key string, target interface{}) error {
	if m.err != nil {
		return m.err
	}
	if val, ok := m.data[key]; ok {
		// In a real implementation, you'd use reflection or a proper serialization
		// For this test, we'll just return the value as-is
		return nil
	}
	return m.err
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}) error {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	return m.err
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return m.err
}

func (m *mockCacheRepository) Clear(ctx context.Context) error {
	m.data = make(map[string]interface{})
	return m.err
}

func TestChallengeService_Sync_NewChallenge(t *testing.T) {
	// Setup
	ctx := context.Background()
	challengeRepo := &mockChallengeRepository{
		challenges: []gzapi.Challenge{},
	}
	cacheRepo := &mockCacheRepository{
		data: make(map[string]interface{}),
	}

	service := NewChallengeService(ChallengeServiceConfig{
		ChallengeRepo:  challengeRepo,
		Cache:          cacheRepo,
		AttachmentRepo: nil,
		FlagRepo:       nil,
		API:            &gzapi.GZAPI{},
		GameID:         1,
	})

	challengeConf := config.ChallengeYaml{
		Name:        "Test Challenge",
		Category:    "test",
		Description: "A test challenge",
		Type:        "StaticAttachment",
		Value:       100,
		Flags:       []string{"FLAG{test}"},
	}

	// Execute
	err := service.Sync(ctx, challengeConf)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestChallengeService_Sync_ExistingChallenge(t *testing.T) {
	// Setup
	ctx := context.Background()
	existingChallenge := gzapi.Challenge{
		Id:          1,
		Title:       "Test Challenge",
		Category:    "test",
		Description: "Old description",
		Type:        "StaticAttachment",
		Value:       50,
	}

	challengeRepo := &mockChallengeRepository{
		challenges: []gzapi.Challenge{existingChallenge},
	}
	cacheRepo := &mockCacheRepository{
		data: make(map[string]interface{}),
	}

	service := NewChallengeService(ChallengeServiceConfig{
		ChallengeRepo:  challengeRepo,
		Cache:          cacheRepo,
		AttachmentRepo: nil,
		FlagRepo:       nil,
		API:            &gzapi.GZAPI{},
		GameID:         1,
	})

	challengeConf := config.ChallengeYaml{
		Name:        "Test Challenge",
		Category:    "test",
		Description: "New description",
		Type:        "StaticAttachment",
		Value:       100,
		Flags:       []string{"FLAG{test}"},
	}

	// Execute
	err := service.Sync(ctx, challengeConf)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestChallengeService_FindChallengeByName(t *testing.T) {
	// Setup
	service := &ChallengeService{}
	challenges := []gzapi.Challenge{
		{Id: 1, Title: "Challenge 1"},
		{Id: 2, Title: "Challenge 2"},
		{Id: 3, Title: "Challenge 3"},
	}

	// Test cases
	tests := []struct {
		name           string
		challengeName  string
		expectedID     int
		expectedFound  bool
	}{
		{
			name:          "find existing challenge",
			challengeName: "Challenge 2",
			expectedID:    2,
			expectedFound: true,
		},
		{
			name:          "challenge not found",
			challengeName: "Non-existent",
			expectedID:    0,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.findChallengeByName(challenges, tt.challengeName)
			
			if tt.expectedFound {
				if result == nil {
					t.Errorf("Expected to find challenge, got nil")
				} else if result.Id != tt.expectedID {
					t.Errorf("Expected challenge ID %d, got %d", tt.expectedID, result.Id)
				}
			} else {
				if result != nil {
					t.Errorf("Expected nil, got challenge with ID %d", result.Id)
				}
			}
		})
	}
}