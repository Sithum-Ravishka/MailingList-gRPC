package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailinglist/mdb"
	"net/http"
)

// setJsonHeader sets the Content-Type header for JSON responses
func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// fromJson reads the request body and unmarshals it into the target object
func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

// returnJson sets the JSON response header and marshals the data or error into JSON format
func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr := withData()

	if serverErr != nil {
		w.WriteHeader(500)
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(serverErrJson)
		return
	}
	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataJson)
}

// returnErr is a helper function to return an error response in JSON format
func returnErr(w http.ResponseWriter, err error, code int) {
	// Call returnJson to set the JSON response header and marshal the error message into JSON
	returnJson(w, func() (interface{}, error) {
		// Create a struct with an "Err" field containing the error message
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}

		// Set the HTTP status code using the provided code argument
		w.WriteHeader(code)

		// Return the error message struct along with a nil error (indicating success)
		return errorMessage, nil
	})
}

// Serve sets up the JSON API server with different endpoints
func Serve(db *sql.DB, bind string) {
	// Create a handler at the specified address
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get_batch", GetEmailBatch(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	log.Printf("JSON API serve listening %v\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("JSON server error: %v", err)
	}
}

// GetEmail handles the GET request for retrieving a single email entry
func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		returnJson(w, func() (interface{}, error) {
			log.Panicf("JSON GetEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// CreateEmail handles the POST request for creating a new email entry
func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Panicf("JSON CreateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// UpdateEmail handles the PUT request for updating an existing email entry
func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// DeleteEmail handles the DELETE request for marking an email as opted-out
func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "DELETE" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Panicf("JSON DeleteEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// GetEmailBatch handles the GET request for retrieving a batch of email entries
func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		queryOptions := mdb.GetEmailBatchQueryParams{}
		fromJson(req.Body, &queryOptions)

		if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
			returnErr(w, errors.New("Page and Count fields are required and must be > 0"), 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Panicf("JSON GetEmailBatch: %v\n", queryOptions)
			return mdb.GetEmailBatch(db, queryOptions)
		})
	})
}
