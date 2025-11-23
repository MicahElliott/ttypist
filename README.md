# TTYpist — the simplest possible terminal-based typing tutor/speed test

## Synopsis

```shell
    TTYP_NWORDS=100 TTYP_PATTERN='(tion|ment|able|ject|ould|ight|ment)' ttypist
```

TTYpist is the simplest possible terminal-based typing tutor/speed test, which
can be scripted around for a plethora of custom features.

## Features

- Simplest possible UI
- Clean and small Zsh script you can change (but shouldn't need to)
- Clear side-by-side formatting of missed words
- Metrics: accuracy, WPMs (actal and raw), penalties, counts, timings, sessions (saved as history)
- Graceful early quitting
- Sophisticated regex-based training set selection
- Retrain on missed words (custom)
- Repeat test (custom)
- Highly configurable: number of words, pool size, regex patterns, penalties, dictionary
- Scriptable: returns timing as exit-code (`$?`) for input to run in loop
- Scriptable: feed a list of words on stdin (as file: `ttypist <(somecmd ...)`
- Emulations: Keybr, Build-up, randomized tests

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

## Typing tips

If you need to backspace more than one character, use `Ctrl-W` to delete the
whole word.

## Recipes

### Keybr

The [keybr.com] site gets you to practice a small set of words containing only 6
letters (`e n i a r l`) until you reach 35 WPM, at which point it advances you
to add the next letter, `t`, and so on. TTYpist can do this too (but much more
flexibly and permits magic keys):

```shellsession
accum=()
for c in t o s u d y c g h p m k b w f z v k x q j
do  # Repeat test till sufficienly fast
    while ! TTYP_POOLBAND=10000 TTYP_PATTERN='^['eniarl$c${(j::)accum}']*'$c'['eniarl$c${(j::)accum}']*$' ttypist; do : ; done
    accum+=$c
done
```

For a better challenge, set `minwpm=60` and `TTYP_POOLBAND=10000`, and even
`TTYP_PENSECS=3` (to force stricter accuracy).

### Build-up: Increase difficulty in sequence of sessions

```shellsession
for level in 1 100 200 300 400 500
do while ! TTYP_POOLBAND=$level-$((level+100)) ttypist; do print 'TOO SLOW'; done
done
```

### Monkeytype default: limit to Top-200 words for speed practice

```shellsession
TTYP_POOLBAND=200 TTYP_NWORDS=60 ttypist
```

### Pomodoro (10-minute)

```shellsession
tend=$(date -d '+10 minutes' +%s)
while (( $(date +%s) < tend )); do ttypist; done
```

### British spellings, using your own dictionary

```shellsession
TTYP_DICT=mybrit.txt ttypist
```

### Prose

Don't scramble the words in a poem!

```shell
TTYP_NOSHUF=1 ttypist mypoem.txt
```

Or even more fun:

```shellsession
% TTYP_NOSHUF=1 ttypist <(fortune)
TTYpist typing session

  One small step for man, one giant stumble for mankind.

Start typing to begin test, <enter> to end.
...
```

### Identify hardest words

```shellsession
rg '\tfail' ~/.local/share/ttypist/ttypist-words.log | cut -f1 | sort | uniq -c | sort -nr | head -20
```

### Practice your magic key

If you want to practice patterns you've set up as a
[magic key](https://github.com/Ikcelaks/keyboard_layouts/blob/main/magic_sturdy/magic_sturdy.md),
capture them in a regex and use like:

```shellsession
TTYP_PATTERN='(print|word|tt|cc|con|tion|ment|able|ject|ould|ight|ment)' ttypist
```

### Plot of accuracy over time

Install something like [asciigraph](https://github.com/guptarohit/asciigraph)
for terminal plotting. (There are many other CLIs that can also do this.)

```shellsession
% cut -f2 ~/.local/share/ttypist/ttypist-sessions.log | asciigraph -h 5
 67.27 ┤       ╭╮    ╭─
 63.61 ┤       ││   ╭╯
 59.95 ┤  ╭╮   ││ ╭╮│
 56.29 ┤╭╮│╰─╮╭╯│ │╰╯
 52.62 ┼╯╰╯  ╰╯ │ │
 48.96 ┤        ╰─╯
```

### Most efficient list of trigrams

I find this to be an interesting and maybe useful
[selection of practice trigrams](https://www.reddit.com/r/typing/comments/172umsd/896_trigrams_in_200_words_a_new_selection_of/).

## Environment

- `TTYP_NWORDS` — number of words in test (default `30`)
- `TTYP_NOSHUF` — turn off random shuffling of words
- `TTYP_POOLBAND` — range of dict words to choose from; ex: `1-50000` means almost everything (default `1-200`)
- `TTYP_PATTERN` — regex selector (quote it!) for words in test (default `.` means all)
- `TTYP_DICT` — file to use as dictionary input
- `TTYP_PENSECS` — time penalty in seconds; miss 3 words -> 3s (default `1`)
- `TTYP_MINWPM` — minimum WPM to complete for successful exit (default `50`)
- `TTYP_MINACC` — minimum accuracy to complete (default `95`)

Sessions are logged by default to
`~/.local/share/ttypist/ttypist-sessions.log`, word stats to
`~/.local/share/ttypist/ttypist-words.log`. These can be used to see
improvements over time, and identify words that need extra practice. Adjust
location as needed via `XDG_CONFIG_HOME`.

## Word list source

I believe (based on poor memory) that my process for generating the provided
`10k-3.num` list file was taking the 1M-word [BNC
corpus](https://www.wordfrequency.info/100k_compare.asp) and making a
rank/frequency/word TSV. It was useful for some keyboard layout analysis. You
could use any word list and change the script to `cut` any column. At 170 KB,
I didn't want to distribute bigger lists, but you may want to grab or generate
for a 50k+ list if you want a bigger `TTYP_POOLBAND`.

## Limitations

There is no per-word timing. It'll take a fancier `read` mechanism.

A test is based on a set of words, not on a timer. So there is no perfect
1-minute test. If you're aiming for 60 wpm and want a 1-minute test, use
`TTYP_NWORDS=60` to approximate it.

## Inspiration

- [monkeytype](https://www.monkeytype.com/)

- [keybr](https://www.keybr.com/)

- [ttyper](https://github.com/max-niederman/ttyper) — has some bugs and panics
  on custom input; submitted bug but found many similar ones not looked at in
  over a year
