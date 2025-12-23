---
trigger: always_on
---

This repository follows a strict commit discipline to keep history readable, reviewable, and revert-friendly.  
Commits must be **small, atomic, and meaningful**—each commit should represent **one logical change**.

## Core Principles
- **Atomic commits:** one intent per commit (feature OR fix OR refactor; do not mix unrelated changes).
- **Readable history:** commit messages should explain **what** changed and **why**, not how. :contentReference[oaicite:1]{index=1}
- **Easy revert:** each commit should be safe to revert without breaking unrelated functionality.
- **Small scope:** prefer multiple small commits over one large commit.

### Type
Use one of the following types. :contentReference[oaicite:3]{index=3}
- **feat**: add a new feature
- **fix**: bug fix
- **docs**: documentation only
- **style**: formatting only (no logic change)
- **refactor**: code restructuring (no behavior change)
- **test**: add or update tests
- **chore**: build tooling, dependency updates, configs, maintenance tasks

