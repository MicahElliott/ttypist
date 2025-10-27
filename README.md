# Zyping

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

## Why

This started as a small, fun exercise to see what could be prototyped in a
small script and maybe eventally become a clone of [ttyper]. But as it started
taking shape, it became clear that I could do most of what ttyper does in ways
that I prefer.

## Inspiration

- [ttyper]() — has some bugs and panics on custom input; submitted bug but
  found many similar ones not looked at in over a year
