# Changelog

All notable changes to spot are listed here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/).

## [0.1.1] - 2026-11-18

### Fixed

- bash: running the same test command twice in a row now counts both runs when `HISTCONTROL` drops duplicates (`ignoredups`, `ignoreboth`, `erasedups`; `ignoreboth` is Ubuntu's default). Before this, the second `go test ./...` of a fix was missed, and so was the comeback.

## [0.1.0] - 2026-11-18

First release.

- Spot, a black labrador with a red collar, drawn in pixel art for the terminal with half blocks: sitting, wagging, sad, hearts, eating, sleeping under a blanket, zoomies, high five, roll over, spin and more.
- Shell hooks for zsh, bash, fish and PowerShell (`spot init`). They spot test commands, add a face to your prompt and play a one-second reaction after a test run.
- `spot run -- CMD` to wrap any command. It keeps the exit code and prints nothing extra when it isn't on a terminal.
- Hunger, joy, energy and bond that change with real time and no background process. Treats for green runs, comebacks, pass streaks, five tricks and twelve badges.
- `spot`: Spot's room, with petting (keys or a mouse click), feeding, a fetch game, tricks, and live reactions to tests in other terminals.
- Colours for dark and light terminals, a 16-colour palette, and a no-colour silhouette.

[0.1.1]: https://github.com/Gitguy268/spot-cli/releases/tag/v0.1.1
[0.1.0]: https://github.com/Gitguy268/spot-cli/releases/tag/v0.1.0
