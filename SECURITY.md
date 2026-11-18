# Security policy

## Reporting a vulnerability

Please report security problems privately, not in a public issue.

Use GitHub's private reporting form: <https://github.com/Gitguy268/spot-cli/security/advisories/new>

Include what you found, how to reproduce it, your OS, shell and terminal. You should get a reply within a week. Please give the fix time to ship before you publish details.

## Scope

spot runs from your shell prompt after every command, so the most useful reports are about:

- a command line that makes the hook script (`spot init …`) run something it shouldn't, or break your shell;
- a crafted `state.json` or `config.toml` that crashes spot or gets text into your terminal unfiltered;
- `spot run` not passing a command through exactly as given, or losing its exit code;
- the GitHub Actions workflows in this repository.

## Supported versions

Only the latest release and the `main` branch are supported.
