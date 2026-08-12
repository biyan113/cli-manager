# Security policy

## Supported versions

Security fixes are applied to the latest release on the default branch.

## Reporting a vulnerability

Do not open a public issue. Use GitHub's private vulnerability reporting feature for this repository. If that is unavailable, contact `gpt123@panw3i.com` with a minimal reproduction and affected version.

Please do not include live API tokens or unrelated personal data. You can expect an initial acknowledgement within seven days. No bounty or disclosure deadline is implied.

## Security model

- Downloaded tools must match SHA-256 values from the same official GitHub Release.
- Tokens are not returned through the Wails configuration API.
- A GitHub repository and its release assets remain a trust boundary; checksums do not protect against a compromised upstream publisher.
- Release artifacts from CI must not be treated as signed unless a future release explicitly documents code-signing verification.
