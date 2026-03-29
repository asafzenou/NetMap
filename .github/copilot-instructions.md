# Repository instructions (NetMap / ReconGraph)

## Git: Smart Commit (project standard)

This repo enables **Smart Commit** in **`.vscode/settings.json`** for everyone using VS Code or Cursor with this workspace. Workspace settings override personal defaults for this project.

- **`git.enableSmartCommit`: `true`** — If nothing is staged, the commit action can still run; the editor stages changes per `git.smartCommitChanges`, then commits.
- **`git.smartCommitChanges`: `all`** — Auto-staging includes all changed files. To only auto-stage **tracked** files, use `"tracked"` in `.vscode/settings.json`.

When suggesting Git workflows, assume this behavior unless stated otherwise.

Do not tell contributors to use a private **User** `settings.json` for Smart Commit for this project — **`.vscode/settings.json`** is the source of truth.

_(The same standard is described for Cursor in `.cursor/rules/git-smart-commit.mdc`. Keep both in sync if you change the policy.)_
