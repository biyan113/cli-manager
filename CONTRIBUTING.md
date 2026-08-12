# Contributing

Thanks for helping improve CLI Manager.

## Before opening a change

- Use an issue for behavior changes that affect configuration compatibility or security.
- Never include API tokens, runner registration tokens, signing certificates, or personal configuration files.
- Keep release patterns tied to official upstream GitHub Releases and require SHA-256 verification.

## Local checks

```bash
gofmt -w app.go main.go internal/**/*.go
go test ./...
go vet ./...
cd frontend
npm ci
npm run build
```

Add focused tests for configuration migration, asset matching, checksum parsing, and archive extraction when those areas change.

## Pull requests

- Keep changes focused and use Conventional Commits.
- Explain supported platforms and the upstream release layout for a new default tool.
- Update both READMEs for user-visible behavior.
- Distinguish source/build validation from real Windows packaging and smoke testing.
