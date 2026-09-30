# AGENTS.md

Guidance for AI coding assistants working in this repository.

## Shared rules

At the start of each session, read `~/.config/ai-rules/README.md`,
`~/.config/ai-rules/global.md`, and `~/.config/ai-rules/INDEX.md`. Then read
only indexed rules relevant to `~/proj/ttypist` and the current task. If the
shared directory is unavailable, continue with this file and mention it.

Preserve existing user changes. Do not commit or push unless explicitly asked.
Use `apply_patch` for edits, keep diffs focused, and run the relevant Makefile
checks after changes.
After each implementation step is verified, print a detailed, pasteable commit
message for that step. When confidence is high, it is also fine to provide the
message speculatively before final confirmation. Do not commit automatically.

## Project intent

Ttypist is a small Go terminal typing tutor that recreates the useful core of
`../zyping/bin/ttypist` while adding reliable per-word timing, rolling prompts,
and local progress statistics. Read `design.md` for the original vision and
`MVP.md` for the current scope and acceptance contract.

Keep dependencies and implementation small. Treat `MVP.md` and its executable
tests as the source of truth when behavior evolves.

When asked to continue building the MVP, inspect the current status and take
the next unchecked item under `MVP.md`'s Implementation sequence. Finish that
vertical slice, update its tests and checklist, run `make check`, and report
the next remaining item. Implement all remaining items only when the user
explicitly asks for an end-to-end MVP pass.

## Interaction rules

- Treat the exercise as one logical sequence of target words.
- Fit each prompt to terminal width. Show one current target line and its input
  line; do not reveal several future prompt lines.
- Each prompt has a body plus two lookahead words by default. Lookahead repeats
  visually in the next prompt but creates no duplicate attempts.
- Leave completed target/input pairs in terminal scrollback.
- Space commits a word. Return is ignored. Backspace removes one character;
  `Ctrl-W` clears the current uncommitted word; `Ctrl-C` exits cleanly.
- Start a word timer with its first printable rune and include correction time.
  Target-WPM timing overrides the default 250 ms per target rune when set.
- Never split a word across lines or accept a word wider than the available
  prompt width.
- Renderer line breaks must be `\r\n`; a bare `\n` can preserve the terminal
  cursor column and indent subsequent scrollback lines. Keep a regression test
  for line-oriented output when changing the renderer.

## Build and verify

```sh
make run WORDS='one two three'
make check                 # formatting check, tests, and vet
make build
make race
```

Use a real terminal or PTY for interactive checks. The Makefile handles the
local `-buildvcs=false` build requirement. Keep `go.mod` and `go.sum` current
when dependencies change.

Between renderer iterations, run at least one real typing session through a
PTY, including the relevant editing or quit path. Use engine tests for exact
timing and layout assertions; use the PTY session to catch cursor, redraw, raw
mode, and scrollback problems.

The main implementation is currently in `engine.go`, `renderer.go`, and
`tt3.go`; their behavior is covered by `*_test.go` files.
