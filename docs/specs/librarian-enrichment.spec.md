# Spec: Bead Enrichment Format

**Task**: hq-qn44
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft

## Overview

This spec defines the format for bead enrichment - the "Required Reading" that Librarian attaches to beads before polecat assignment. It also specifies how polecats consume this enrichment.

---

## 1. Enrichment Structure

### Full Enrichment Format

```markdown
## Required Reading

> Enriched by Librarian on 2026-01-14 | Depth: standard | Time: 45s

### Summary
<1-2 sentence summary of what context is provided>

### Files to Read
- `path/to/file.go:45-120` - <description of relevance>
- `path/to/another.go` - <description>

### Prior Work
- **<bead-id>** (<status>): "<title>" - <what to learn from it>

### Documentation
- [<title>](<url>) - <what it covers>

### Key Patterns
- **<pattern-name>**: <brief description> (see `path/to/example`)

### Context Notes
<any additional context, warnings, or suggestions>
```

### Minimal Enrichment (Quick Depth)

```markdown
## Required Reading

> Enriched by Librarian on 2026-01-14 | Depth: quick | Time: 12s

### Files to Read
- `path/to/file.go` - Primary file to understand
- `path/to/test.go` - Test patterns

### Key Patterns
- **<pattern>**: <brief description>
```

---

## 2. Section Specifications

### Files to Read

Lists codebase files the polecat should read before starting.

**Format:**
```markdown
### Files to Read
- `<path>:<line-start>-<line-end>` - <relevance description>
- `<path>` - <relevance description>
```

**Rules:**
- Always use relative paths from rig root
- Include line numbers when specific sections are relevant
- Order by importance (most critical first)
- Limit to 10 files max (respect context budget)
- Description should explain WHY to read it

**Examples:**
```markdown
### Files to Read
- `src/auth/middleware.go:23-89` - Existing auth middleware pattern to follow
- `pkg/jwt/claims.go` - JWT claims structure used throughout
- `tests/auth/middleware_test.go` - Test patterns and fixtures
- `config/auth.yaml` - Current auth configuration schema
```

---

### Prior Work

Lists related beads (open or closed) that provide relevant context.

**Format:**
```markdown
### Prior Work
- **<bead-id>** (<status>): "<title>" - <learning/relevance>
```

**Rules:**
- Include bead ID, status (open/closed), and title
- Explain what to learn from each
- Prioritize closed beads with successful implementations
- Include open beads if they're dependencies
- Limit to 5 beads max

**Examples:**
```markdown
### Prior Work
- **gt-abc** (closed): "Add JWT refresh tokens" - Similar implementation, follow the approach
- **gt-xyz** (closed): "Auth middleware refactor" - Current middleware patterns established here
- **gt-123** (open): "Update JWT library" - Dependency, may affect implementation
```

---

### Documentation

Links to external documentation.

**Format:**
```markdown
### Documentation
- [<title>](<url>) - <what it covers>
```

**Rules:**
- Use descriptive titles (not raw URLs)
- Include brief description of relevance
- Prefer official documentation over blog posts
- Verify links are accessible (no login required)
- Limit to 5 links max

**Examples:**
```markdown
### Documentation
- [Go JWT Library](https://pkg.go.dev/github.com/golang-jwt/jwt/v5) - JWT creation and validation
- [OAuth 2.0 RFC](https://tools.ietf.org/html/rfc6749) - Protocol specification
- [Project Auth Guide](./docs/auth.md) - Internal auth architecture
```

---

### Key Patterns

Describes patterns and conventions to follow.

**Format:**
```markdown
### Key Patterns
- **<pattern-name>**: <description> (see `<example-path>`)
```

**Rules:**
- Name the pattern clearly
- Brief description of what/why
- Include reference to example in codebase
- Focus on patterns this specific task needs
- Limit to 5 patterns max

