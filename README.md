# TTYpist — the simplest possible terminal-based typing tutor/speed test

## Synopsis

```shell
    ttypist --count 100 --pattern '(tt|cc|con|tion|ment|able|ject|ould|ight|ment)' --level 3
```

Zyping is the simplest possible terminal-based typing tutor/speed test.

## Features

- Simplest possible UI
- Tiny Zsh script you can change
- Side-by-side formatting of missed words
- Metrics: accuracy, WPMs (actal and raw), penalties, counts, timings
- Graceful early quitting
- Sophisticated regex-based training set selection
- Retrain on missed words (soon)
- Repeat test
- Highly configurable: number of words, pool size, regex patterns, penalties, dictionary
- Scriptable: returns timing on stderr as input to run in loop
- Scriptable: feed a list of words on stdin (as file: `ttypist <(somecmd ...)`

## Demo

```shellsession
% ~/proj/zyping/bin/zyping

  being before among whether all all and question name six form someone to real
  food mind face evidence large big put public has keep think say economic every
  a often

Start typing to begin test, <enter> to end.

> being before among whether all all and quetion name gix form someone to real food mind face evidence lange big put budlic has keep think say economic every a forme

quetion -> question
gix     -> six
lange   -> large
budlic  -> public
forme   -> often

Missed words: large often public question six

Test of 30 words took 31 seconds.
WPM: 54.7 (raw: 63.5)
Acc: 83% (25/30)

Type these missed words (free-form, as many times as you like):

  often large often often often question six question often public large public
  six public question question six public six large large public large question
  six

> often lange often often often quetion six quistion ofret public large publi csix ...
```

## Why

This started as a small, fun exercise to see what could be prototyped in a
small script and maybe eventally become a clone of [ttyper]. But as it started
taking shape, it became clear that I could do most of what ttyper does in ways
that I prefer.

## Recipes

Keybr: `rg '\t[eniarl]+$' 3-10k.num`

Then keep adding to build up to full sequence `eniarl tosudycghpmkbwf

### Top-200

```shell
poolsize=200
```

### British spellings

`TTYP_DICT=mybrit.txt`

### Prose

```shell
TTYP_NOSHUF=1 ttypist mypoem.txt
```

## Environment

- `TTYP_NWORDS` — number of words in test (default `30`)
- `TTYP_NOSHUF` — turn off random shuffling of words
- `TTYP_POOLSIZE` — top-N words to choose from; ex: `50000` means almost everything (default `200`)
- `TTYP_PATTERN` — regex selector (quote it!) for words in test (default `.` means all)
- `TTYP_DICT` — file to use as dictionary input
- `TTYP_PENSECS` — time penalty in seconds; miss 3 words -> 3s (default `1`)


## Inspiration

- [monkeytype]()

- [keybr](https://www.keybr.com/)

- [ttyper](https://github.com/max-niederman/ttyper) — has some bugs and panics
  on custom input; submitted bug but found many similar ones not looked at in
  over a year
