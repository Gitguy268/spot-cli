# Contributing

Thanks for helping with spot. Spot thanks you too, with his whole back half.

By taking part you agree to the [Code of Conduct](CODE_OF_CONDUCT.md).

## Reporting a bug or asking for a feature

Open an [issue](https://github.com/Gitguy268/spot-cli/issues/new/choose) and use the template. For a bug, say which part (shell hook, `spot run`, the room), your shell, OS and terminal, and what you ran. A screenshot helps with anything that looks wrong.

A test runner Spot doesn't recognise is a good first pull request: add it to `Defaults` in `internal/hook/match.go` and a line to the table in `match_test.go`.

Report security problems privately: see [SECURITY.md](SECURITY.md).

## Making a change

1. Fork the repository and create a branch from `main`.
2. Build and run it. You need Go 1.25 or newer.

   ```sh
   go build -o spot ./cmd/spot
   export SPOT_HOME=$(mktemp -d) SPOT_CONFIG=$(mktemp -d)   # a throwaway Spot
   ./spot
   ```

   `SPOT_HOME` and `SPOT_CONFIG` keep test runs away from your own Spot.

3. Make your change. Keep commits small, each with a clear message.
4. Run the checks below, and update the README if behaviour changed.
5. Open a pull request and fill in the template.

## Checks

```sh
gofmt -l .                              # prints nothing
go vet ./...
go test -race ./...
python3 tests/shell_smoke.py ./spot     # real zsh, bash and fish through a pty
go run ./internal/cmd/frames            # then git diff --exit-code site/
```

## Drawing Spot

Sprites are grids of letters in `internal/art`. Each letter is a colour in `palette.go` (`K` fur, `O` the rim light that keeps a black dog visible on a black terminal, `R` his collar, `G` the tag), and `.` is see-through. Most poses are the sitting grid in `spot.go` with small patches on top.

To look at your work:

```sh
go run ./internal/cmd/frames -png /tmp/spot -scale 10 -only wag,mini-wag
open /tmp/spot/dark.png /tmp/spot/light.png
```

Spot is a real dog, so please keep him recognisable: a black lab with floppy ears, warm eyes and a red collar with a gold tag. The tests check that the collar and tag are there.

## Layout

| Path | What |
| --- | --- |
| `cmd/spot` | The `spot` command |
| `internal/pet` | Spot's stats, moods, reactions, tricks and badges. Pure functions of state and time |
| `internal/store` | `state.json` (atomic writes and a lock) and `config.toml` |
| `internal/hook` | Which commands are tests, and the shell scripts `spot init` prints |
| `internal/react` | The prompt face and the reactions in the shell |
| `internal/tui` | The room |
| `internal/art` | Sprites, palettes and the half-block renderer |
| `site/` | The web page, which plays the same frames from `frames.js` |
| `docs/tapes` | VHS tapes for the README pictures (`vhs docs/tapes/hook.tape`) |
