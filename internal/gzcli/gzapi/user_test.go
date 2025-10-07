package gzapi

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGZAPI_Users(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/admin/users": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []User{
					{Id: "user1", UserName: "Alice"},
					{Id: "user2", UserName: "Bob"},
				},
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	users, err := api.Users()
	if err != nil {
		t.Errorf("Users() failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Verify API is set
	for _, user := range users {
		if user.API == nil {
			t.Error("Expected API to be set for user")
		}
	}
}

func TestUser_Delete(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/admin/users/user123": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"deleted": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	user := &User{
		Id:  "user123",
		API: api,
	}

	err = user.Delete()
	if err != nil {
		t.Errorf("User.Delete() failed: %v", err)
	}
}

// Helper functions are in common_test.go
