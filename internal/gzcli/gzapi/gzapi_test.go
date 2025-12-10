//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package gzapi

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestInit_Success(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{
		Username: "testuser",
		Password: "testpass",
	}

	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if api == nil {
		t.Fatal("Init() returned nil API")
	}

	if api.Url != server.URL {
		t.Errorf("Expected URL %s, got %s", server.URL, api.Url)
	}

	if api.Creds != creds {
		t.Error("Credentials not set correctly")
	}

	if api.Client == nil {
		t.Error("HTTP client not initialized")
	}
}

func TestInit_LoginFailure(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid credentials"}`))
		},
	})
	defer server.Close()

	creds := &Creds{
		Username: "wronguser",
		Password: "wrongpass",
	}

	_, err := Init(server.URL, creds)
	if err == nil {
		t.Fatal("Expected Init() to fail with invalid credentials")
	}
}

func TestInit_TrimTrailingSlash(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{
		Username: "testuser",
		Password: "testpass",
	}

	// Test with trailing slash
	api, err := Init(server.URL+"/", creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify trailing slash was removed
	if api.Url != server.URL {
		t.Errorf("Expected URL %s (without trailing slash), got %s", server.URL, api.Url)
	}
}

func TestInit_ReusesCachedCookies(t *testing.T) {
	originalWD, _ := os.Getwd()
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to switch working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })

	loginCount := 0
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			loginCount++
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   "token123",
				Path:    "/",
				Expires: time.Now().Add(time.Hour),
			})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/protected": func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value != "token123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}

	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	var resp map[string]bool
	if err := api.get("/api/protected", &resp); err != nil {
		t.Fatalf("Protected request failed with fresh login: %v", err)
	}
	if loginCount != 1 {
		t.Fatalf("Expected 1 login, got %d", loginCount)
	}

	apiCached, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() with cached cookie failed: %v", err)
	}

	resp = map[string]bool{}
	if err := apiCached.get("/api/protected", &resp); err != nil {
		t.Fatalf("Protected request with cached cookie failed: %v", err)
	}

	if loginCount != 1 {
		t.Fatalf("Expected cached cookies to prevent re-login, got %d logins", loginCount)
	}
}

