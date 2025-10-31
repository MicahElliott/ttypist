# TTYpist — the simplest possible terminal-based typing tutor/speed test

## Synopsis

```shell
    ttypist --count 100 --pattern '(tt|cc|con|tion|ment|able|ject|ould|ight|ment)' --level 3
```

TTYpist is the simplest possible terminal-based typing tutor/speed test.

## Features

- Simplest possible UI
- Clean and small Zsh script you can change (but shouldn't need to)
- Side-by-side formatting of missed words
- Metrics: accuracy, WPMs (actal and raw), penalties, counts, timings
- Graceful early quitting
- Sophisticated regex-based training set selection
- Retrain on missed words (soon)
- Repeat test
- Highly configurable: number of words, pool size, regex patterns, penalties, dictionary
- Scriptable: returns timing as exit-code (`$?`) for input to run in loop
- Scriptable: feed a list of words on stdin (as file: `ttypist <(somecmd ...)`

## Demo

```shellsession
% ttypist

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
taking shape, it became clear that it could do most of what ttyper does in ways
that I prefer, without errors.

## Recipes

### Keybr

The keybr.com site gets you to practice a small set of words containing only 6
letters (`e n i a r l`) until you reach 35 WPM, at which point it advances you
to add the next letter, `t`, and so on. TTYpist can do this too:

```shell
acc=() minwpm=35
for c in t o s u d y c g h p m k b w f z v k x q j
do   # Repeat test till sufficienly fast
     while (( $? < minwpm )); do TTYP_PATTERN='\t['eniarl$c${(j::)acc}']*'$c'['eniarl$c${(j::)acc}']*$' ttypist; done
     # Then advance to next letter
     acc+=$c
done
```

### Monkeytype default: limit to Top-200 words for speed practice

```shell
TTYP_POOLSIZE=200 ttypist
```

### British spellings, using your own dictionary

`TTYP_DICT=mybrit.txt`

### Prose

Don't scramble the words in a poem!

```shell
TTYP_NOSHUF=1 ttypist mypoem.txt
```

Or even better:

```shellsession
% TTYP_NOSHUF=1 ttypist <(fortune)
TTYpist typing session

  One small step for man, one giant stumble for mankind.

Start typing to begin test, <enter> to end.
...
```

## Environment

- `TTYP_NWORDS` — number of words in test (default `30`)
- `TTYP_NOSHUF` — turn off random shuffling of words
- `TTYP_POOLSIZE` — top-N words to choose from; ex: `50000` means almost everything (default `200`)
- `TTYP_PATTERN` — regex selector (quote it!) for words in test (default `.` means all)
- `TTYP_DICT` — file to use as dictionary input
- `TTYP_PENSECS` — time penalty in seconds; miss 3 words -> 3s (default `1`)

## Limitations

There is no per-word timing. It'll take a fancier `read` mechanism.

If you get off by a word (miss a space, etc), accuracy will get very confused.

## Inspiration

- [monkeytype]()

- [keybr](https://www.keybr.com/)

- [ttyper](https://github.com/max-niederman/ttyper) — has some bugs and panics
  on custom input; submitted bug but found many similar ones not looked at in
  over a year
