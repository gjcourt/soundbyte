# Copilot Instructions - Go Project

## Code Style & Conventions

- **Formatting**:
  - Always use `gofmt` (or `goimports`) to format code.
  - Line length should generally be kept reasonable (~80-100 characters), but readability is priority.

- **Naming**:
  - Use `PascalCase` for exported identifiers (e.g., `ExportedFunction`).
  - Use `camelCase` for private identifiers (e.g., `privateVariable`).
  - Acronyms should be all caps (e.g., `ServeHTTP`, `ID`, `URL`) unless at the start of a private variable (`url`).
  - Variable names should be short but descriptive (e.g., `i` for loops, `req` for *http.Request`).

- **Error Handling**:
  - Always check returned errors. Never use `_` to suppress an error unless strictly necessary and documented.
  - Return errors wrapped with context using `fmt.Errorf("context: %w", err)` when appropriate.
  - Handle errors early and return (Guard Clauses) to avoid deep nesting.

- **Comments**:
  - **All exported** types, functions, constants, and variables **MUST** have a comment starting with the identifier name.
  - Example: `// NewServer creates a new instance of Server.`
  - Package comments should be present in at least one file per package (usually the primary one).

- **Concurrency**:
  - Use channels for communication, mutexes for state synchronization.
  - Avoid global state where possible.
  - Ensure goroutines are properly managed and can be terminated (using `context.Context` or `done` channels).

- **Testing**:
  - Use the standard `testing` package.
  - Use Table Driven Tests for data transformation logic.
  - Use `t.Parallel()` for independent tests.

- **Constructors**:
  - struct literals are preferred for simple structs.
  - Use `NewName(...)` functions for complex initialization.

## Application Specifics

- **Project**: UDP Audio Streamer
- **Architecture**: Client/Server
- **Codec**: Raw PCM (S16LE, 48kHz, Stereo) for simplicity and purity.
- **Dependencies**:
  - `gopxl/beep/v2` (Playback)
  - `protocol` (Internal package for packet definitions)
  - `jitter` (Internal package for packet buffering)

## Container Strategy

- **Versioning**: Tag images with the current date in `YYYY-MM-DD` format (e.g., `2026-02-15`).
- **Deduplication**: If a tag for today already exists, append `-v2`, `-v3`, etc. (e.g., `2026-02-15-v2`).
- **Push Policy**: Push updated images for all significant changes.

## Quality Assurance

### CI quality gate

**Linting and tests must pass before creating or updating a PR.**

- Run `golangci-lint run ./...` (or `make lint`) locally — must report **0 issues**.
- Run `go test -race ./...` (or `make test`) locally — all tests must **pass**.
- The CI pipeline runs both checks on every push. Never push code that you know
  has lint warnings or test failures.
- If CI fails after pushing, fix the issues and force-push the branch before
  requesting review.

### General

- Ensure efficient memory usage (avoid unnecessary allocations in hot loops, e.g., audio processing).
