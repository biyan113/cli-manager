# CLI Manager

[![CI](https://github.com/biyan113/cli-manager/actions/workflows/ci.yml/badge.svg)](https://github.com/biyan113/cli-manager/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A cross-platform desktop app for installing and maintaining CLI tools directly from verified GitHub Release assets.

[简体中文](README.zh-CN.md)

## Highlights

- Install, update, downgrade, and uninstall CLI tools with live progress.
- Verify every download against an official SHA-256 checksum before installation.
- Detect installed versions and compare them with the latest stable release.
- Add repositories from a GitHub URL and customize asset naming rules.
- Generate bilingual tool summaries with an optional DeepSeek API key.
- Switch between system language, Simplified Chinese, and English.
- Keep GitHub and DeepSeek tokens local; secrets are never returned to the UI.

## Curated tools

The first launch includes release patterns verified for these projects:

| Tool | Purpose | Platforms |
| --- | --- | --- |
| [asc](https://github.com/rorkai/App-Store-Connect-CLI) | App Store Connect automation | macOS, Linux, Windows |
| [gh](https://github.com/cli/cli) | GitHub CLI | macOS, Linux, Windows |
| [jq](https://github.com/jqlang/jq) | JSON processor | macOS, Linux, Windows |
| [yq](https://github.com/mikefarah/yq) | YAML/JSON/XML processor | macOS, Linux, Windows |
| [fzf](https://github.com/junegunn/fzf) | Fuzzy finder | macOS, Linux, Windows |

CLI Manager only supports releases that publish a downloadable executable or ZIP/TAR.GZ archive and a compatible SHA-256 checksum file. Release layouts can change upstream; customize a tool's patterns if needed.

## Requirements

For development:

- Go 1.23 or newer
- Node.js 20 or newer and npm
- [Wails v2.12](https://wails.io/docs/gettingstarted/installation)
- Platform requirements for Wails (WebView2 on Windows, WebKit on Linux)

## Development

```bash
git clone https://github.com/biyan113/cli-manager.git
cd cli-manager
wails dev
```

Run verification locally:

```bash
go test ./...
cd frontend && npm ci && npm run build
```

Build the desktop app:

```bash
wails build
```

Outputs are written under `build/bin/`. On Windows, build an NSIS installer with:

```powershell
wails build -clean -platform windows/amd64 -nsis
```

## Configuration

Configuration and state are stored under the operating system's user configuration directory:

- Windows: `%AppData%\cli-manager\`
- macOS: `~/Library/Application Support/cli-manager/`
- Linux: `$XDG_CONFIG_HOME/cli-manager/` or `~/.config/cli-manager/`

Default install directories are `~/bin` on Windows and `~/.local/bin` elsewhere. Add that directory to `PATH` if it is not already present.

Tokens are stored in `tools.json` with user-only file permissions where supported. For stronger secret isolation, leave tokens unset and use anonymous GitHub access.

## Releases

Windows packages are built by the repository's dedicated self-hosted Windows runner. Runner setup and release procedures are documented in [docs/windows-runner.md](docs/windows-runner.md). A release is not considered verified until the workflow completes on the real Windows host and its artifact is smoke-tested.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Please report vulnerabilities according to [SECURITY.md](SECURITY.md), not in a public issue.

## License

[MIT](LICENSE) © 2026 biyan
