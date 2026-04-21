package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/dale-tomson/dogs-api/models"
	"github.com/go-chi/chi/v5"
)

var breedNameRe = regexp.MustCompile(`^[a-z]+$`)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func normalise(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validateName(name string) bool {
	return breedNameRe.MatchString(name)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := models.Ping(); err != nil {
		slog.Error("health check DB ping failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"db":     "unreachable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"db":     "ok",
	})
}

func ListDogs(w http.ResponseWriter, r *http.Request) {
	dogs, err := models.GetAll()
	if err != nil {
		slog.Error("ListDogs DB error", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch breeds")
		return
	}
	writeJSON(w, http.StatusOK, dogs)
}

func GetDog(w http.ResponseWriter, r *http.Request) {
	breedName := normalise(chi.URLParam(r, "breed"))

	dog, err := models.GetByName(breedName)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "breed not found")
		return
	}
	if err != nil {
		slog.Error("GetDog DB error", "breed", breedName, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch breed")
		return
	}
	writeJSON(w, http.StatusOK, dog)
}

func CreateDog(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Breed     string   `json:"breed"`
		SubBreeds []string `json:"sub_breeds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	body.Breed = normalise(body.Breed)
	if body.Breed == "" {
		writeError(w, http.StatusBadRequest, "breed name is required")
		return
	}
	if !validateName(body.Breed) {
		writeError(w, http.StatusBadRequest, "breed name must contain only lowercase letters a-z")
		return
	}

	cleaned := make([]string, 0, len(body.SubBreeds))
	for _, s := range body.SubBreeds {
		s = normalise(s)
		if s == "" {
			continue
		}
		if !validateName(s) {
			writeError(w, http.StatusBadRequest, "sub-breed names must contain only lowercase letters a-z")
			return
		}
		cleaned = append(cleaned, s)
	}

	dog, err := models.Create(body.Breed, cleaned)
	if errors.Is(err, models.ErrDuplicate) {
		writeError(w, http.StatusConflict, "breed already exists")
		return
	}
	if err != nil {
		slog.Error("CreateDog DB error", "breed", body.Breed, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create breed")
		return
	}
	writeJSON(w, http.StatusCreated, dog)
}

func UpdateDog(w http.ResponseWriter, r *http.Request) {
	breedName := normalise(chi.URLParam(r, "breed"))

	var body struct {
		SubBreeds []string `json:"sub_breeds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	cleaned := make([]string, 0, len(body.SubBreeds))
	for _, s := range body.SubBreeds {
		s = normalise(s)
		if s == "" {
			continue
		}
		if !validateName(s) {
			writeError(w, http.StatusBadRequest, "sub-breed names must contain only lowercase letters a-z")
			return
		}
		cleaned = append(cleaned, s)
	}

	dog, err := models.Update(breedName, cleaned)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "breed not found")
		return
	}
	if err != nil {
		slog.Error("UpdateDog DB error", "breed", breedName, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update breed")
		return
	}
	writeJSON(w, http.StatusOK, dog)
}

func DeleteDog(w http.ResponseWriter, r *http.Request) {
	breedName := normalise(chi.URLParam(r, "breed"))

	err := models.Delete(breedName)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "breed not found")
		return
	}
	if err != nil {
		slog.Error("DeleteDog DB error", "breed", breedName, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete breed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
