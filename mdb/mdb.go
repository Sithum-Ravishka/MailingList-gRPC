package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

// Reading and Writing from database
type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

// create databse and table
func TryCreate(db *sql.DB) { // use "_" ignore return value
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
			// Code 1 == "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func EmailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmed_at int64
	var opt_out bool

	err := row.Scan(&id, &email, &confirmed_at, &opt_out) //Going to scan row in database / same order in colum appere

	if err != nil {
		log.Println(err)
		return nil, err
	}

	t := time.Unix(confirmed_at, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: opt_out}, nil
}

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

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {

	// run query and get back data as rows
	rows, err := db.Query(`
	SELECT id, email, confirmed_at, opt_out
	FROM emails
	WHERE email = ?`, email)
	// pass email to "?"

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() { //get back data
		return EmailEntryFromRow(rows)
	}
	return nil, nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
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

type GetEmailBatchQueryParams struct {
	Page  int // for paginate / for don't overlap email
	Count int // number of email returns
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	//select all information need in using EmailEntry
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
	defer rows.Close() // don't need after return outoff rows

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := EmailEntryFromRow(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *email)
	}

	return emails, nil
}
