package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/librearchive/librearchive/internal/config"
	"github.com/librearchive/librearchive/internal/database"
	api "github.com/librearchive/librearchive/internal/http"
)

type apiFixture struct {
	t                                                                               *testing.T
	handler                                                                         http.Handler
	adminToken, readerToken                                                         string
	adminID, readerID, bookID, categoryID, tagID, collectionID, highlightID, noteID uint
}

func TestAllAPIEndpoints(t *testing.T) {
	f := newFixture(t)
	f.expect(http.MethodGet, "/health", nil, "", http.StatusOK, nil)
	f.expect(http.MethodGet, "/ready", nil, "", http.StatusOK, nil)

	// Administrator user management, including reader authorization denial.
	f.expect(http.MethodGet, "/api/v1/users", nil, f.adminToken, http.StatusOK, nil)
	createdReader := f.expect(http.MethodPost, "/api/v1/users", map[string]string{"username": "reader", "name": "Reader"}, f.adminToken, http.StatusCreated, nil)
	f.readerID = nestedUint(t, createdReader, "user", "id")
	readerPassphrase := nestedString(t, createdReader, "passphrase")
	f.expect(http.MethodPatch, "/api/v1/users/"+id(f.readerID), map[string]string{"name": "Updated Reader"}, f.adminToken, http.StatusOK, nil)
	reset := f.expect(http.MethodPost, "/api/v1/users/"+id(f.readerID)+"/reset-passphrase", nil, f.adminToken, http.StatusOK, nil)
	readerPassphrase = nestedString(t, reset, "passphrase")
	f.readerToken = f.login("reader", readerPassphrase)
	f.expect(http.MethodGet, "/api/v1/users", nil, f.readerToken, http.StatusForbidden, nil)

	// Administrator catalog metadata and book management.
	category := f.expect(http.MethodPost, "/api/v1/categories", map[string]string{"name": "Science"}, f.adminToken, http.StatusCreated, nil)
	f.categoryID = uintValue(t, category, "id")
	f.expect(http.MethodGet, "/api/v1/categories", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/categories/"+id(f.categoryID), map[string]string{"name": "Science Fiction"}, f.adminToken, http.StatusOK, nil)
	tag := f.expect(http.MethodPost, "/api/v1/tags", map[string]string{"name": "space"}, f.adminToken, http.StatusCreated, nil)
	f.tagID = uintValue(t, tag, "id")
	f.expect(http.MethodGet, "/api/v1/tags", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/tags/"+id(f.tagID), map[string]string{"name": "classic"}, f.adminToken, http.StatusOK, nil)
	f.bookID = f.createBook()
	f.expect(http.MethodGet, "/api/v1/books?search=Test&page=1&pageSize=20", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/books/"+id(f.bookID), nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/books/"+id(f.bookID)+"/file", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/books/"+id(f.bookID)+"/cover", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/books/"+id(f.bookID), map[string]string{"title": "Updated Test Book"}, f.adminToken, http.StatusOK, nil)
	f.expect(http.MethodPut, "/api/v1/books/"+id(f.bookID)+"/categories/"+id(f.categoryID), nil, f.adminToken, http.StatusOK, nil)
	f.expect(http.MethodPut, "/api/v1/books/"+id(f.bookID)+"/tags/"+id(f.tagID), nil, f.adminToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/books?categoryId="+id(f.categoryID)+"&tagId="+id(f.tagID), nil, f.readerToken, http.StatusOK, nil)

	// Reader profile, progress, downloads, favorites, and collections.
	f.expect(http.MethodGet, "/api/v1/me", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/me", map[string]string{"name": "Reader Profile"}, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/books/"+id(f.bookID)+"/reading-progress", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPut, "/api/v1/books/"+id(f.bookID)+"/reading-progress", map[string]interface{}{"status": "in_progress", "currentPage": 2, "progressPercent": 10.5, "isDownloaded": true}, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/reading-progress?status=in_progress&downloaded=true", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodGet, "/api/v1/favorites", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPut, "/api/v1/books/"+id(f.bookID)+"/favorite", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodDelete, "/api/v1/books/"+id(f.bookID)+"/favorite", nil, f.readerToken, http.StatusNoContent, nil)
	f.expect(http.MethodGet, "/api/v1/collections", nil, f.readerToken, http.StatusOK, nil)
	collection := f.expect(http.MethodPost, "/api/v1/collections", map[string]string{"name": "To Read"}, f.readerToken, http.StatusCreated, nil)
	f.collectionID = uintValue(t, collection, "id")
	f.expect(http.MethodGet, "/api/v1/collections/"+id(f.collectionID), nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/collections/"+id(f.collectionID), map[string]string{"name": "Reading"}, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPut, "/api/v1/collections/"+id(f.collectionID)+"/books/"+id(f.bookID), nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodDelete, "/api/v1/collections/"+id(f.collectionID)+"/books/"+id(f.bookID), nil, f.readerToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/collections/"+id(f.collectionID), nil, f.readerToken, http.StatusNoContent, nil)

	// Highlights and notes, including each list/filter and mutation endpoint.
	highlight := f.expect(http.MethodPost, "/api/v1/books/"+id(f.bookID)+"/highlights", map[string]interface{}{"selectedText": "A memorable passage", "color": "yellow", "page": 2, "startOffset": 0, "endOffset": 18, "chapterRef": "Chapter 1"}, f.readerToken, http.StatusCreated, nil)
	f.highlightID = uintValue(t, highlight, "id")
	f.expect(http.MethodGet, "/api/v1/books/"+id(f.bookID)+"/annotations?type=highlight&color=yellow", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/highlights/"+id(f.highlightID), map[string]interface{}{"selectedText": "Updated passage", "color": "blue", "page": 2}, f.readerToken, http.StatusOK, nil)
	note := f.expect(http.MethodPost, "/api/v1/highlights/"+id(f.highlightID)+"/notes", map[string]string{"content": "A personal note"}, f.readerToken, http.StatusCreated, nil)
	f.noteID = uintValue(t, note, "id")
	f.expect(http.MethodGet, "/api/v1/books/"+id(f.bookID)+"/annotations?type=note&color=blue", nil, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodPatch, "/api/v1/notes/"+id(f.noteID), map[string]string{"content": "An updated note"}, f.readerToken, http.StatusOK, nil)
	f.expect(http.MethodDelete, "/api/v1/notes/"+id(f.noteID), nil, f.readerToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/highlights/"+id(f.highlightID), nil, f.readerToken, http.StatusNoContent, nil)

	// Final administrator-only removal endpoints.
	f.expect(http.MethodDelete, "/api/v1/books/"+id(f.bookID)+"/categories/"+id(f.categoryID), nil, f.adminToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/books/"+id(f.bookID)+"/tags/"+id(f.tagID), nil, f.adminToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/categories/"+id(f.categoryID), nil, f.adminToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/tags/"+id(f.tagID), nil, f.adminToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/books/"+id(f.bookID), nil, f.adminToken, http.StatusNoContent, nil)
	f.expect(http.MethodDelete, "/api/v1/users/"+id(f.readerID), nil, f.adminToken, http.StatusNoContent, nil)
}

func newFixture(t *testing.T) *apiFixture {
	t.Helper()
	root := t.TempDir()
	storage := filepath.Join(root, "storage")
	if err := os.MkdirAll(filepath.Join(storage, "books"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(storage, "covers"), 0750); err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(filepath.Join(root, "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	server, err := api.NewServer(config.Config{DatabasePath: filepath.Join(root, "library.db"), StoragePath: storage, JWTSecret: "integration-test-secret", MaxUploadBytes: 250 << 20}, db)
	if err != nil {
		t.Fatal(err)
	}
	f := &apiFixture{t: t, handler: server.Router()}
	bootstrap := f.expect(http.MethodPost, "/api/v1/auth/bootstrap", map[string]string{"username": "admin", "name": "Administrator"}, "", http.StatusCreated, nil)
	f.adminID = nestedUint(t, bootstrap, "user", "id")
	f.adminToken = f.login("admin", nestedString(t, bootstrap, "passphrase"))
	return f
}
func (f *apiFixture) login(username, phrase string) string {
	response := f.expect(http.MethodPost, "/api/v1/auth/login", map[string]string{"username": username, "passphrase": phrase}, "", http.StatusOK, nil)
	return nestedString(f.t, response, "token")
}
func (f *apiFixture) createBook() uint {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "Test Book")
	_ = writer.WriteField("author", "Test Author")
	pdf, _ := writer.CreateFormFile("file", "book.pdf")
	_, _ = pdf.Write([]byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n%%EOF"))
	cover, _ := writer.CreateFormFile("cover", "cover.png")
	_, _ = cover.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	_ = writer.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/books", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+f.adminToken)
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		f.t.Fatalf("create book status=%d body=%s", response.Code, response.Body.String())
	}
	var value map[string]interface{}
	_ = json.NewDecoder(response.Body).Decode(&value)
	return uintValue(f.t, value, "id")
}
func (f *apiFixture) expect(method, target string, body interface{}, token string, want int, _ interface{}) map[string]interface{} {
	f.t.Helper()
	var encoded bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&encoded).Encode(body); err != nil {
			f.t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, target, &encoded)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)
	if response.Code != want {
		f.t.Fatalf("%s %s: status=%d want=%d body=%s", method, target, response.Code, want, response.Body.String())
	}
	var decoded interface{}
	if response.Body.Len() > 0 && response.Header().Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
			f.t.Fatal(err)
		}
	}
	value, _ := decoded.(map[string]interface{})
	return value
}
func id(value uint) string { return strconv.FormatUint(uint64(value), 10) }
func uintValue(t *testing.T, value map[string]interface{}, key string) uint {
	t.Helper()
	number, ok := value[key].(float64)
	if !ok {
		t.Fatalf("missing numeric %s in %#v", key, value)
	}
	return uint(number)
}
func nestedUint(t *testing.T, value map[string]interface{}, parent, key string) uint {
	t.Helper()
	child, ok := value[parent].(map[string]interface{})
	if !ok {
		t.Fatalf("missing object %s", parent)
	}
	return uintValue(t, child, key)
}
func nestedString(t *testing.T, value map[string]interface{}, key string) string {
	t.Helper()
	text, ok := value[key].(string)
	if !ok || text == "" {
		t.Fatalf("missing string %s in %#v", key, value)
	}
	return text
}

var _ = fmt.Sprintf
