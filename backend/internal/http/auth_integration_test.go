package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/librearchive/librearchive/internal/config"
	"github.com/librearchive/librearchive/internal/database"
	api "github.com/librearchive/librearchive/internal/http"
)

func TestBootstrapLoginRefreshAndLogout(t *testing.T) {
	temporaryDirectory := t.TempDir()
	cfg := config.Config{DatabasePath: filepath.Join(temporaryDirectory, "library.db"), StoragePath: temporaryDirectory, JWTSecret: "test-secret"}
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		t.Fatal(err)
	}
	server, err := api.NewServer(cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	bootstrap := request(server.Router(), http.MethodPost, "/api/v1/auth/bootstrap", map[string]string{"username": "admin", "name": "Administrator"}, "")
	if bootstrap.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d, body = %s", bootstrap.Code, bootstrap.Body.String())
	}
	var bootstrapBody struct {
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(bootstrap.Body).Decode(&bootstrapBody); err != nil {
		t.Fatal(err)
	}

	login := request(server.Router(), http.MethodPost, "/api/v1/auth/login", map[string]string{"username": "admin", "passphrase": bootstrapBody.Passphrase}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d", login.Code)
	}
	var loginBody struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(login.Body).Decode(&loginBody); err != nil {
		t.Fatal(err)
	}
	if loginBody.Token == "" || loginBody.RefreshToken == "" {
		t.Fatal("login did not return both tokens")
	}

	refresh := request(server.Router(), http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refreshToken": loginBody.RefreshToken}, "")
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status = %d", refresh.Code)
	}
	var refreshBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(refresh.Body).Decode(&refreshBody); err != nil {
		t.Fatal(err)
	}
	logout := request(server.Router(), http.MethodPost, "/api/v1/auth/logout", nil, refreshBody.Token)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", logout.Code)
	}
	me := request(server.Router(), http.MethodGet, "/api/v1/me", nil, refreshBody.Token)
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("revoked token status = %d", me.Code)
	}
}

func request(handler http.Handler, method, target string, body interface{}, token string) *httptest.ResponseRecorder {
	var encoded bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&encoded).Encode(body)
	}
	request := httptest.NewRequest(method, target, &encoded)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
