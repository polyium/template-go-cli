# Architecture & Specification

## Overview

The following document ranges anywhere from high-level to code-implementation-level of detail, and is intended
to serve as an outline for `polyforge`'s specification.

### Summary

`polyforge` is a `go` CLI templating utility for producing and consuming repository templates. A repository is
considered *"polyforge-aware"* if it contains a root-level `.poly-template` descriptor. `polyforge` renders both file
names and
file contents using Go’s [`text/template` engine](https://pkg.go.dev/text/template) (with configurable
delimiters/prefix).

- Language: Go (latest stable).
- Distribution: Homebrew (macOS), standalone binaries for Linux/Windows.
- Name: `polyforge`.

### Key Concepts

#### Repository "Modes"

A `.poly-template` declares a `type` discriminator:

- `producer`: A repository, or subdirectory[^1], that defines templates to be rendered elsewhere.
- `consumer`: A repository where templates are applied (rendered into the working tree).

A repository can contain both producer and consumer subtrees[^1] (via subdirectory `.poly-template` files).

#### Template Application Surfaces

- Filename templating (optional): supports template delimiters or a prefix style.
- Examples:
  - `{{example-key-name}}_test.go`
  - `{{ custom-prefix.unique-key-name }}.md`
- File content templating[^2]: file-systems, and therefore including repositories, can be represented as a
  *Directed-Acyclic-Graph* (DAG).
  Therefore, all files and directories (nodes) are "walked" and evaluated for templating for Producers. Using a DAG with
  deterministic
  filepaths implies unique keys where each key is the relative file-system descriptor.

#### Template Inheritance (Compounding)

Templates may declare a parent chain. For example, a `bash` base template acts as parent to a `python` specialized
template. Inheritance supports additive/override merges for:

- Variables (defaults and prompts)
- Include/exclude patterns
- Hook steps
- Functions/helpers
- Delimiters/prefix settings (child overrides parent)

#### Official Template Registry

`polyforge` became a tool out of necessity; as such, Polyium provides an official "registry" of curated templates. The
CLI can:

- List and search official templates
- Inspect and fetch a specific version
- Integrate official templates into existing repos/directories with minimal friction

### Installation and Runtime

#### Homebrew

```bash
brew tap polyium/tap

brew install polyium/tap/polyforge
```

#### Supported OS

- macOS (Homebrew)
- Linux (static binary tarball)
- Windows (zip)

#### Configuration Precedence

1. CLI flags
2. Environment (`POLYFORGE_*`)[^4]
3. `.poly-template` in the active directory (or targeted path)
4. User configuration file (`~/.config/polyforge/.polyrc[.json]`)[^5]

### JSON Schema

`polyforge` produces and validates JSON Schemas for `.poly-template`. Can be added to IDEs including
any JetBrains product and VSCode.

### Safety, Idempotency, and Git

- Collision Policies: skip|prompt|force|backup (default prompt).
- Symlink & Path Traversal: disallow traversing outside SOURCE root; follow symlinks only if --follow-symlinks.
- Permissions: preserve executable bit; normalize LF endings unless --keep-eol.
- .gitignore awareness: honor .gitignore by default; --no-gitignore to override.
- Dry Runs & Diffs: --dry-run produces a file list and unified diff without touching disk.

### Extensibility

- Hooks: pre and post hooks run in a controlled sandbox (env has resolved variables; POLYFORGE_* exposed).
- Helpers/Funcs: a stable, versioned set of extra template funcs (e.g., slug, pascal, kebab, snake, now, trim,
  default).
  - Implementation: small, internal library to avoid heavy dependencies.
  - Future: optional --enable-sprig for power users (clearly marked "experimental").
- Registry Providers: official://name@ver, git://url#ref, file://path. Checksums required for official; optional
  signing (cosign) is a later milestone.

### Observability & UX

- Logging: stderr by default; --json outputs structured events.
- Levels: error, warn, info, debug, trace.
- Key checkpoints: config discovery, parent chain, variable resolution, file match sets, hook exec, write outcomes,
  collisions.
  - Metrics/Telemetry: disabled by default; opt-in via config. If enabled, only aggregate counts (no PII).
  - Timestamps: ISO-8601 in logs; monotonic time for durations.

### Performance

- Parallel file walking with bounded concurrency (configurable; default based on CPU).
- Stable deterministic ordering before write (sort by path) to guarantee reproducible runs.
- Local materialized parent trees (content-addressed by checksum).

### Testing & CI

- `polyforge` test:
  - Executes a dry-run render with sample vars, validates against golden outputs.
  - Verifies hooks exit codes and logs.
  - Provide a GitHub Action:
  - Validate .poly-template
  - Render with --dry-run --json
  - Fail CI on schema/variable mismatch or drift.

## Footnotes

[^1]: Planned support for `.poly-template` subdirectory overrides and related scoping behaviors.
[^2]: Planned support for arbitrary file-content rendering without `.tmpl` file extensions.
[^3]: Planned support for producer-template versioning via `git` tagging.
[^4]: Planned support for runtime configuration options via `POLYFORGE_` environment variables.
[^5]: Planned support for `.polyrc[.json]` user configuration.
