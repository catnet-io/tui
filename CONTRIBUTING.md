# Contributing

Thank you for your interest in contributing to CatNet.

## Before you start
- Open an issue before large changes.
- Keep changes focused and small.
- Align with repository scope.
- Prefer incremental pull requests.

## Pull request checklist
- Code builds successfully.
- Tests pass locally.
- Documentation is updated when applicable.
- No unrelated files were changed.
- The PR description explains what changed and why.

## Scope rule
UI repositories must not duplicate core scanning logic.

## Branching & Commit Policy (DevSecOps)
- **Collaboration Branch**: The `develop` branch is the primary integration branch for development. All contributor pull requests must target `develop`.
- **Main Branch Restrictions**: The `main` branch is reserved for stable releases. Pull requests targeting `main` must:
  - Come exclusively from `develop`.
  - Be automatically created by `github-actions[bot]`.
- **Signed Commits**: All commits in pull requests targeting `main` must be signed (GPG or SSH signature) to ensure verification and integrity.
