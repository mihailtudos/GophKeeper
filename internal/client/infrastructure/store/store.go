package store

import (
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mihailtudos/gophkeeper/cmd/client"
	"time"
)

type Record struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SType     string    `json:"s_type"`
	SName     string    `json:"s_name"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

	if _, err = s.con.Exec(createTabeSmt); err != nil {
		return err
	}

	return nil
}

func (s *Store) UploadSecrets(secrets []main.Secret) error {
	return nil
}

func (s *Store) GetRecords() ([]Record, error) {
	time.Sleep(time.Second * 5)
	rows, err := s.con.Query("SELECT id, title, body FROM records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Record{}
	for rows.Next() {
		var note Record
		if err := rows.Scan(&note.ID, &note.SName, &note.Data); err != nil {
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

	_, err := s.con.Exec(upsertStmt, r.ID, r.SName, r.Data)
	if err != nil {
		return err
	}

	return nil
}
