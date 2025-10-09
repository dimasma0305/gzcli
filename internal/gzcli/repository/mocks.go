package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// MockChallengeRepository is a mock implementation of ChallengeRepository
type MockChallengeRepository struct {
	challenges map[string]*gzapi.Challenge
	challengesByID map[int]*gzapi.Challenge
}

// NewMockChallengeRepository creates a new mock challenge repository
func NewMockChallengeRepository() *MockChallengeRepository {
	return &MockChallengeRepository{
		challenges:     make(map[string]*gzapi.Challenge),
		challengesByID: make(map[int]*gzapi.Challenge),
	}
}

// FindByName finds a challenge by its name
func (m *MockChallengeRepository) FindByName(name string) (*gzapi.Challenge, error) {
	if challenge, exists := m.challenges[name]; exists {
		return challenge, nil
	}
	return nil, fmt.Errorf("challenge not found")
}

// FindByID finds a challenge by its ID
func (m *MockChallengeRepository) FindByID(id int) (*gzapi.Challenge, error) {
	if challenge, exists := m.challengesByID[id]; exists {
		return challenge, nil
	}
	return nil, fmt.Errorf("challenge not found")
}

// List returns all challenges
func (m *MockChallengeRepository) List() ([]gzapi.Challenge, error) {
	challenges := make([]gzapi.Challenge, 0, len(m.challenges))
	for _, challenge := range m.challenges {
		challenges = append(challenges, *challenge)
	}
	return challenges, nil
}

// Create creates a new challenge
func (m *MockChallengeRepository) Create(challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	m.challenges[challenge.Title] = challenge
	m.challengesByID[challenge.Id] = challenge
	return challenge, nil
}

// Update updates an existing challenge
func (m *MockChallengeRepository) Update(challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	if _, exists := m.challenges[challenge.Title]; !exists {
		return nil, fmt.Errorf("challenge not found")
	}
	m.challenges[challenge.Title] = challenge
	m.challengesByID[challenge.Id] = challenge
	return challenge, nil
}

// Delete deletes a challenge
func (m *MockChallengeRepository) Delete(id int) error {
	challenge, exists := m.challengesByID[id]
	if !exists {
		return fmt.Errorf("challenge not found")
	}
	delete(m.challenges, challenge.Title)
	delete(m.challengesByID, id)
	return nil
}

// Exists checks if a challenge exists by name
func (m *MockChallengeRepository) Exists(name string) (bool, error) {
	_, exists := m.challenges[name]
	return exists, nil
}

// MockAttachmentRepository is a mock implementation of AttachmentRepository
type MockAttachmentRepository struct {
	attachments map[int][]gzapi.Attachment
}

// NewMockAttachmentRepository creates a new mock attachment repository
func NewMockAttachmentRepository() *MockAttachmentRepository {
	return &MockAttachmentRepository{
		attachments: make(map[int][]gzapi.Attachment),
	}
}

// Create creates a new attachment
func (m *MockAttachmentRepository) Create(challengeID int, attachment gzapi.CreateAttachmentForm) error {
	// Mock implementation - just add to the list
	attachments := m.attachments[challengeID]
	attachments = append(attachments, gzapi.Attachment{
		Id:          len(attachments) + 1,
		ChallengeId: challengeID,
		Type:        attachment.AttachmentType,
		Url:         attachment.RemoteUrl,
	})
	m.attachments[challengeID] = attachments
	return nil
}

