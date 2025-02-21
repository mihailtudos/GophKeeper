package main

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Record struct {
	ID    string
	Title string
	Body  string
}

type Store struct {
	con *sql.DB
}

func (s *Store) Init() error {
	var err error
	s.con, err = sql.Open("sqlite3", "gophkeeper.db")
	if err != nil {
		return err
	}

	createTabeSmt := `
		CREATE TABLE IF NOT EXISTS records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			body TEXT
		);
	`

	if _, err := s.con.Exec(createTabeSmt); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetRecords() ([]Record, error) {
	rows, err := s.con.Query("SELECT id, title, body FROM records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Record{}
	for rows.Next() {
		var note Record
		if err := rows.Scan(&note.ID, &note.Title, &note.Body); err != nil {
			return nil, err
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (s *Store) SaveRecord(r Record) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	upsertStmt := `
		INSERT INTO records (id, title, body)
		VALUES (?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title =excluded.title, body =excluded.body;
	`

	_, err := s.con.Exec(upsertStmt, r.ID, r.Title, r.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Login(f loginForm) error {
	return nil
}
