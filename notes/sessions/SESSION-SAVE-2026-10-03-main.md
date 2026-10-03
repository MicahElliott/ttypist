# Session save — 2026-10-03 — main

## Goal

Continue the ttypist CLI work and add Twelve-Factor-friendly environment and
optional structured configuration file support while keeping `urfave/cli` as
the command parser and source of truth.

## Completed work

- Replaced the standard library flag parser with `urfave/cli` v3.
- Added the explicit `run` command while preserving shorthand invocation.
- Added zsh completion and generated man-page commands.
- Added `TTYP_*` environment sources for all CLI settings.
- Added `--config` and `TTYP_CONFIG` for optional JSON, YAML, and TOML files.
- Defined precedence as explicit CLI flags, environment variables, selected
  config file, then built-in defaults.
- Added validation for missing, malformed, and unsupported config files.
- Added parser tests for precedence, all supported formats, invalid files,
  environment help hints, completion, and man-page generation.
- Documented the configuration behavior in `MVP.md`.

## Files changed

- `cli.go`
- `cli_test.go`
- `go.mod`
- `go.sum`
- `MVP.md`

## Decisions

- Do not add `cleanenv`; `urfave/cli` value sources and `urfave/cli-altsrc`
  provide the needed integration without a second configuration model.
- Use `TTYP_`-prefixed variables and require an explicit config path instead of
  silently searching the current directory.
- Match flat config keys to flag names, such as `nwords` and `target-wpm`.

## Verification

- `make check` passed: formatting check, tests, and vet.
- `make man` passed.
- `make completion-zsh` passed.
- A PTY session using `TTYP_SEED=123 ./ttypist one` completed successfully.
- Environment validation and missing-config error paths returned the expected
  exit status.

## Known problems and next step

No known problems remain from this session. The next unchecked MVP item is the
`stats` command for recent sessions and hardest words.

## Git state

- Branch: `main`
- HEAD: `3c1efcd feat(cli): add environment and config file sources`
- The branch is two commits ahead of `origin/main`.
- This session record is intentionally untracked until the user stages it.