**Examples:**
```markdown
### Key Patterns
- **Error handling**: Use `pkg/errors.Wrap()` for context, never bare errors (see `src/api/handler.go:45`)
- **Middleware chain**: Add to `middleware/chain.go`, register in `cmd/server/main.go`
- **Testing**: Table-driven tests with `testify/assert` (see `tests/auth/handler_test.go`)
- **Logging**: Structured logging via `log.WithField()` (see `pkg/log/logger.go`)
```

---

### Context Notes

Free-form notes for anything that doesn't fit other sections.

**Format:**
```markdown
### Context Notes
<paragraph or bullet points>
```

**Use for:**
- Warnings about gotchas
- Historical context
- Team conventions not documented elsewhere
- Dependencies on in-progress work
- Suggestions for approach

**Examples:**
```markdown
### Context Notes
- The auth package was recently refactored (gt-xyz). Some older PRs may reference the old structure.
- Team prefers explicit error messages over generic ones. Check `errors/messages.go` for conventions.
- This feature was previously attempted in gt-old but abandoned due to scope creep. Focus on MVP.
```

---

## 3. Bead Schema Changes

### New Fields

```yaml
# In bead JSONL/DB
enrichment:
  status: pending | enriching | enriched | failed
  depth: quick | standard | deep
  enriched_at: "2026-01-14T10:00:00Z"
  enriched_by: librarian
  content: "<markdown content>"
  files_count: 5
  prior_beads_count: 2
  docs_count: 3
  patterns_count: 4
```

### Status Values

| Status | Description |
|--------|-------------|
| `pending` | Enrichment requested but not started |
| `enriching` | Librarian currently researching |
| `enriched` | Enrichment complete and attached |
| `failed` | Enrichment failed (see notes) |

---

## 4. Polecat CLAUDE.md Integration

### Enrichment Reference

Polecat CLAUDE.md should include:

```markdown
## Working with Enriched Beads

When you receive a bead, check if it has enrichment:

```bash
bd show <bead-id> --enrichment
```

If enriched, **read the Required Reading first** before starting work. This context
was gathered specifically for your task and saves you research time.

### Required Reading Protocol

1. **Files to Read**: Open and skim each file listed
2. **Prior Work**: Review listed beads with `bd show <bead-id>`
3. **Documentation**: Open in browser if you need deeper understanding
4. **Key Patterns**: Note these before implementing
5. **Context Notes**: Read for warnings and suggestions

**Do NOT skip the enrichment.** It exists because context matters.
```

### Startup Protocol Enhancement

```bash
# Step 1: Check your hook
gt hook

# Step 2: Work hooked? → Check for enrichment
bd show <bead-id> --enrichment

# Step 3: If enriched, read Required Reading first
# Then begin execution

# Step 4: If not enriched, proceed normally
# (or request enrichment if task is complex)
```

---

## 5. Example: Complete Enriched Bead

### Bead Before Enrichment

```yaml
id: gt-abc
type: task
title: "Add JWT token refresh endpoint"
status: open
priority: P2
description: |
  Add an endpoint to refresh expired JWT tokens.
  Should follow existing auth patterns.
```

### Bead After Enrichment