func TestInit_ReloginWhenCachedCookiesExpired(t *testing.T) {
	originalWD, _ := os.Getwd()
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to switch working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })

	loginCount := 0
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			loginCount++
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   "fresh",
				Path:    "/",
				Expires: time.Now().Add(time.Hour),
			})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/protected": func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value != "fresh" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok": true}`))
		},
	})
	defer server.Close()

	store, err := newCookieStore(server.URL)
	if err != nil {
		t.Fatalf("Failed to create cookie store: %v", err)
	}
	expiredJar := store.newJar()
	if expiredJar == nil {
		t.Fatal("Expected cookie jar to be created")
	}
	expiredJar.SetCookies(store.baseURL, []*http.Cookie{
		{
			Name:    "session",
			Value:   "expired",
			Path:    "/",
			Expires: time.Now().Add(-1 * time.Hour),
		},
	})
	if err := store.save(expiredJar); err != nil {
		t.Fatalf("Failed to seed expired cookie cache: %v", err)
	}

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	var resp map[string]bool
	if err := api.get("/api/protected", &resp); err != nil {
		t.Fatalf("Protected request failed after re-login: %v", err)
	}

	if loginCount != 1 {
		t.Fatalf("Expected login when cached cookie is expired, got %d logins", loginCount)
	}
}
func TestGZAPI_Get_Success(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/test": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"result": "success"})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	var result map[string]string
	err = api.get("/api/test", &result)
	if err != nil {
		t.Fatalf("get() failed: %v", err)
	}

	if result["result"] != "success" {
		t.Errorf("Expected result 'success', got %s", result["result"])
	}
}

func TestGZAPI_Get_NotFound(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/notfound": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	err = api.get("/api/notfound", nil)
	if err == nil {
		t.Fatal("Expected get() to fail with 404 status")
	}

	if !contains(err.Error(), "404") {
		t.Errorf("Expected error to mention 404, got: %v", err)
	}
}

func TestGZAPI_Post_Success(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/create": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var reqBody map[string]string
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody["name"] != "test" {
				t.Errorf("Expected name 'test', got %s", reqBody["name"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"name": reqBody["name"],
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	requestData := map[string]string{"name": "test"}
	var response map[string]interface{}

	err = api.post("/api/create", requestData, &response)
	if err != nil {
		t.Fatalf("post() failed: %v", err)
	}

	if response["name"] != "test" {
		t.Errorf("Expected response name 'test', got %v", response["name"])
	}
}

func TestGZAPI_Put_Success(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/update/1": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}

			var reqBody map[string]string
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      1,
				"updated": true,
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	updateData := map[string]string{"field": "value"}
	var response map[string]interface{}

	err = api.put("/api/update/1", updateData, &response)
	if err != nil {
		t.Fatalf("put() failed: %v", err)
	}

	if updated, ok := response["updated"].(bool); !ok || !updated {
		t.Error("Expected updated to be true")
	}
}

func TestGZAPI_Delete_Success(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/delete/1": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	var response map[string]bool
	err = api.delete("/api/delete/1", &response)
	if err != nil {
		t.Fatalf("delete() failed: %v", err)
	}

	if !response["deleted"] {
		t.Error("Expected deleted to be true")
	}
}

func TestGZAPI_PostMultiPart_FileNotFound(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/upload": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	err = api.postMultiPart("/api/upload", "/nonexistent/file.txt", nil)
	if err == nil {
		t.Fatal("Expected postMultiPart() to fail with non-existent file")
	}

	if !contains(err.Error(), "file not found") {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}
}

func TestGZAPI_PostMultiPart_Success(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	content := []byte("test file content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/upload": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			if err := r.ParseMultipartForm(32 << 20); err != nil {
				t.Errorf("Failed to parse multipart form: %v", err)
			}

			file, _, err := r.FormFile("files")
			if err != nil {
				t.Errorf("Failed to get form file: %v", err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					t.Errorf("Failed to close file: %v", err)
				}
			}()

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"uploaded": "success"})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	var response map[string]string
	err = api.postMultiPart("/api/upload", tmpFile.Name(), &response)
	if err != nil {
		t.Fatalf("postMultiPart() failed: %v", err)
	}

	if response["uploaded"] != "success" {
		t.Errorf("Expected uploaded 'success', got %s", response["uploaded"])
	}
}

func TestGZAPI_PutMultiPart_Success(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test file content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/update": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}

			if err := r.ParseMultipartForm(32 << 20); err != nil {
				t.Errorf("Failed to parse multipart form: %v", err)
			}

			file, _, err := r.FormFile("file")
			if err != nil {
				t.Errorf("Failed to get form file: %v", err)
			}
			defer file.Close()

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"updated": "success"})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	var response map[string]string
	err = api.putMultiPart("/api/update", tmpFile.Name(), &response)
	if err != nil {
		t.Fatalf("putMultiPart() failed: %v", err)
	}

	if response["updated"] != "success" {
		t.Errorf("Expected updated 'success', got %s", response["updated"])
	}
}

func TestRegister_Success(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/register": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
	})
	defer server.Close()

	regForm := &RegisterForm{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "testpass",
	}

	api, err := Register(server.URL, regForm)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	if api == nil {
		t.Fatal("Register() returned nil API")
	}

	if api.Creds.Username != regForm.Username {
		t.Errorf("Expected username %s, got %s", regForm.Username, api.Creds.Username)
	}

	if api.Creds.Password != regForm.Password {
		t.Errorf("Expected password %s, got %s", regForm.Password, api.Creds.Password)
	}
}

// Test Logout
func TestGZAPI_Logout(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/account/logout": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	err = api.Logout()
	if err != nil {
		t.Errorf("Logout() failed: %v", err)
	}
}

// Test Register error case
func TestRegister_Failure(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/register": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "user already exists"}`))
		},
	})
	defer server.Close()

	regForm := &RegisterForm{
		Email:    "existing@example.com",
		Username: "existinguser",
		Password: "testpass",
	}

	_, err := Register(server.URL, regForm)
	if err == nil {
		t.Fatal("Expected Register() to fail when user already exists")
	}
}

// Test Challenge CRUD operations
func TestChallenge_Delete(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	challenge := &Challenge{
		Id:     5,
		GameId: 1,
		Title:  "Test Challenge",
		CS:     api,
	}

	err = challenge.Delete()
	if err != nil {
		t.Errorf("Challenge.Delete() failed: %v", err)
	}
}

func TestChallenge_Update(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			var updated Challenge
			json.NewDecoder(r.Body).Decode(&updated)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(updated)
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	challenge := &Challenge{
		Id:     5,
		GameId: 1,
		Title:  "Original Title",
		CS:     api,
	}

	updateData := Challenge{
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	result, err := challenge.Update(updateData)
	if err != nil {
		t.Errorf("Challenge.Update() failed: %v", err)
	}

	if result.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", result.Title)
	}
}

