package challenge

import (
	"errors"
	"sort"
	"strings"
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

func TestRemoveDuplicateChallenges(t *testing.T) {
	challenges := []gzapi.Challenge{
		{Id: 2, Title: "web/xss"},
		{Id: 1, Title: "web/xss"}, // lower ID should be kept
		{Id: 3, Title: "pwn/rop"},
	}

	var deletedIDs []int
	deduped, err := RemoveDuplicateChallenges(challenges, func(c *gzapi.Challenge) error {
		deletedIDs = append(deletedIDs, c.Id)
		return nil
	})

	if err != nil {
		t.Fatalf("RemoveDuplicateChallenges returned error: %v", err)
	}

	if len(deduped) != 2 {
		t.Fatalf("expected 2 challenges after dedupe, got %d", len(deduped))
	}

	sort.Ints(deletedIDs)
	if len(deletedIDs) != 1 || deletedIDs[0] != 2 {
		t.Fatalf("expected duplicate id 2 to be deleted, got %v", deletedIDs)
	}

	kept := make(map[int]string)
	for _, c := range deduped {
		kept[c.Id] = c.Title
	}
	if kept[1] != "web/xss" || kept[3] != "pwn/rop" {
		t.Fatalf("unexpected deduped challenges: %+v", kept)
	}
}

func TestRemoveDuplicateChallenges_PropagatesDeleteError(t *testing.T) {
	challenges := []gzapi.Challenge{
		{Id: 2, Title: "crypto/block"},
		{Id: 1, Title: "crypto/block"},
	}

	_, err := RemoveDuplicateChallenges(challenges, func(c *gzapi.Challenge) error {
		if c.Id == 2 {
			return errors.New("delete failed")
		}
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "delete failed") {
		t.Fatalf("expected delete error to propagate, got %v", err)
	}
}

func TestHandleExistingChallengeSetsGameId(t *testing.T) {
	conf := &config.Config{
		Event: gzapi.Game{Id: 42},
	}
	// cache returns challenge without GameId/CS
	getCache := func(key string, v interface{}) error {
		ptr, ok := v.(**gzapi.Challenge)
		if !ok {
			t.Fatalf("unexpected cache type")
		}
		*ptr = &gzapi.Challenge{Id: 7, Title: "web/xss"}
		return nil
	}

	challengeConf := config.ChallengeYaml{Name: "web/xss", Category: "web"}
	result, err := handleExistingChallenge(conf, challengeConf, &gzapi.GZAPI{}, getCache)
	if err != nil {
		t.Fatalf("handleExistingChallenge returned error: %v", err)
	}

	if result.GameId != 42 {
		t.Fatalf("expected GameId to be set to 42, got %d", result.GameId)
	}
	if result.CS == nil {
		t.Fatalf("expected CS to be set")
	}
}

func TestDetermineSyncPathPrefersRemoteChallengeOverCache(t *testing.T) {
	conf := &config.Config{
		Event: gzapi.Game{Id: 10},
	}
	challengeConf := config.ChallengeYaml{
		Name:     "web/xss",
		Category: "web",
	}
	remoteChallenges := []gzapi.Challenge{
		{Id: 2, Title: "web/xss", GameId: 10},
	}

	getCache := func(_ string, v interface{}) error {
		ptr, ok := v.(**gzapi.Challenge)
		if !ok {
			t.Fatalf("unexpected cache type")
		}
		// Return stale challenge id that should be ignored in favor of remote data
		*ptr = &gzapi.Challenge{Id: 99, Title: challengeConf.Name}
		return nil
	}

	orch := &SyncOrchestrator{
		conf:          conf,
		challengeConf: challengeConf,
		challenges:    remoteChallenges,
		api:           &gzapi.GZAPI{},
		getCache:      getCache,
	}

	if err := orch.determineSyncPath(); err != nil {
		t.Fatalf("determineSyncPath returned error: %v", err)
	}

	if orch.challengeData.Id != 2 {
		t.Fatalf("expected remote challenge id 2, got %d", orch.challengeData.Id)
	}
	if orch.challengeData.GameId != conf.Event.Id {
		t.Fatalf("expected GameId %d, got %d", conf.Event.Id, orch.challengeData.GameId)
	}
	if orch.challengeData.CS == nil {
		t.Fatalf("expected CS to be set from API")
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

	// Mock config with event name
	conf := &config.Config{
		EventName: "test-event",
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
				cacheData["test-event/Web/Test/challenge"] = gzapi.Challenge{
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
				cacheData["test-event/Web/Test/challenge"] = gzapi.Challenge{
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
			if got := IsConfigEdited(conf, &tt.challengeConf, &tt.challengeData, getCache); got != tt.want {
				t.Errorf("IsConfigEdited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeChallengeDataWithCategoryNormalization(t *testing.T) {
	tests := []struct {
		name          string
		challengeConf config.ChallengeYaml
		checkFunc     func(*testing.T, *gzapi.Challenge)
	}{
		{
			name: "Game Hacking category normalization",
			challengeConf: config.ChallengeYaml{
				Name:        "test-challenge",
				Author:      "test-author",
				Description: "Test description",
				Category:    "Game Hacking",
				Value:       500,
			},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				expectedTitle := "[Game Hacking] test-challenge"
				expectedCategory := "Reverse"
				if result.Title != expectedTitle {
					t.Errorf("Expected Title %q, got %q", expectedTitle, result.Title)
				}
				if result.Category != expectedCategory {
					t.Errorf("Expected Category %q, got %q", expectedCategory, result.Category)
				}
			},
		},
		{
			name: "Normal category no normalization",
			challengeConf: config.ChallengeYaml{
				Name:        "test-challenge",
				Author:      "test-author",
				Description: "Test description",
				Category:    "Web",
				Value:       500,
			},
			checkFunc: func(t *testing.T, result *gzapi.Challenge) {
				expectedTitle := "test-challenge"
				expectedCategory := "Web"
				if result.Title != expectedTitle {
					t.Errorf("Expected Title %q, got %q", expectedTitle, result.Title)
				}
				if result.Category != expectedCategory {
					t.Errorf("Expected Category %q, got %q", expectedCategory, result.Category)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challengeData := &gzapi.Challenge{}
			result := MergeChallengeData(&tt.challengeConf, challengeData)
			tt.checkFunc(t, result)
		})
	}
}

func TestIsChallengeExistWithNormalizedNames(t *testing.T) {
	// Test that IsChallengeExist correctly handles normalized names
	challenges := []gzapi.Challenge{
		{Title: "[Game Hacking] static-attachment", Category: "Reverse"},
		{Title: "web-challenge", Category: "Web"},
		{Title: "crypto-challenge", Category: "Crypto"},
	}

	tests := []struct {
		name          string
		challengeName string
		want          bool
	}{
		{
			name:          "normalized Game Hacking challenge exists",
			challengeName: "[Game Hacking] static-attachment",
			want:          true,
		},
		{
			name:          "original name without prefix doesn't exist",
			challengeName: "static-attachment",
			want:          false,
		},
		{
			name:          "normal challenge exists",
			challengeName: "web-challenge",
			want:          true,
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
