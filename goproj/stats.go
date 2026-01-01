package main

import (
	"fmt"
	"os"
	"context"
	"log"
	"reflect"
	_ "embed"

	"database/sql"
	_ "modernc.org/sqlite"

	// _ "github.com/go-sql-driver/mysql"

	"micahelliott/ttypist/stats"
)

var db *sql.DB

//go:embed schema.sql
var schemaSQL string

func run() error {
	ctx := context.Background()

	// db, err := sql.Open("mysql", "user:password@/dbname?parseTime=true")

	// Create DB if needed
	// dbPath := "./ttypist.db?_foreign_keys=on"
	dbPath := "./ttypist.db"

	bootstrap := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		bootstrap = true

        file, err := os.Create(dbPath)
        if err != nil {log.Fatal(err)}
        file.Close()
        fmt.Println("Database file created")
    }

	var err error
	// db, err = sql.Open("sqlite", ":memory:")
	db, err = sql.Open("sqlite", dbPath)
	if err != nil { return err }
	defer db.Close()

	// Apply the schema using the schema.sql file content (manual step here for simplicity)
    // For a production system, use migrations
	if bootstrap {
		fmt.Println("Opening schema file and applying it")
		// TODO use go embedded file
		// schemaSQL, err := os.ReadFile("schema.sql")
		// if err != nil {log.Fatal(err)}
		if _, err := db.Exec(string(schemaSQL)); err != nil {log.Fatal(err)}
	}

	queries := stats.New(db)

	// list all authors
	words, err := queries.ListLearnables(ctx)
	if err != nil { return err }
	log.Println(words)

	// create an author
	result, err := queries.CreateLearnable(ctx, stats.CreateTsessionParams{
		Canon:   "supercontrived",
		AsTyped: sql.NullString{String: "suderconrthnrou", Valid: true},
		IsCorrect: false,
		Esecs: 1766509487,
		TimeTaken: float32,
		// Bio:  sql.NullString{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
	})
	if err != nil {return err}

	insertedAuthorID, err := result.LastInsertId()
	if err != nil {return err}
	log.Println(insertedAuthorID)

	// get the author we just inserted
	fetchedAuthor, err := queries.GetLearnable(ctx, insertedAuthorID)
	if err != nil {return err}

	// prints true
	log.Println(reflect.DeepEqual(insertedAuthorID, fetchedAuthor.ID))
	return nil
}

// func main() {
// 	if err := run(); err != nil {
// 		log.Fatal(err)
// 	}
// }
