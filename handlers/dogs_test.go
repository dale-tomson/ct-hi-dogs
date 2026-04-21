package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dale-tomson/dogs-api/db"
	"github.com/dale-tomson/dogs-api/handlers"
	appMiddleware "github.com/dale-tomson/dogs-api/middleware"
	"github.com/go-chi/chi/v5"
)

// newTestRouter creates a fresh in-memory DB and a fully-wired Chi router.
// Call this at the start of every test function.
func newTestRouter(t *testing.T) *chi.Mux {
	t.Helper()
	db.Init(":memory:", "") // empty seedFile so no seed runs
	os.Setenv("API_KEY", "test-key")
	t.Cleanup(func() { os.Unsetenv("API_KEY") })

	r := chi.NewRouter()
	r.Get("/api/health", handlers.HealthCheck)
	r.Route("/api/dogs", func(r chi.Router) {
		r.Use(appMiddleware.APIKeyAuth)
		r.Get("/", handlers.ListDogs)
		r.Post("/", handlers.CreateDog)
		r.Get("/{breed}", handlers.GetDog)
		r.Put("/{breed}", handlers.UpdateDog)
		r.Delete("/{breed}", handlers.DeleteDog)
	})
	return r
}

func authedGET(t *testing.T, path string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-API-Key", "test-key")
	return req
}

func authedPOST(t *testing.T, path string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")
	return req
}

func authedPUT(t *testing.T, path string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")
	return req
}

func authedDELETE(t *testing.T, path string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set("X-API-Key", "test-key")
	return req
}

func TestHealthCheck(t *testing.T) {
	r := newTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("health: want 200 got %d", w.Code)
	}
}

func TestListDogs_EmptyDB(t *testing.T) {
	r := newTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedGET(t, "/api/dogs/"))
	if w.Code != http.StatusOK {
		t.Fatalf("list: want 200 got %d: %s", w.Code, w.Body)
	}
	var result []any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("list: decode error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("list: want empty array, got %d items", len(result))
	}
}

func TestCreateAndGetDog(t *testing.T) {
	r := newTestRouter(t)

	// Create
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedPOST(t, "/api/dogs/",
		`{"breed":"husky","sub_breeds":["siberian","alaskan"]}`))
	if w.Code != http.StatusCreated {
		t.Fatalf("create: want 201 got %d: %s", w.Code, w.Body)
	}

	// Get
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, authedGET(t, "/api/dogs/husky"))
	if w2.Code != http.StatusOK {
		t.Fatalf("get: want 200 got %d: %s", w2.Code, w2.Body)
	}
	var dog map[string]any
	json.NewDecoder(w2.Body).Decode(&dog)
	if dog["breed"] != "husky" {
		t.Errorf("get: want breed=husky got %v", dog["breed"])
	}
	subs, _ := dog["sub_breeds"].([]any)
	if len(subs) != 2 {
		t.Errorf("get: want 2 sub-breeds got %d", len(subs))
	}
}

func TestCreateDog_Duplicate(t *testing.T) {
	r := newTestRouter(t)
	body := `{"breed":"pug","sub_breeds":[]}`

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, authedPOST(t, "/api/dogs/", body))
	if w1.Code != http.StatusCreated {
		t.Fatalf("first create: want 201 got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, authedPOST(t, "/api/dogs/", body))
	if w2.Code != http.StatusConflict {
		t.Fatalf("duplicate create: want 409 got %d", w2.Code)
	}
}

func TestUpdateDog(t *testing.T) {
	r := newTestRouter(t)

	// Create first
	r.ServeHTTP(httptest.NewRecorder(),
		authedPOST(t, "/api/dogs/", `{"breed":"collie","sub_breeds":["border"]}`))

	// Update
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedPUT(t, "/api/dogs/collie", `{"sub_breeds":["border","rough"]}`))
	if w.Code != http.StatusOK {
		t.Fatalf("update: want 200 got %d: %s", w.Code, w.Body)
	}
	var dog map[string]any
	json.NewDecoder(w.Body).Decode(&dog)
	subs, _ := dog["sub_breeds"].([]any)
	if len(subs) != 2 {
		t.Errorf("update: want 2 sub-breeds got %d", len(subs))
	}
}

func TestDeleteDog(t *testing.T) {
	r := newTestRouter(t)

	// Create
	r.ServeHTTP(httptest.NewRecorder(),
		authedPOST(t, "/api/dogs/", `{"breed":"boxer","sub_breeds":[]}`))

	// Delete
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedDELETE(t, "/api/dogs/boxer"))
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204 got %d", w.Code)
	}

	// Confirm 404
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, authedGET(t, "/api/dogs/boxer"))
	if w2.Code != http.StatusNotFound {
		t.Fatalf("after-delete get: want 404 got %d", w2.Code)
	}
}

func TestGetDog_NotFound(t *testing.T) {
	r := newTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedGET(t, "/api/dogs/doesnotexist"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("not-found: want 404 got %d", w.Code)
	}
}

func TestMissingAPIKey(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/dogs/", nil)
	// Deliberately omit X-API-Key
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no-key: want 401 got %d", w.Code)
	}
}

func TestCreateDog_InvalidName(t *testing.T) {
	r := newTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedPOST(t, "/api/dogs/",
		`{"breed":"bad name!","sub_breeds":[]}`))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid name: want 400 got %d", w.Code)
	}
}