func TestChallenge_Refresh(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Challenge{
				Id:      5,
				Title:   "Refreshed Challenge",
				Content: "Refreshed Content",
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	challenge := &Challenge{
		Id:     5,
		GameId: 1,
		CS:     api,
	}

	refreshed, err := challenge.Refresh()
	if err != nil {
		t.Errorf("Challenge.Refresh() failed: %v", err)
	}

	if refreshed.Title != "Refreshed Challenge" {
		t.Errorf("Expected title 'Refreshed Challenge', got %s", refreshed.Title)
	}

	if refreshed.GameId != 1 {
		t.Errorf("Expected GameId 1, got %d", refreshed.GameId)
	}

	if refreshed.CS == nil {
		t.Error("Expected CS to be set after refresh")
	}
}

// Test Game challenge operations
func TestGame_CreateChallenge(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var form CreateChallengeForm
			json.NewDecoder(r.Body).Decode(&form)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Challenge{
				Id:       10,
				Title:    form.Title,
				Category: form.Category,
				Type:     form.Type,
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	game := &Game{
		Id: 1,
		CS: api,
	}

	form := CreateChallengeForm{
		Title:    "New Challenge",
		Category: "Web",
		Tag:      "Web",
		Type:     "StaticAttachment",
	}

	challenge, err := game.CreateChallenge(form)
	if err != nil {
		t.Errorf("Game.CreateChallenge() failed: %v", err)
	}

	if challenge.Id != 10 {
		t.Errorf("Expected challenge ID 10, got %d", challenge.Id)
	}

	if challenge.GameId != 1 {
		t.Errorf("Expected GameId 1, got %d", challenge.GameId)
	}

	if challenge.CS == nil {
		t.Error("Expected CS to be set")
	}
}

func TestGame_GetChallenges(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			// Return list first, then individual challenges
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]Challenge{
				{Id: 1, Title: "Challenge 1"},
				{Id: 2, Title: "Challenge 2"},
			})
		},
		"/api/edit/games/1/challenges/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Challenge{Id: 1, Title: "Challenge 1", Content: "Full content 1"})
		},
		"/api/edit/games/1/challenges/2": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Challenge{Id: 2, Title: "Challenge 2", Content: "Full content 2"})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	game := &Game{Id: 1, CS: api}

	challenges, err := game.GetChallenges()
	if err != nil {
		t.Errorf("Game.GetChallenges() failed: %v", err)
	}

	if len(challenges) != 2 {
		t.Errorf("Expected 2 challenges, got %d", len(challenges))
	}

	// Verify CS and GameId are set
	for _, c := range challenges {
		if c.CS == nil {
			t.Error("Expected CS to be set for challenge")
		}
		if c.GameId != 1 {
			t.Errorf("Expected GameId 1, got %d", c.GameId)
		}
	}
}

func TestGame_GetChallenge(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]Challenge{
				{Id: 1, Title: "Challenge 1"},
				{Id: 2, Title: "Target Challenge"},
			})
		},
		"/api/edit/games/1/challenges/2": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Challenge{
				Id:      2,
				Title:   "Target Challenge",
				Content: "Full content",
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	game := &Game{Id: 1, CS: api}

	challenge, err := game.GetChallenge("Target Challenge")
	if err != nil {
		t.Errorf("Game.GetChallenge() failed: %v", err)
	}

	if challenge.Id != 2 {
		t.Errorf("Expected challenge ID 2, got %d", challenge.Id)
	}

	if challenge.Title != "Target Challenge" {
		t.Errorf("Expected title 'Target Challenge', got %s", challenge.Title)
	}

	if challenge.CS == nil {
		t.Error("Expected CS to be set")
	}

	if challenge.GameId != 1 {
		t.Errorf("Expected GameId 1, got %d", challenge.GameId)
	}
}

func TestGame_GetChallenge_NotFound(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]Challenge{
				{Id: 1, Title: "Challenge 1"},
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	game := &Game{Id: 1, CS: api}

	_, err = game.GetChallenge("Nonexistent Challenge")
	if err == nil {
		t.Error("Expected error for nonexistent challenge")
	}

	if !contains(err.Error(), "challenge not found") {
		t.Errorf("Expected 'challenge not found' error, got: %v", err)
	}
}

// Test Flag operations
func TestChallenge_GetFlags(t *testing.T) {
	creds := &Creds{Username: "test", Password: "test"}
	api := &GZAPI{
		Url:   "http://test.com",
		Creds: creds,
	}

	challenge := &Challenge{
		Id:     1,
		GameId: 1,
		CS:     api,
		Flags: []Flag{
			{Id: 1, Flag: "FLAG{test1}"},
			{Id: 2, Flag: "FLAG{test2}"},
		},
	}

	flags := challenge.GetFlags()

	if len(flags) != 2 {
		t.Errorf("Expected 2 flags, got %d", len(flags))
	}

	// Verify CS, GameId, and ChallengeId are set
	for _, flag := range flags {
		if flag.CS == nil {
			t.Error("Expected CS to be set for flag")
		}
		if flag.GameId != 1 {
			t.Errorf("Expected GameId 1, got %d", flag.GameId)
		}
		if flag.ChallengeId != 1 {
			t.Errorf("Expected ChallengeId 1, got %d", flag.ChallengeId)
		}
	}
}

