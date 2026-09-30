# Ttypist MVP

Status: draft

Ttypist is a local, single-user terminal typing tutor. Its first release should
replace the main `zyping/bin/ttypist` workflow while adding reliable per-word
timing and a rolling, line-oriented display.

## Product boundary

The MVP includes:

- one interactive typing session over a selected sequence of words;
- terminal-width rendering with a two-word lookahead by default;
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

The exercise is a single logical sequence of target words. The renderer shows
one prompt window at a time. Each window contains:

1. a body of words that the user will type before the next redraw; and
2. a small lookahead, initially two words, that gives the user a view of what
   is coming next.

For example, a prompt may show:

```text
  always man good same from going most after made again small which day first
> 
```

The body ends at `which`; `day first` are lookahead words. After the user
commits `which `, the next prompt begins with `day first` and appends newly
selected words. The user continues typing without pressing Return. The
lookahead words are displayed twice across the redraw but are attempted once.

The body and lookahead are chosen to fit the terminal width. A word is never
split across lines. If the next word would flirt with the right edge, it moves
to the next prompt window. The renderer keeps the input cursor directly below
the current focus word.

Completed target/input pairs remain in terminal scrollback. When the current
body is complete, the renderer appends exactly one new target line and one new
input line below the history. It does not reveal several future prompt lines
at once. The engine remains independent of terminal escape sequences.

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

The slow-word threshold is configurable. The first default is equivalent to
250 milliseconds per target rune, matching the existing Go spike. The CLI
should also support a target-WPM setting that derives the threshold from the
target word length. The selected threshold and timing mode belong in the saved
session metadata.

Session metrics are:

- elapsed session time;
- raw WPM based on target characters divided by five;
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
  --lookahead 2
  --slow-ms-per-rune 250
  --target-wpm 50
  --penalty-seconds 1
  --min-wpm 50
  --min-accuracy 92
```

The default word list is the existing `10k-3.num` data file embedded in the
binary. A custom dictionary or input source can override it. Editing the
default word list is outside this MVP.

## Persistence and stats

The MVP uses one local data store selected through the normal XDG data
location. It records session configuration and outcome, plus each attempted
word, entered text, correctness, timing, and activity metadata. The schema
should leave room for later courses and activities without requiring those
features now.

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

The missed-word practice round presents target words repeatedly in a free-form
line. It is untimed and can be ended early.

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

### Rolling prompt

Given a terminal width and a two-word lookahead, the renderer exposes the same
logical lookahead words at the end of one prompt and the beginning of the next,
while the engine creates only one attempt for each logical target word.

### Quit safety

Given an active session, `Ctrl-C` returns an interrupted result and the terminal
adapter restores the original terminal mode.

### Return handling

Given an active session, Return does not commit a word, advance the prompt, or
end the session.

### Deterministic selection

Given the same dictionary, selection configuration, and seed, two sessions
produce the same target sequence.

## Implementation sequence

Complete these as small vertical slices. Each slice should leave the program
buildable and should add or update its acceptance tests.

- [x] Establish the Go module, timing/session engine, prompt-window planner,
  and append-only terminal renderer.
- [x] Add the embedded default dictionary and deterministic word selection.
- [x] Add CLI configuration for word count, pool, pattern, custom input, seed,
  lookahead, timing, penalties, and completion thresholds.
- [ ] Add session and per-word persistence under the XDG data directory.
- [ ] Add the post-session missed-word practice round.
- [ ] Add `stats` with recent sessions and hardest words.
- [ ] Add PTY coverage for completion, `Ctrl-C`, terminal restoration, and
  prompt advancement.
- [ ] Recheck the release gate and document the finished command examples.

The keybr, work-up, keyboard-pattern, and longest-word activities come after
the MVP release gate and reuse the same session engine.

## Release gate

The MVP is ready when a user can run a normal session from a fresh checkout,
complete it in a typical 80-column terminal, see the rolling prompts and
per-word colors, review metrics, practice misses, restart the program, and see
the resulting session in `stats`. The engine tests pass, and a PTY smoke test
verifies terminal restoration on completion and `Ctrl-C`.

## Open decisions

- Whether the first public CLI should accept `ttypist` as shorthand for
  `ttypist run`.
- Whether the stats graph should plot accuracy, penalized WPM, or offer both.
