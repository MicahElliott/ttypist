# Ttypist Design

I'm building a typing tutor in Golang that is a simple TUI that is unique in several
ways:
 
- its ability to **react to terminal input** for typical typing needs without
  getting confused about where in the line it is and how to move around a bit:
  backspace, delete-backward-word (`C-w`), RET to go to new line, `C-c` to
  exit, etc (but not arrow keys)

- do **timing** of things, particularly a single word (and also a session of
  course -- what else?)

- **color words** according to timing and correctness

- keep cursor just below the focus word being attempted (other tutors fail at
  this but web-based are often good at it)
  
- work well in almost any terminal

- track sessions and show progress stats

It is based on prior work done in Zsh in repo `~/proj/zyping/bin/ttypist` (and
its sister scripts in same dir) and `~/proj/zyping/README.md` (read that as a
resource for a lot of details). A sister script `ttypist-workup` is
particularly effective in accomplishing a lot of good utility with very little
code, and other scripts there do a set of features on par with most other
tutors). I use zyping regularly and it works well. But with zsh I wasn't able
to figure out how I could possibly do individual word timing with the way its
loops work. So I decided to recreate much of it in Go, which is this Ttypist
project.

I think the simplest way to make Ttypist work well is by displaying the
complete target paragraph up front, then typing through an input paragraph
that uses the same word breaks. The target text stays static; only the current
input line is redrawn, and a new input line is appended when the last word on
the current paragraph line is committed.

## Historical seed

The initial Go prototype was committed as `tt3.go`. The active entry point
now lives in `main.go`; the original seed remains recoverable from Git with:

```sh
git show e9241bc:tt3.go > tt3.go
```

## Example session in zyping

```
% ./bin/ttypist
╭────────────────────────────────────╮
│    TTYpist Typing Session #324     │
╰────────────────────────────────────╯
POOLBAND: 1-200 | NWORDS: 50 | MINWPM: 50 | MINACC: 92 | DICT: 10k-3.num | PATTERN: .

Start typing to begin test, <enter> to end.

  always man good same from going most after made again small which day first
  next been one than one great yeah his why each although system government know
  now would

> always man good same from going host after made again small which day ff next been one than one great yeah his why each although systeh goevrnment know now would 

didi  -> did
ewill -> will
year  -> yeah

Test of 50 words took 40 seconds.
WPM: 73.7 (raw: 79.2)
Acc: 94%  (47/50)

Type these missed words (untimed free-form, as many times as you like):

  will yeah did will will will will yeah will will will yeah yeah yeah yeah yeah
  yeah will did yeah yeah did did will yeah will will yeah
```

### How ttypist should become

```
% ./bin/ttypist
╭────────────────────────────────────╮
│    TTYpist Typing Session #324     │
╰────────────────────────────────────╯
POOLBAND: 1-200 | NWORDS: 50 | MINWPM: 50 | MINACC: 92 | DICT: 10k-3.num | PATTERN: .

Start typing to begin test, <enter> to end.


  always man good same from going most after made again small which day first

> always man good same from going most after made again small which


  day first next been one than one great yeah his why each although system 

> day first next been one than one great yeah his why each


  although system now would
  
> although system now would
```

Stats in zyping (using go https://github.com/guptarohit/asciigraph):

```
% ./bin/ttypist-stats 
Sessions (last 50):
 82.70 ┤                                             ╭╮
 73.52 ┤                          ╭─╮             ╭──╯│
 64.34 ┤       ╭╮   ╭──╮╭──────╮  │ ╰╮ ╭╮  ╭──╮   │   │ ╭
 55.16 ┼───────╯│ ╭─╯  ╰╯      ╰──╯  ╰─╯╰──╯  ╰╮╭─╯   ╰╮│
 45.98 ┤        ╰─╯                            ││      ╰╯
 36.80 ┤                                       ╰╯

Hardest words (top-20):
     13 mean
     12 first
     12 different
     11 themselves
     11 government
     11 come
     11 children
     10 system
     10 scheme
     10 most
     10 health
     10 ever
      9 where
      9 today
      9 political
      9 one
      9 become
      9 because
      9 actually
      8 programme
```

## Other typing tutors and work to mimic

I've liked the way monkeytype, keybr, and ttyper work. Their input UIs are
good models to consult.

In repo `../zyping/goproj` I believe I had some code working to fire up a
menu-driven system (totally incomplete) for selecting from learning modules to
study. That project was half-baked, but did a lot to show how I might want
ttypist to work. Look at its code to do similar menu driving with
tcell/tvxwidgets/tview etc, and use those libs -- see its `go.mod` file.
