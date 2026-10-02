# Ttypist MVP

Status: draft

Ttypist is a local, single-user terminal typing tutor. Its first release should
replace the main `zyping/bin/ttypist` workflow while adding reliable per-word
timing and a paragraph-aligned, line-oriented display.

## Product boundary

The MVP includes:

- one interactive typing session over a selected sequence of words;
- terminal-width rendering with a static target paragraph and aligned input;
- printable input, Backspace, `Ctrl-W`, `Ctrl-C`, and clean terminal restore;
- per-word elapsed time, correctness, slow-word coloring, and session timing;
- WPM, raw WPM, accuracy, missed-word pairs, and a post-session missed-word
  practice round;
- dictionary selection by rank range, regular expression, custom dictionary,
  and supplied text;
- local persistence of sessions and attempts;
- a stats command for recent sessions and hardest words;
- command-line configuration and useful exit status for shell loops.

The MVP does not include a menu-driven course system, adaptive strength
calculation, per-letter statistics, mouse support, multiple user profiles, or
the `keybr`, `workup`, `kbpatts`, and `longest` activities. Those activities
should be able to use the same session engine later.

## Interaction contract

The exercise is a single logical sequence of target words. The renderer wraps
the complete target paragraph before typing begins and uses those same word
breaks for the input lines:

```text
  when this made while from however
  some between now world

> when this made while from however
  some between now world
```

The target paragraph remains unchanged. The current input line is redrawn as
the user types; after the final word on a line is committed, the renderer
appends exactly one new input line. A word is never split across lines, and the
engine remains independent of terminal escape sequences.

Space commits the current word. A committed word advances the focus and
records its result. A word is correct when its entered text exactly matches the
target. The target is shown in the result color after commitment, then the
next word becomes active.

The timer starts with the first printable character of a word and stops when
the word is committed. Backspace and `Ctrl-W` edit the current word and do not
reset its timer; time spent correcting an entry is part of typing performance.
The session timer starts with the first printable character and ends when the
session completes or the user quits.

Return is ignored during a session. `Ctrl-C` exits cleanly and restores the
terminal. A completed session is the configured number of words, or the words
supplied by a finite custom source. An early quit is saved as an interrupted
session when it contains attempts.

## Timing and metrics

The slow-word threshold is configurable. The default is 150 milliseconds per
target rune. The CLI
should also support a target-WPM setting that derives the threshold from the
target word length. The selected threshold and timing mode belong in the saved
session metadata.

Session metrics are:

- elapsed session time;
- raw WPM based on target characters, including spaces between words, divided
  by five;
- accuracy as correct committed words divided by committed target words;
- penalized WPM, using a configurable penalty per incorrect word;
- correct, incorrect, and unattempted counts.

Incorrect words are reported as `entered -> target`. Slow correct words receive
their own color so correctness and speed remain distinguishable.

## Configuration

The initial CLI should support equivalent capabilities to `zyping`:

```text
ttypist run
  --nwords 30
  --pool 1-200
  --pattern .
  --dict path/to/words
  --input path/to/text
  --seed 1234
  --slow-ms-per-rune 150
  --target-wpm 50
  --penalty-seconds 1
  --min-wpm 50
  --min-accuracy 92
```

The CLI generates a random seed when `--seed` is omitted. Pass a fixed value,
such as `--seed 1234`, when a repeatable word sequence is useful for debugging
or practice.

The command uses GNU-style long options with short aliases for common flags.
Generate shell completion and a man page with:

```sh
make completion-zsh
source ttypist.zsh
make man
man ./ttypist.1
```

The default word list is the existing `10k-3.num` data file embedded in the
binary. A custom dictionary or input source can override it. Editing the
default word list is outside this MVP.

## Persistence and stats

The MVP uses an append-only JSONL data store at
`$XDG_DATA_HOME/ttypist/sessions.jsonl`, falling back to
`~/.local/share/ttypist/sessions.jsonl`. It records session configuration and
outcome, plus each attempted word, entered text, correctness, timing, and
activity metadata. The schema should leave room for later courses and
activities without requiring those features now.

`ttypist stats` should show recent session results and the most frequently
missed target words. A compact built-in terminal graph is sufficient; an
external plotting dependency is not required for the first release.

After a completed session, the summary shows each miss as entered text followed
by the target, for example:

```text
didi  -> did
ewill -> will
year  -> yeah

Test of 50 words took 40 seconds.
WPM: 73.7 (raw: 79.2)
Acc: 94% (47/50)
```

When timing crosses the configured slow-word threshold, the summary also
shows a yellow `Slow` section with each affected word and its elapsed seconds.
Incorrect words use `target/entered` in that section; correct words use the
target text. Slow correct words remain visibly distinct from incorrect words
while typing.

The missed-word practice round presents unique target words repeatedly in a
free-form line. It is untimed, accepts the same editing keys, ends on Return or
`Ctrl-C`, and reports how many entered words matched a missed target.

## Executable acceptance scenarios

These scenarios describe tests of the session engine. They should use a fake
clock and synthetic key events so timing tests do not sleep.

### Basic completion

Given targets `one two`, when the events are `o n e Space t w o Space`, the
engine records two correct attempts and completes the session.

### Per-word timing

Given a clock and targets `one two`, advancing the clock separately between
the first printable character and each Space records two independent durations
and a correct session duration.

### Editing

Given target `same`, the events `s a m x Ctrl-W s a m e Space` record entered
text `same`, include the whole correction interval in the duration, and mark
the attempt correct.

### Incorrect word

Given target `which`, the events `w h o Space` record entered text `who`, mark
the attempt incorrect, and advance to the next target.

### Paragraph layout

Given a terminal width, the renderer wraps all target words once, prints the
target paragraph before input begins, and uses the same word ranges for the
input lines without duplicating attempts.

### Quit safety

Given an active session, `Ctrl-C` returns an interrupted result and the terminal
adapter restores the original terminal mode.

### Return handling

Given an active session, Return does not commit a word, advance the prompt, or
end the session.

### Deterministic selection

Given the same dictionary, selection configuration, and seed, two sessions
produce the same target sequence.

### Missed-word practice

Given a completed session with missed targets, the renderer presents those
targets repeatedly, accepts free-form input without affecting session timing,
and ends on Return or `Ctrl-C` with a practice result.

## Implementation sequence

Complete these as small vertical slices. Each slice should leave the program
buildable and should add or update its acceptance tests.

- [x] Establish the Go module, timing/session engine, paragraph planner,
  and append-only terminal renderer.
- [x] Add the embedded default dictionary and deterministic word selection.
- [x] Add CLI configuration for word count, pool, pattern, custom input, seed,
  timing, penalties, and completion thresholds.
- [x] Add session and per-word persistence under the XDG data directory.
- [x] Add the post-session missed-word practice round.
- [ ] Add `stats` with recent sessions and hardest words.
- [ ] Add PTY coverage for completion, `Ctrl-C`, terminal restoration, and
  prompt advancement.
- [ ] Recheck the release gate and document the finished command examples.

The keybr, work-up, keyboard-pattern, and longest-word activities come after
the MVP release gate and reuse the same session engine.

## Release gate

The MVP is ready when a user can run a normal session from a fresh checkout,
complete it in a typical 80-column terminal, see the static target paragraph and
per-word colors, review metrics, practice misses, restart the program, and see
the resulting session in `stats`. The engine tests pass, and a PTY smoke test
verifies terminal restoration on completion and `Ctrl-C`.

## Open decisions

- Whether the first public CLI should accept `ttypist` as shorthand for
  `ttypist run`.
- Whether the stats graph should plot accuracy, penalized WPM, or offer both.
