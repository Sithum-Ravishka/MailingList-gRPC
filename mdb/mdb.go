package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

// EmailEntry represents an entry in the emails table.
type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

// TryCreate attempts to create the emails table in the database if it does not exist.
func TryCreate(db *sql.DB) {
	// Use "_" to ignore the return value
	_, err := db.Exec(`
		CREATE TABLE emails (
			id  			INTEGER PRIMARY KEY,
			email 			TEXT UNIQUE,
			confirmed_at 	INTEGER,
			opt_out 		INTEGER
		);
	`)
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			// Code 1 indicates "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

// EmailEntryFromRow creates an EmailEntry from a database row.
func EmailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmed_at int64
	var opt_out bool

	// Scan the row into variables
	err := row.Scan(&id, &email, &confirmed_at, &opt_out)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Convert confirmed_at from UNIX timestamp to time.Time
	t := time.Unix(confirmed_at, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: opt_out}, nil
}

// CreateEmail inserts a new email entry into the emails table.
func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`INSERT INTO
		emails(email, confirmed_at, opt_out) 
		VALUES(?, 0, false)`, email)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetEmail retrieves an email entry from the database based on the email address.
func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	// Run query to get data as rows
	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE email = ?`, email)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		// Get data from row using EmailEntryFromRow function
		return EmailEntryFromRow(rows)
	}
	return nil, nil
}

// UpdateEmail updates an existing email entry in the emails table.
func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	// Convert confirmed_at to UNIX timestamp
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`
	INSERT INTO emails (id, email, confirmed_at, opt_out)
	VALUES (?, ?, ?, ?)
	ON CONFLICT (id) DO UPDATE SET
		email = excluded.email,
		confirmed_at = excluded.confirmed_at,
		opt_out = excluded.opt_out;
`, entry.Id, entry.Email, t, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// DeleteEmail marks an email as opted-out in the emails table.
func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		UPDATE emails
		SET opt_out=true
		WHERE email=?
	`, email)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetEmailBatchQueryParams represents parameters for batch email retrieval.
type GetEmailBatchQueryParams struct {
	Page  int // Page for pagination
	Count int // Number of emails to return
}

// GetEmailBatch retrieves a batch of emails from the emails table.
func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	// Select all required information using EmailEntry
	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out 
		FROM emails
		WHERE opt_out = false
		ORDER BY id ASC
		LIMIT ? OFFSET ?`, params.Count, (params.Page-1)*params.Count)

	if err != nil {
		log.Println(err)
		return empty, err
	}
	defer rows.Close()

	// Create a slice to store the retrieved emails
	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		// Get EmailEntry from row and append to the slice
		email, err := EmailEntryFromRow(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *email)
	}

	return emails, nil
}
