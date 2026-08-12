# Windows self-hosted runner

The Windows release workflow requires a dedicated x64 runner installed on the `D:` drive with these labels:

```text
self-hosted, Windows, X64, cli-box
```

## Prerequisites

- Windows 10/11 or Windows Server 2016+
- Git for Windows
- Go compatible with `go.mod`
- Node.js 22 LTS
- WebView2 runtime
- NSIS 3 with `makensis.exe` on the system `PATH`
- Outbound HTTPS access to GitHub, Go modules, and npm

## Registration

Create the runner from the repository's **Settings → Actions → Runners → New self-hosted runner** page. GitHub issues a short-lived registration token; never commit or log it.

On the Windows host, use an elevated PowerShell session:

```powershell
New-Item -ItemType Directory -Force D:\actions-runner | Out-Null
Set-Location D:\actions-runner
# Download and verify the current runner package using the commands GitHub shows.
# Then configure it with the repository-specific URL and short-lived token:
.\config.cmd --url https://github.com/biyan113/cli-box --token <SHORT_LIVED_TOKEN> --name win-cli-box --labels cli-box --work D:\actions-runner\_work --runasservice
```

Do not reuse a runner that processes untrusted public pull requests. The packaging workflow only runs manually or for tags, and does not run on pull requests.

## Verification

1. Confirm the runner is **Idle** in repository settings.
2. Run **Actions → Windows package → Run workflow**.
3. Download `cli-box-windows-amd64` and verify `SHA256SUMS.txt`.
4. Install on a clean Windows user profile, launch the app, change the language, refresh tools, and install/uninstall one checksum-backed CLI.
5. Only create a version tag after the manual workflow succeeds.

The workflow rejects runner workspaces outside `D:` and fails if NSIS does not produce an installer.
