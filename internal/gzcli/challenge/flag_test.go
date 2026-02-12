//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package challenge

import (
	"net/http"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestIsFlagExist(t *testing.T) {
	flags := []gzapi.Flag{
		{Flag: "FLAG{test1}"},
		{Flag: "FLAG{test2}"},
		{Flag: "FLAG{test3}"},
	}

	tests := []struct {
		name string
		flag string
		want bool
	}{
		{
			name: "existing flag",
			flag: "FLAG{test2}",
			want: true,
		},
		{
			name: "non-existing flag",
			flag: "FLAG{notfound}",
			want: false,
		},
		{
			name: "empty flag",
			flag: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFlagExist(tt.flag, flags); got != tt.want {
				t.Errorf("IsFlagExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateChallengeFlags_CreateNew(t *testing.T) {
	flagCreated := false

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5/flags": func(w http.ResponseWriter, r *http.Request) {
			flagCreated = true
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})
	defer cleanup()

	conf := &config.Config{
		Event: gzapi.Game{
			Id: 1,
			CS: api,
		},
	}

	challengeConf := config.ChallengeYaml{
		Flags: []string{"FLAG{new}"},
	}

	challengeData := &gzapi.Challenge{
		Id:     5,
		GameId: 1,
		CS:     api,
		Flags:  []gzapi.Flag{}, // No existing flags
	}

	err := UpdateChallengeFlags(conf, challengeConf, challengeData)
	if err != nil {
		t.Errorf("UpdateChallengeFlags() failed: %v", err)
	}

	if !flagCreated {
		t.Error("Expected flag to be created")
	}
}

func TestUpdateChallengeFlags_DeleteOld(t *testing.T) {
	flagDeleted := false

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5/flags/10": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			flagDeleted = true
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"deleted": true}`))
		},
	})
	defer cleanup()

	conf := &config.Config{
		Event: gzapi.Game{
			Id: 1,
			CS: api,
		},
	}

	challengeConf := config.ChallengeYaml{
		Flags: []string{"FLAG{keep}"},
	}

	challengeData := &gzapi.Challenge{
		Id:     5,
		GameId: 1,
		CS:     api,
		Flags: []gzapi.Flag{
			{Id: 10, Flag: "FLAG{remove}"}, // This should be deleted
			{Id: 11, Flag: "FLAG{keep}"},   // This should be kept
		},
	}

	err := UpdateChallengeFlags(conf, challengeConf, challengeData)
	if err != nil {
		t.Errorf("UpdateChallengeFlags() failed: %v", err)
	}

	if !flagDeleted {
		t.Error("Expected old flag to be deleted")
	}
}

func TestUpdateChallengeFlags_NoChanges(t *testing.T) {
	api, cleanup := mockGZAPI(t, nil)
	defer cleanup()

	conf := &config.Config{
		Event: gzapi.Game{
			Id: 1,
			CS: api,
		},
	}

	challengeConf := config.ChallengeYaml{
		Flags: []string{"FLAG{existing}"},
	}

	challengeData := &gzapi.Challenge{
		Id:     5,
		GameId: 1,
		CS:     api,
		Flags: []gzapi.Flag{
			{Id: 1, Flag: "FLAG{existing}"},
		},
	}

	err := UpdateChallengeFlags(conf, challengeConf, challengeData)
	if err != nil {
		t.Errorf("UpdateChallengeFlags() with no changes failed: %v", err)
	}
}