func TestChallenge_CreateFlag(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5/flags": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var flags []CreateFlagForm
			json.NewDecoder(r.Body).Decode(&flags)

			if len(flags) != 1 {
				t.Errorf("Expected 1 flag in array, got %d", len(flags))
			}

			if flags[0].Flag != "FLAG{test123}" {
				t.Errorf("Expected flag 'FLAG{test123}', got %s", flags[0].Flag)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	challenge := &Challenge{
		Id:     5,
		GameId: 1,
		CS:     api,
	}

	err = challenge.CreateFlag(CreateFlagForm{Flag: "FLAG{test123}"})
	if err != nil {
		t.Errorf("Challenge.CreateFlag() failed: %v", err)
	}
}

func TestFlag_Delete(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/1/challenges/5/flags/3": func(w http.ResponseWriter, r *http.Request) {
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

	flag := &Flag{
		Id:          3,
		GameId:      1,
		ChallengeId: 5,
		CS:          api,
	}

	err = flag.Delete()
	if err != nil {
		t.Errorf("Flag.Delete() failed: %v", err)
	}
}

// TestGZAPI_Post_EmptyResponse tests handling of empty response body
func TestGZAPI_Post_EmptyResponse(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/empty": func(w http.ResponseWriter, r *http.Request) {
			// Return 200 with empty body (simulates second login scenario)
			w.WriteHeader(http.StatusOK)
			// No body written
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test POST with empty response body - should not error
	var response LoginResponse
	err = api.post("/api/empty", nil, &response)
	if err != nil {
		t.Fatalf("post() with empty response failed: %v", err)
	}

	t.Log("Empty POST response handled successfully")
}

// TestGZAPI_Get_EmptyResponse tests handling of empty response body for GET
//
//nolint:dupl // Similar test structure for different HTTP methods is intentional
func TestGZAPI_Get_EmptyResponse(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/empty": func(w http.ResponseWriter, r *http.Request) {
			// Return 200 with empty body
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test GET with empty response body - should not error
	var response map[string]interface{}
	err = api.get("/api/empty", &response)
	if err != nil {
		t.Fatalf("get() with empty response failed: %v", err)
	}

	t.Log("Empty GET response handled successfully")
}

// TestGZAPI_Put_EmptyResponse tests handling of empty response body for PUT
func TestGZAPI_Put_EmptyResponse(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/update": func(w http.ResponseWriter, r *http.Request) {
			// Return 200 with empty body
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test PUT with empty response body - should not error
	var response map[string]interface{}
	err = api.put("/api/update", map[string]string{"key": "value"}, &response)
	if err != nil {
		t.Fatalf("put() with empty response failed: %v", err)
	}

	t.Log("Empty PUT response handled successfully")
}

// TestGZAPI_Delete_EmptyResponse tests handling of empty response body for DELETE
//
//nolint:dupl // Similar test structure for different HTTP methods is intentional
func TestGZAPI_Delete_EmptyResponse(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/resource/123": func(w http.ResponseWriter, r *http.Request) {
			// Return 200 with empty body
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test DELETE with empty response body - should not error
	var response map[string]interface{}
	err = api.delete("/api/resource/123", &response)
	if err != nil {
		t.Fatalf("delete() with empty response failed: %v", err)
	}

	t.Log("Empty DELETE response handled successfully")
}

// TestGZAPI_Login_SecondTime tests logging in when already authenticated
func TestGZAPI_Login_SecondTime(t *testing.T) {
	loginCount := 0
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			loginCount++
			w.WriteHeader(http.StatusOK)
			// First login returns proper JSON, subsequent logins return empty body
			// This simulates the real-world scenario where session is already active
			if loginCount == 1 {
				w.Write([]byte(`{"succeeded": true}`))
			} // Empty response for subsequent logins (already authenticated)
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}

	// First login
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("First Init() failed: %v", err)
	}

	// Second login attempt (reusing same session)
	err = api.Login()
	if err != nil {
		t.Fatalf("Second Login() failed (empty body should be handled): %v", err)
	}

	if loginCount != 2 {
		t.Errorf("Expected 2 login attempts, got %d", loginCount)
	}

	t.Log("Second login with empty response handled successfully")
}

// TestGZAPI_Post_NilData tests POST with nil data parameter
func TestGZAPI_Post_NilData(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/account/login": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		},
		"/api/action": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Empty response
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test POST with nil data parameter - should not error
	err = api.post("/api/action", map[string]string{"action": "test"}, nil)
	if err != nil {
		t.Fatalf("post() with nil data failed: %v", err)
	}

	t.Log("POST with nil data parameter handled successfully")
}

// Helper functions are in common_test.go
