---
name: feedback_make_commands
description: Use make run/make dev to start the server, not manual go run or binary builds
type: feedback
---

Use `make run` or `make dev` to start the server, not manual `go run` or custom shell commands. The Makefile handles environment loading, port checks, and infrastructure startup.

**Why:** The user expects the same workflow they use locally. Manual commands skip env setup and are harder to follow.

**How to apply:** When testing server startup or verifying the app works, use the existing Makefile targets.
