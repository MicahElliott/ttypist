package main

// TODO
// break display of multiple lines into typing prompts
// display wrong words
// organize into fns
//

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
	"math"
	"context"
	// "os/exec"
	// "errors"
	// "log"
	"database/sql"

	"golang.org/x/term"

	"micahelliott/ttypist/stats"
)

var db *sql.DB

type Word struct {
	stats.Learnable
	stats.Question
	stats.Encounter
}

// type	Learnable struct {
// 	Canon     string
// 	AsTyped   string
// 	IsCorrect bool
// 	TimeTaken time.Duration
// }

func readIncoming() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		// Log the error, but the program can continue or exit depending on your needs.
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
	}
	fmt.Printf("Read %d bytes from stdin:\n%s", len(data), string(data))
}

const (
	ColorRed    = "\x1b[31m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m"
	ColorReset  = "\x1b[0m"
)

func	printTypos(words []stats.Encounter) {
	fmt.Println("\nTypos: ...")
	for _, w := range words {
		if !w.Iscorrect { fmt.Println(w) }
	}
}

// Conduct a single-line typing test.
func	conductTestLine(words *[]Word, tgts []string, st0 time.Time) {
	fmt.Println("\n  " + strings.Join(tgts, " "))
	fmt.Print("> ")

	// Get the file descriptor for standard input. This messes with printing
	// further output, so all from here needs to be handled carefully and then
	// "restored"
	oldState, err := term.MakeRaw(int(syscall.Stdin))
	if err != nil { panic(err) }

	var typedAcc strings.Builder
	timeDurMult := 250 // millis per char, or is considered "slow"
	// Set up for not starting timer till first char it typed
	word1 := true

	for	_, tgt := range tgts {
		t0 := time.Now()
		for	{ // Loop over every char (as it's typed)
			var buf [1]byte
			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				if err == io.EOF { break }
				panic(err)
			}
			if	n == 0 {continue}
			asciiChar := buf[0]
			if	word1 { t0 = time.Now(); word1 = false }

			char     := string(asciiChar)
			// fmt.Printf("%d", asciiChar)

			// charByte := make([]byte, 1) // read one character
			// _, err   := os.Stdin.Read(charByte); if err != nil { fmt.Println("Error reading input:", err); break }
			// char     := string(charByte)

			if	char == " " { // Check for space (end of word)
				// Process and record word
				typedStr := typedAcc.String()
				if	typedStr != "" { // accumulator was just "Reset", so at the end of word
					maxDur  := time.Duration(timeDurMult * len(typedStr) + timeDurMult) // compute threshold for time allotment
					elapsed := time.Now().Sub(t0)
					// word := Word{ Lid: tgt, AsTyped: typedStr, IsCorrect: typedStr == tgt, TimeTaken: elapsed, Esecs: st0,
					word    := Word{
						Learnable: stats.Learnable{Lname: tgt},
						Question:  stats.Question{Lid: XXX}
						Encounter: stats.Encounter{
							Entered: typedStr,	Correct: typedStr == tgt,
							Timer: elapsed,	Estamp: st0 },
					}
					// fmt.Printf("\nRead typedStr: %s\n", typedStr)
					if        typedStr == "exit" { // fmt.Printf("\nbreaking\n"); break
					} else if typedStr == "cat"  {
						fmt.Print("\b\b\b")
						fmt.Printf(ColorGreen + typedStr + ColorReset)
					} else if typedStr != tgt { // wrong
						fmt.Print(strings.Repeat("\b", len(typedStr)))
						fmt.Print(strings.Repeat(" ",  len(typedStr))) // refill with spaces to force-clear
						fmt.Print(strings.Repeat("\b", len(typedStr)))
						fmt.Print(ColorRed + tgt + ColorReset)
						// maybe add to slow words now (of after training over)
					} else if time.Millisecond * maxDur < elapsed { // too slow
						fmt.Print(strings.Repeat("\b", len(typedStr)))
						fmt.Print(ColorYellow + tgt + ColorReset)
					} else {
						fmt.Print(strings.Repeat("\b", len(typedStr)))
						fmt.Print(ColorGreen + tgt + ColorReset)
					}
					fmt.Printf(" ")
					*words = append(*words, word)
				}
				typedAcc.Reset() // Clear for the next typedStr
				break
			} else if asciiChar == 127 {
				fmt.Print("\b")  // backspace
				s := typedAcc.String()
				if 0 < len(s) {
					typedAcc.Reset()
					s = s[:len(s)-1]
					typedAcc.WriteString(s)
				}
				// fmt.Printf("\nDetected key: %c (ASCII %d)\n", asciiChar, asciiChar)
			} else if asciiChar == 9 {
				// TODO maybe skip word? Interesting feature
				fmt.Print("TAB")
			} else if asciiChar == 8 { // Ctrl-backspace
				// errase whole word
				s := typedAcc.String()
				if 0 < len(s) {
					s := typedAcc.String()
					fmt.Print(strings.Repeat("\b", len(s)))
					typedAcc.Reset()
				}
			} else {
				typedAcc.WriteString(char)
				// somehow being recognized with C-backspace
				if char == "\b" { fmt.Print("XXX") } // can't detect backspace since in terminal "cooked" mode
				fmt.Print(char)
			}}}
	fmt.Print(ColorReset + "\n")
	term.Restore(int(syscall.Stdin), oldState)
}

