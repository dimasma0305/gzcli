package challenge

import (
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestIsChallengeExist(t *testing.T) {
	challenges := []gzapi.Challenge{
		{Title: "Challenge 1"},
		{Title: "Challenge 2"},
		{Title: "Challenge 3"},
	}

	tests := []struct {
		name          string
		challengeName string
		want          bool
	}{
		{
			name:          "existing challenge",
			challengeName: "Challenge 2",
			want:          true,
		},
		{
			name:          "non-existing challenge",
			challengeName: "Challenge 99",
			want:          false,
		},
		{
			name:          "empty name",
			challengeName: "",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsChallengeExist(tt.challengeName, challenges); got != tt.want {
				t.Errorf("IsChallengeExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExistInArray(t *testing.T) {
	array := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "existing value",
			value: "banana",
			want:  true,
		},
		{
			name:  "non-existing value",
			value: "orange",
			want:  false,
		},
		{
			name:  "empty value",
			value: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExistInArray(tt.value, array); got != tt.want {
				t.Errorf("IsExistInArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeChallengeData(t *testing.T) {
	tests := []struct {
		name          string
		challengeConf config.ChallengeYaml
		challengeData gzapi.Challenge
		checkFunc     func(*testing.T, *gzapi.Challenge)
	}{
		{
			name: "merge with container limits",
			challengeConf: config.ChallengeYaml{
				Name:        "Test Challenge",
				Author:      "test-author",
				Description: "Test description",
				Value:       500,
				Container: config.Container{
					MemoryLimit:  512,
					CpuCount:     2,
					StorageLimit: 256,
				},
			},
			challengeData: gzapi.Challenge{},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				if result.MemoryLimit != 512 {
					t.Errorf("Expected MemoryLimit 512, got %d", result.MemoryLimit)
				}
				if result.CpuCount != 2 {
					t.Errorf("Expected CpuCount 2, got %d", result.CpuCount)
				}
				if result.StorageLimit != 256 {
					t.Errorf("Expected StorageLimit 256, got %d", result.StorageLimit)
				}
				if result.Title != "Test Challenge" {
					t.Errorf("Expected Title 'Test Challenge', got %s", result.Title)
				}
				if result.OriginalScore != 500 {
					t.Errorf("Expected OriginalScore 500, got %d", result.OriginalScore)
				}
			},
		},
		{
			name: "merge without container limits (use defaults)",
			challengeConf: config.ChallengeYaml{
				Name:   "Default Challenge",
				Author: "test-author",
				Value:  50,
			},
			challengeData: gzapi.Challenge{},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				if result.MemoryLimit != 128 {
					t.Errorf("Expected default MemoryLimit 128, got %d", result.MemoryLimit)
				}
				if result.CpuCount != 1 {
					t.Errorf("Expected default CpuCount 1, got %d", result.CpuCount)
				}
				if result.StorageLimit != 128 {
					t.Errorf("Expected default StorageLimit 128, got %d", result.StorageLimit)
				}
			},
		},
		{
			name: "high score sets min score rate",
			challengeConf: config.ChallengeYaml{
				Name:   "High Value Challenge",
				Author: "test-author",
				Value:  1000,
			},
			challengeData: gzapi.Challenge{},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				if result.MinScoreRate != 0.10 {
					t.Errorf("Expected MinScoreRate 0.10 for score >= 100, got %f", result.MinScoreRate)
				}
			},
		},
		{
			name: "low score sets min score rate to 1",
			challengeConf: config.ChallengeYaml{
				Name:   "Low Value Challenge",
				Author: "test-author",
				Value:  50,
			},
			challengeData: gzapi.Challenge{},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				if result.MinScoreRate != 1 {
					t.Errorf("Expected MinScoreRate 1 for score < 100, got %f", result.MinScoreRate)
				}
			},
		},
		{
			name: "merge with author in content",
			challengeConf: config.ChallengeYaml{
				Name:        "Authored Challenge",
				Author:      "John Doe",
				Description: "This is a test",
				Value:       100,
			},
			challengeData: gzapi.Challenge{},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				expectedContent := "Author: **John Doe**\n\nThis is a test"
				if result.Content != expectedContent {
					t.Errorf("Expected Content %q, got %q", expectedContent, result.Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeChallengeData(&tt.challengeConf, &tt.challengeData)
			tt.checkFunc(t, result)
		})
	}
}

func TestIsConfigEdited(t *testing.T) {
	// Mock cache functions
	cacheData := make(map[string]interface{})

	//nolint:unparam // error return kept for interface consistency in test
	getCache := func(key string, v interface{}) error {
		if data, ok := cacheData[key]; ok {
			if ptr, ok := v.(*gzapi.Challenge); ok {
				if cached, ok := data.(gzapi.Challenge); ok {
					*ptr = cached
					return nil
				}
			}
		}
		return nil // Cache miss returns nil error but doesn't populate v
	}

	tests := []struct {
		name          string
		challengeConf config.ChallengeYaml
		challengeData gzapi.Challenge
		setupCache    func()
		want          bool
	}{
		{
			name: "cache miss - considered edited",
			challengeConf: config.ChallengeYaml{
				Name:     "Test",
				Category: "Web",
			},
			challengeData: gzapi.Challenge{
				Title: "Test",
			},
			setupCache: func() {
				cacheData = make(map[string]interface{})
			},
			want: true,
		},
		{
			name: "cache hit - data same",
			challengeConf: config.ChallengeYaml{
				Name:     "Test",
				Category: "Web",
			},
			challengeData: gzapi.Challenge{
				Title:    "Test",
				Category: "Web",
				Hints:    []string{},
			},
			setupCache: func() {
				cacheData = make(map[string]interface{})
				cacheData["Web/Test/challenge"] = gzapi.Challenge{
					Title:    "Test",
					Category: "Web",
					Hints:    []string{},
				}
			},
			want: false,
		},
		{
			name: "cache hit - data different",
			challengeConf: config.ChallengeYaml{
				Name:     "Test",
				Category: "Web",
			},
			challengeData: gzapi.Challenge{
				Title:    "Test",
				Category: "Web",
				Content:  "New content",
				Hints:    []string{},
			},
			setupCache: func() {
				cacheData = make(map[string]interface{})
				cacheData["Web/Test/challenge"] = gzapi.Challenge{
					Title:    "Test",
					Category: "Web",
					Content:  "Old content",
					Hints:    []string{},
				}
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			if got := IsConfigEdited(&tt.challengeConf, &tt.challengeData, getCache); got != tt.want {
				t.Errorf("IsConfigEdited() = %v, want %v", got, tt.want)
			}
		})
	}
}