```yaml
id: gt-abc
type: task
title: "Add JWT token refresh endpoint"
status: open
priority: P2
description: |
  Add an endpoint to refresh expired JWT tokens.
  Should follow existing auth patterns.
enrichment:
  status: enriched
  depth: standard
  enriched_at: "2026-01-14T10:30:00Z"
  enriched_by: librarian
  files_count: 4
  prior_beads_count: 2
  docs_count: 2
  patterns_count: 3
  content: |
    ## Required Reading

    > Enriched by Librarian on 2026-01-14 | Depth: standard | Time: 38s

    ### Summary
    Context for implementing JWT refresh tokens, following existing auth middleware patterns.

    ### Files to Read
    - `src/auth/handler.go:45-120` - Existing token generation, follow this pattern
    - `src/auth/middleware.go:23-67` - Token validation logic to extend
    - `pkg/jwt/claims.go` - Claims structure, add refresh-specific fields
    - `tests/auth/handler_test.go` - Test patterns with mock tokens

    ### Prior Work
    - **gt-xyz** (closed): "Add JWT authentication" - Original JWT implementation
    - **gt-123** (closed): "Auth middleware refactor" - Current middleware architecture

    ### Documentation
    - [JWT Refresh Token Best Practices](https://auth0.com/blog/refresh-tokens-what-are-they-and-when-to-use-them/) - Security considerations
    - [go-jwt Library Docs](https://pkg.go.dev/github.com/golang-jwt/jwt/v5) - API reference

    ### Key Patterns
    - **Token generation**: Use `jwt.NewWithClaims()` with custom claims (see `src/auth/handler.go:67`)
    - **Error responses**: Use `api.ErrorResponse()` for consistent format (see `src/api/errors.go`)
    - **Testing**: Table-driven with `httptest.NewRecorder()` (see `tests/auth/handler_test.go:34`)

    ### Context Notes
    - Refresh tokens should have longer expiry than access tokens (configured in `config/auth.yaml`)
    - Store refresh token hash in DB, not the token itself (security requirement)
    - Team discussed this in gt-xyz comments - see notes about token rotation
```

---

## 6. Display Formats

### bd show (default)

```
? gt-abc · Add JWT token refresh endpoint   [● P2 · ENRICHED]
Owner: mayor · Assignee: gastown/polecats/capable · Type: task
Created: 2026-01-14 · Updated: 2026-01-14

DESCRIPTION
Add an endpoint to refresh expired JWT tokens.
Should follow existing auth patterns.

ENRICHMENT
  Status: enriched | Depth: standard | 38s
  Files: 4 | Prior beads: 2 | Docs: 2 | Patterns: 3

  Use 'bd show gt-abc --enrichment' for full Required Reading
```

### bd show --enrichment

```
? gt-abc · Add JWT token refresh endpoint   [● P2 · ENRICHED]
...

## Required Reading

> Enriched by Librarian on 2026-01-14 | Depth: standard | Time: 38s

### Summary
Context for implementing JWT refresh tokens, following existing auth middleware patterns.

### Files to Read
- `src/auth/handler.go:45-120` - Existing token generation, follow this pattern
...
```

### bd list (with enrichment indicator)

```
○ gt-abc [● P2 · ENRICHED] [task] - Add JWT token refresh endpoint
○ gt-xyz [● P2] [task] - Update error messages
○ gt-123 [● P2 · ENRICHING] [task] - Add rate limiting
```

---

## 7. Enrichment Lifecycle

```
           ┌─────────────────────────────────────────────────────┐
           │                                                     │
           ▼                                                     │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌──────┴──────┐
│   (none)    │───▶│   pending   │───▶│  enriching  │───▶│  enriched   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                         │                  │
                         │                  │
                         ▼                  ▼
                   ┌─────────────┐    ┌─────────────┐
                   │  cancelled  │    │   failed    │
                   └─────────────┘    └─────────────┘
```

---

## Implementation Notes

### Storage

Enrichment content is stored:
1. **In bead**: `enrichment.content` field (for portability)
2. **As file**: `<town>/librarian/.enrichments/<bead-id>.md` (for reference)

### Size Limits

| Field | Limit |
|-------|-------|
| Total enrichment | 10KB max |
| Files to Read | 10 entries |
| Prior Work | 5 entries |
| Documentation | 5 entries |
| Key Patterns | 5 entries |
| Context Notes | 1KB |

### Caching

Librarian may cache research results:
- File contents cached for 1 hour
- Prior bead searches cached for 5 minutes
- Web results cached for 24 hours

---

## Testing Checklist

- [ ] Enrichment format parses correctly
- [ ] All sections render in bd show
- [ ] Enrichment indicator shows in bd list
- [ ] Size limits enforced
- [ ] Polecat CLAUDE.md references enrichment
- [ ] Lifecycle transitions work correctly
- [ ] Cache invalidation works
