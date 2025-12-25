package main

// TODO
// break display of multiple lines into typing prompts
// display wrong words
// organize into fns
//

import (
	"fmt"
	// "os/exec"
	"io"
	// "errors"
	// "log"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type Word struct {
	Canon     string
	AsTyped   string
	IsCorrect bool
	TimeTaken time.Duration
}

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

func printTypos(words []Word) {
	fmt.Println("\nTypos: ...")
	for _, w := range words {
		if !w.IsCorrect { fmt.Println(w) }
	}
}

func conductTest(words *[]Word, tgts []string, termState *term.State) {
	var typedAcc strings.Builder
	timeDurMult := 300 // millis per char, or is "slow"

	for _, tgt := range tgts {
		t0 := time.Now()
		for { // Loop over every char (as it's typed)
			var buf [1]byte
			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				if err == io.EOF { break }
				panic(err)
			}
			if n == 0 {continue}
			asciiChar := buf[0]

			char     := string(asciiChar)
			// fmt.Printf("%d", asciiChar)

			// charByte := make([]byte, 1) // read one character
			// _, err   := os.Stdin.Read(charByte); if err != nil { fmt.Println("Error reading input:", err); break }
			// char     := string(charByte)

			if char == " " { // Check for space (end of word)
				// Process and record word
				typedStr := typedAcc.String()
				if  typedStr != "" { // accumulator was just "Reset", so at the end of word
					maxDur  := time.Duration(timeDurMult * len(typedStr)) // compute threshold for time allotment
					elapsed := time.Now().Sub(t0)
					word    := Word{ Canon:     tgt,             AsTyped:   typedStr,
						             IsCorrect: typedStr == tgt, TimeTaken: elapsed }
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
						// maybe add to slow words now (of after session over)
					} else if time.Millisecond * maxDur < elapsed { // too slow
						fmt.Print(strings.Repeat("\b", len(typedStr)))
						fmt.Print(ColorYellow + tgt + ColorReset)
					} else {
						fmt.Print(strings.Repeat("\b", len(typedStr)))
						fmt.Print(ColorGreen + tgt + ColorReset)
					}
					fmt.Printf(" ")
					*words = append(*words, word)
					// fmt.Print(elapsed)
					// words = append(words, Word{IsCorrect: true, TimeTaken: elapsed})
					// word.TimeTaken = time.Now().Sub(t0)
				}
				typedAcc.Reset() // Clear for the next typedStr
				// fmt.Printf("\nbreaking2", typedStr, "\n")
				break
			} else if asciiChar == 127 {
				// if asciiChar == 127 { fmt.Print("X") }
				// fmt.Print("\b")  // backspace
				fmt.Print("\b")  // backspace
				// typedAcc.WriteString(char)
				s := typedAcc.String()
				if 0 < len(s) {
					typedAcc.Reset()
					s = s[:len(s)-1]
					typedAcc.WriteString(s)
				}
				// fmt.Printf("\nDetected key: %c (ASCII %d)\n", asciiChar, asciiChar)
			} else if asciiChar == 9 {
				// maybe skip word? Interesting feature
				fmt.Print("TAB")
			} else if asciiChar == 8 { // Ctrl-backspace
				// errase whole word
				// fmt.Print("Z")
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
			}
		}
	}
	fmt.Print(ColorReset + "\n\n")
	term.Restore(int(syscall.Stdin), termState)
}

func main() {
	// words := []Word
	if len(os.Args) < 2 {
		// fmt.Errorf()
		fmt.Fprintln(os.Stderr, "Must pass an input string of words")
		// fmt.Println(errors.New("Must pass an input string of words"))
		os.Exit(1)
		// log.Fatal("Must pass an input string of words")
	}

	sessionT0 := time.Now()

	fmt.Println(os.Args)
	tgtsStr := os.Args[1]
	fmt.Printf("%T", tgtsStr)
	tgts := strings.Fields(tgtsStr)
	fmt.Println(tgts)
	// tgtsStr := os.Args
	// var words = []Word
	// TODO Should move words fully into conductTest and not be a ptr
	words := []Word{}
	// words := []Word{ {IsCorrect: true, TimeTaken: 0.0} }
	fmt.Println(words)
	// words = append()

	words = append(words, Word{"stoner", "stnr", false, time.Now().Sub(sessionT0)})
	fmt.Println(words)

	// readIncoming()

	fmt.Println("Enter words (type 'exit' to quit):")
	fmt.Println("\n  " + strings.Join(tgts, " "))
	fmt.Print("> ")

	// Get the file descriptor for standard input. This messes with printing
	// further output, so all from here needs to be handled carefully and then
	// "restored"
	oldState, err := term.MakeRaw(int(syscall.Stdin))
	if err != nil { panic(err) }
	defer term.Restore(int(syscall.Stdin), oldState) // Ensure terminal is restored on exit

	conductTest(&words, tgts, oldState)

	// fmt.Println(ColorReset)
	// term.Restore(int(syscall.Stdin), oldState)

	printTypos(words)

	fmt.Println("\n", words)
	fmt.Println("\n\nExiting.")
	fmt.Println("foo\tbar\tbpd")
	fmt.Println("foo\tbar\tbpd")
}
