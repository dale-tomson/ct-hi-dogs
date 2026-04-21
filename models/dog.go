package models

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/dale-tomson/dogs-api/db"
)

type Breed struct {
	ID        int      `json:"id"`
	Breed     string   `json:"breed"`
	SubBreeds []string `json:"sub_breeds"`
}

var (
	ErrNotFound  = errors.New("breed not found")
	ErrDuplicate = errors.New("breed already exists")
)

func Ping() error {
	return db.DB.Ping()
}

func GetAll() ([]Breed, error) {
	// Query 1: all breeds ordered by name
	breedRows, err := db.DB.Query(`SELECT id, name FROM breeds ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer breedRows.Close()

	var results []Breed
	idToIndex := make(map[int]int) // maps breed.id -> index in results slice

	for breedRows.Next() {
		var b Breed
		if err := breedRows.Scan(&b.ID, &b.Breed); err != nil {
			return nil, err
		}
		b.SubBreeds = []string{} // ensure never nil
		idToIndex[b.ID] = len(results)
		results = append(results, b)
	}
	if err := breedRows.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return []Breed{}, nil
	}

	subRows, err := db.DB.Query(
		`SELECT breed_id, name FROM sub_breeds ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer subRows.Close()

	for subRows.Next() {
		var breedID int
		var name string
		if err := subRows.Scan(&breedID, &name); err != nil {
			return nil, err
		}
		if idx, ok := idToIndex[breedID]; ok {
			results[idx].SubBreeds = append(results[idx].SubBreeds, name)
		}
	}
	if err := subRows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func GetByName(name string) (*Breed, error) {
	b := &Breed{SubBreeds: []string{}}

	err := db.DB.QueryRow(
		`SELECT id, name FROM breeds WHERE name = ?`, name,
	).Scan(&b.ID, &b.Breed)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := db.DB.Query(
		`SELECT name FROM sub_breeds WHERE breed_id = ? ORDER BY name ASC`, b.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub string
		if err := rows.Scan(&sub); err != nil {
			return nil, err
		}
		b.SubBreeds = append(b.SubBreeds, sub)
	}
	return b, rows.Err()
}

func Create(name string, subBreeds []string) (*Breed, error) {
	if subBreeds == nil {
		subBreeds = []string{}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO breeds (name) VALUES (?)`, name)
	if err != nil {
		if isSQLiteUnique(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}

	breedID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, sub := range subBreeds {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO sub_breeds (breed_id, name) VALUES (?, ?)`,
			breedID, sub,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Breed{ID: int(breedID), Breed: name, SubBreeds: subBreeds}, nil
}

func Update(name string, subBreeds []string) (*Breed, error) {
	if subBreeds == nil {
		subBreeds = []string{}
	}

	var breedID int
	err := db.DB.QueryRow(`SELECT id FROM breeds WHERE name = ?`, name).Scan(&breedID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM sub_breeds WHERE breed_id = ?`, breedID); err != nil {
		return nil, err
	}

	for _, sub := range subBreeds {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO sub_breeds (breed_id, name) VALUES (?, ?)`,
			breedID, sub,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Breed{ID: breedID, Breed: name, SubBreeds: subBreeds}, nil
}

func Delete(name string) error {
	res, err := db.DB.Exec(`DELETE FROM breeds WHERE name = ?`, name)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func isSQLiteUnique(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
