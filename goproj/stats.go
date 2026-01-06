package main

import (
	"fmt"
	"os"
	"context"
	"log"
	"time"
	// "reflect"
	_ "embed"

	"database/sql"
	_ "modernc.org/sqlite"

	"micahelliott/ttypist/stats"

	// _ "github.com/go-sql-driver/mysql"
	"github.com/guregu/null/v6"
)

//go:embed schema.sql
var schemaSQL string

//go:embed words-populate.sql
var wordsSQL string

func run() error {
	var db *sql.DB
	ctx := context.Background()

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
	db, err = sql.Open("sqlite", dbPath)
	if err != nil { return err }
	defer db.Close()

	// Apply the schema using the schema.sql file content (manual step here for simplicity)
    // For a production system, use migrations
	if bootstrap {
		// TODO use go embedded file
		// schemaSQL, err := os.ReadFile("schema.sql")
		// if err != nil {log.Fatal(err)}
		// Create tables and populate questions
		fmt.Println("Opening schema and question population file and applying it")
		if _, err := db.Exec(string(schemaSQL)); err != nil {log.Fatal(err)}
		// Populate learnables (sql generated manually by gen-learnables.sh)
		fmt.Println("Opening learnable words file and applying it")
		if _, err := db.Exec(string(wordsSQL));  err != nil {log.Fatal(err)}
	}

	queries := stats.New(db)

	// list all authors
	words, err := queries.ListLearnables(ctx)
	if err != nil { return err }
	log.Println("words:", words)


	// low, hi := int64(1), int64(20)
	xs, err := queries.GetQuestionsInBand(ctx, stats.GetQuestionsInBandParams{
		// Low: &low, Hi: &hi, Bandsize: 200})
		// Low: sql.NullInt64{Int64: 1, Valid: true}, Hi: sql.NullInt64{Int64: 20, Valid: true}, Bandsize: 200})
		Low: null.IntFrom(1), Hi: null.IntFrom(20), Bandsize: 20})




	fmt.Println("\nquestions: ", xs )
	// fmt.Println("\nquestions1: ", xs[0] )
	// fmt.Println("\nquestions1: ", xs[0].Lrank, xs[0].Defn )
	for i := range xs { fmt.Println(xs[i].Lid, xs[i].Course, xs[i].Lrank.Int64, xs[i].Defn.String, xs[i].Lname,
		xs[i].Qid, xs[i].Qtype, xs[i].Defn.Valid) }


	// // create an author
	// result, err := queries.CreateLearnable(ctx, stats.CreateTrainingParams{
	// 	Canon:   "supercontrived",
	// 	AsTyped: sql.NullString{String: "suderconrthnrou", Valid: true},
	// 	IsCorrect: false,
	// 	Esecs: 1766509487,
	// 	TimeTaken: float32,
	// 	// Bio:  sql.NullString{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
	// })
	// if err != nil {return err}

	// var x *string
	// y := "keybr"
	fmt.Println("creating encounter")
	t0 := time.Now()
	result, err := queries.CreateEncounter(ctx, stats.CreateEncounterParams{
		Qid: 21,
		Tstamp: 123, Estamp: time.Now(), Entered: "stnr", Timer: time.Now().Sub(t0), Correct: true,
		// Qid: sql.NullString{String: "", Valid: true},
		// Acty: &y,
		// Acty: sql.NullString{"keybr", true},
	})
	fmt.Println(result)
	fmt.Println(err)


	// insertedEncounterID, err := result.LastInsertId()
	// if err != nil {return err}
	// log.Println(insertedEncounterID)
	// // get the author we just inserted
	// fetchedLearnable , err := queries.GetLearnable(ctx, insertedEncounterID)
	// if err != nil {return err}
	// // prints true
	// log.Println(reflect.DeepEqual(insertedEncounterID, fetchedLearnable.Lid))

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