// Delete deletes an attachment
func (m *MockAttachmentRepository) Delete(challengeID int, attachmentID int) error {
	attachments := m.attachments[challengeID]
	for i, attachment := range attachments {
		if attachment.Id == attachmentID {
			m.attachments[challengeID] = append(attachments[:i], attachments[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("attachment not found")
}

// List returns all attachments for a challenge
func (m *MockAttachmentRepository) List(challengeID int) ([]gzapi.Attachment, error) {
	return m.attachments[challengeID], nil
}

// MockFlagRepository is a mock implementation of FlagRepository
type MockFlagRepository struct {
	flags map[int][]gzapi.Flag
}

// NewMockFlagRepository creates a new mock flag repository
func NewMockFlagRepository() *MockFlagRepository {
	return &MockFlagRepository{
		flags: make(map[int][]gzapi.Flag),
	}
}

// Create creates a new flag
func (m *MockFlagRepository) Create(challengeID int, flag gzapi.CreateFlagForm) error {
	// Mock implementation - just add to the list
	flags := m.flags[challengeID]
	flags = append(flags, gzapi.Flag{
		Id:           len(flags) + 1,
		ChallengeId:  challengeID,
		Flag:         flag.Flag,
	})
	m.flags[challengeID] = flags
	return nil
}

// Delete deletes a flag
func (m *MockFlagRepository) Delete(flagID int) error {
	for challengeID, flags := range m.flags {
		for i, flag := range flags {
			if flag.Id == flagID {
				m.flags[challengeID] = append(flags[:i], flags[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("flag not found")
}

// List returns all flags for a challenge
func (m *MockFlagRepository) List(challengeID int) ([]gzapi.Flag, error) {
	return m.flags[challengeID], nil
}

// Exists checks if a flag exists
func (m *MockFlagRepository) Exists(challengeID int, flag string) (bool, error) {
	flags := m.flags[challengeID]
	for _, f := range flags {
		if f.Flag == flag {
			return true, nil
		}
	}
	return false, nil
}

// MockGameRepository is a mock implementation of GameRepository
type MockGameRepository struct {
	games     map[string]*gzapi.Game
	gamesByID map[int]*gzapi.Game
}

// NewMockGameRepository creates a new mock game repository
func NewMockGameRepository() *MockGameRepository {
	return &MockGameRepository{
		games:     make(map[string]*gzapi.Game),
		gamesByID: make(map[int]*gzapi.Game),
	}
}

// FindByTitle finds a game by its title
func (m *MockGameRepository) FindByTitle(title string) (*gzapi.Game, error) {
	if game, exists := m.games[title]; exists {
		return game, nil
	}
	return nil, fmt.Errorf("game not found")
}

// FindByID finds a game by its ID
func (m *MockGameRepository) FindByID(id int) (*gzapi.Game, error) {
	if game, exists := m.gamesByID[id]; exists {
		return game, nil
	}
	return nil, fmt.Errorf("game not found")
}

// List returns all games
func (m *MockGameRepository) List() ([]*gzapi.Game, error) {
	games := make([]*gzapi.Game, 0, len(m.games))
	for _, game := range m.games {
		games = append(games, game)
	}
	return games, nil
}

// Create creates a new game
func (m *MockGameRepository) Create(game gzapi.CreateGameForm) (*gzapi.Game, error) {
	newGame := &gzapi.Game{
		Id:    len(m.games) + 1,
		Title: game.Title,
		Start: gzapi.CustomTime{Time: game.Start},
		End:   gzapi.CustomTime{Time: game.End},
	}
	m.games[game.Title] = newGame
	m.gamesByID[newGame.Id] = newGame
	return newGame, nil
}

// Update updates an existing game
func (m *MockGameRepository) Update(game *gzapi.Game) (*gzapi.Game, error) {
	if _, exists := m.games[game.Title]; !exists {
		return nil, fmt.Errorf("game not found")
	}
	m.games[game.Title] = game
	m.gamesByID[game.Id] = game
	return game, nil
}

// Delete deletes a game
func (m *MockGameRepository) Delete(id int) error {
	game, exists := m.gamesByID[id]
	if !exists {
		return fmt.Errorf("game not found")
	}
	delete(m.games, game.Title)
	delete(m.gamesByID, id)
	return nil
}

// MockCacheRepository is a mock implementation of CacheRepository
type MockCacheRepository struct {
	cache map[string]interface{}
}

// NewMockCacheRepository creates a new mock cache repository
func NewMockCacheRepository() *MockCacheRepository {
	return &MockCacheRepository{
		cache: make(map[string]interface{}),
	}
}

// Get retrieves a value from cache
func (m *MockCacheRepository) Get(key string, dest interface{}) error {
	value, exists := m.cache[key]
	if !exists {
		return fmt.Errorf("key not found")
	}
	// In a real implementation, this would use reflection to copy the value
	// For now, just return an error as this is complex to implement properly
	_ = value // Avoid unused variable warning
	return fmt.Errorf("mock cache get not fully implemented")
}

// Set stores a value in cache
func (m *MockCacheRepository) Set(key string, value interface{}) error {
	m.cache[key] = value
	return nil
}

// Delete removes a value from cache
func (m *MockCacheRepository) Delete(key string) error {
	delete(m.cache, key)
	return nil
}

// Exists checks if a key exists in cache
func (m *MockCacheRepository) Exists(key string) (bool, error) {
	_, exists := m.cache[key]
	return exists, nil
}