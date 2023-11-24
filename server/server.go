package main

import (
	"database/sql"
	"log"
	"mailinglist/grpcapi"
	"mailinglist/jsonapi"
	"mailinglist/mdb"
	"sync"

	"github.com/alexflint/go-arg"
)

// Define a struct to hold command-line arguments.
var args struct {
	DbPath   string `arg:"env:MAILINGLIST_DB"`        // Database file path, configurable via environment variable MAILINGLIST_DB
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON"` // JSON API server binding address, configurable via environment variable MAILINGLIST_BIND_JSON
	BindGrpc string `arg:"env:MAILINGLIST_BIND_GRPC"` // gRPC API server binding address, configurable via environment variable MAILINGLIST_BIND_GRPC
}

func main() {
	// Parse command-line arguments using the go-arg library.
	arg.MustParse(&args)

	// Set default values if the corresponding command-line arguments are not provided.
	if args.DbPath == "" {
		args.DbPath = "list.db"
	}
	if args.BindJson == "" {
		args.BindJson = ":8080"
	}
	if args.BindGrpc == "" {
		args.BindGrpc = ":8081"
	}

	// Log the selected configuration.
	log.Printf("Using Database '%v'\n", args.DbPath)

	// Open a connection to the SQLite database.
	db, err := sql.Open("sqlite3", args.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Defer closing the database connection to ensure it happens before exiting the program.
	defer db.Close()

	// Try to create the necessary tables in the database.
	mdb.TryCreate(db)

	// Use a WaitGroup to wait for both goroutines to finish before exiting the program.
	var wg sync.WaitGroup

	// Start the JSON API server in a goroutine.
	wg.Add(1)
	go func() {
		log.Printf("starting JSON API server....\n")
		jsonapi.Serve(db, args.BindJson)
		wg.Done()
	}()

	// Start the gRPC API server in a goroutine.
	wg.Add(1)
	go func() {
		log.Printf("starting gRPC API server....\n")
		grpcapi.Serve(db, args.BindGrpc)
		wg.Done()
	}()

	// Wait for both goroutines to finish before exiting the program.
	wg.Wait()
}