func main4() {
	// TODO re-enable reading words from stdin
	// if len(os.Args) < 2 {fmt.Fprintln(os.Stderr, "Must pass an input string of words"); os.Exit(1)}
	nqtns := int64(10)

	ctx := context.Background()
	dbPath := "./ttypist.db" // FIXME be smarter about location
	var err error
	db, err = sql.Open("sqlite", dbPath)

	if err != nil { panic(err) }
	defer db.Close()

	st0 := time.Now()
	dbq := stats.New(db)
	// training := stats.Training{}

	// fmt.Println(os.Args)
	tgtsStr := os.Args[1]
	// fmt.Printf("%T", tgtsStr)
	tgts := strings.Fields(tgtsStr)
	// fmt.Println(tgts)
	// TODO Should maybe move words fully into conductTestLine and not be a ptr
	low, hi := int64(1), int64(20)
	words, err := dbq.GetQuestionsInBand(ctx, stats.GetQuestionsInBandParams{
		Low: &low, Hi: &hi, Bandsize: 200})
	// words := []Learnable{}
	// fmt.Println(words)

	// words = append(words, Learnable{"stoner", "stnr", false, time.Now().Sub(sessionT0)})
	// fmt.Println(words)

	// readIncoming()

	fmt.Println("Enter words (type 'exit' to quit):")
	// defer term.Restore(int(syscall.Stdin), oldState) // Ensure terminal is restored on exit

	fd := int(os.Stdin.Fd())
	termWidth, _, err := term.GetSize(fd)
	if err != nil { fmt.Printf("Error getting size: %v\n", err) }

	// wpl := 15 // words per line
	wpl := termWidth / 7 // words per line
	fmt.Printf("wpl: %d, width: %d\n", wpl, termWidth)
	nLines := int(math.Ceil(float64(len(tgts)) / float64(wpl)))
	for i := range nLines {
		eor := min(i*wpl+wpl-1, len(tgts))
		conductTestLine(&words, tgts[i*wpl:eor])
	}
	// conductTestLine(&words, tgts[0:9])

	// fmt.Println(ColorReset)
	// term.Restore(int(syscall.Stdin), oldState)

	printTypos(words)

	training, err := dbq.CreateTraining(ctx, stats.CreateTrainingParams{
		Tstamp: st0, Nqtns: nqtns, Speed: 42.0, Accy: 69.0})
	fmt.Println("training: ", training)

	fmt.Println("\n", words)
	fmt.Println("\n\nExiting.")
	fmt.Println("foo\tbar\tbpd")
	fmt.Println("foo\tbar\tbpd")
}
