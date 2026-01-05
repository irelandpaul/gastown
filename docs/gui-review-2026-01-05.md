# Gas Town GUI End-to-End Review

**Date:** 2026-01-05
**Reviewer:** Codex CLI
**Scope:** GUI web app (`gui/`), bridge server (`gui/server.js`), client JS/CSS, UX, performance, and security posture for local dev usage.

## Executive Summary

Overall the GUI implementation is solid and functional, with good coverage of core workflows and thoughtful UI states. I made several hardening fixes (server binding/CORS/static exposure, input validation for polecat routes, HTTPS WebSocket support, and XSS-safe error rendering). Remaining gaps are mostly around test reliability and a few performance/security improvements that should be tracked as follow‑ups.

**Merge recommendation:** _Not fully approved yet_ because quality gates are currently failing in this environment (see below). Once test reliability issues are addressed or excluded appropriately, the changes are suitable for a PR back to `main`.

## Changes Applied in This Review (Fixed)

- **Server hardening:** Default bind to `127.0.0.1`, tightened CORS to local origins, limited static file exposure to `/assets`, `/css`, `/js`, and disabled `x-powered-by` header.
- **Input validation:** Validated `rig` and `name` parameters for polecat endpoints to prevent path traversal or malformed agent names.
- **WebSocket protocol:** Client now uses `wss://` when served over HTTPS.
- **XSS safety:** Escaped error messages inserted into HTML in issue/PR/formula lists and GitHub repo modal.

## Findings & Follow‑Ups

### High Priority

1) **Test reliability (GUI integration/e2e):** `npm test` reports timeouts waiting for UI state transitions and WebSocket data. This looks like a fixture/mock timing issue rather than a functional regression, but it blocks a clean quality gate. Tracked in #2.

2) **Test environment dependency (Go integration):** `go test ./...` fails because the beads DB is missing. Tests assume a DB exists but do not set it up or skip gracefully. Tracked in #3.

### Medium Priority

3) **Mail feed performance:** `/api/mail/all` reads the full `.feed.jsonl` file on every request before paginating. This will get slow as the feed grows. Consider incremental pagination or caching the parsed feed index. Tracked in #4.

4) **Shell execution hardening:** `executeGT`/`executeBD` are safe due to strong quoting, but switching to `execFile`/`spawn` with args would eliminate shell invocation entirely. Tracked in #5.

## Quality Gates (Local Run)

- `go test ./...` **FAILED**
  - Reason: beads DB missing (`bd list`/`bd ready`/`bd blocked` require DB).
- `npm test` (GUI) **FAILED / TIMED OUT**
  - Multiple E2E/integration tests timed out waiting for view changes or WebSocket events.

## Notes on Usability & UX

- The UI layout and interaction model are cohesive. Loading states, error toasts, and modals are consistently designed.
- The app is strongly optimized for local usage; remote exposure should remain opt‑in.

## Suggested Next Steps

1) Fix or gate the failing test suites so CI is deterministic.
2) Add pagination or incremental scanning for the mail feed endpoint.
3) Replace shell `exec` with `execFile`/`spawn` for extra hardening.
