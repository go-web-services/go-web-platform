# go-web-platform

**Module:** `github.com/Lomank123/go-web-platform`

Shared Go library imported by all services in the go-web ecosystem. It provides the common plumbing — structured logging, request/response logging middleware, centralized error handling, HTTP response helpers, database transaction management, and inter-service HTTP communication — so each individual service only needs to implement its own business logic.

---

## Packages

| Package | Purpose |
|---|---|
| [`entrypoint`](#entrypoint) | Bootstrap function that wires platform middleware and routes into a Gin router |
| [`logger`](#logger) | Structured, colored console logger |
| [`middleware`](#middleware) | Logging and error-handling Gin middlewares |
| [`error`](#error) | All error types, DTOs, and pre-defined sentinel errors |
| [`transport/http`](#transporthttp) | `Ok` response helper and platform route registration |
| [`db/session`](#dbsession) | PostgreSQL transaction abstraction via context propagation |
| [`utils`](#utils) | `GetEnv` and `SendRequest` helpers |
| [`constants`](#constants) | Shared error codes, environment names, and header names |
| [`types`](#types) | Shared type definitions (`ErrorCode`, `Environment`, pagination, logging config) |

---

## entrypoint

**Import:** `github.com/Lomank123/go-web-platform/entrypoint`

### What it does

`SetupPlatform` is the single call every service makes in `main.go` to apply all platform-level setup to their Gin router:

1. Switches Gin to release mode when `env == "prod"`.
2. Registers a `TagNameFunc` on the validator so validation errors report JSON field names (`"email"`) instead of Go struct field names (`"Email"`).
3. Registers `gin.Recovery()` to convert panics into 500 responses.
4. Attaches `LoggingMiddleware` for structured request/response logging.
5. Attaches `ErrorHandlerMiddleware` for centralized error responses.
6. Registers `/health`, `/ready`, and (in non-production) `/swagger/*any` routes.

### Signature

```go
func SetupPlatform(
    router        *gin.Engine,
    log           logger.Logger,
    readinessFunc func() error,       // nil = always ready
    loggingConfig types.LoggingConfig,
    env           types.Environment,
)
```

### Usage example

```go
// cmd/app/main.go
import (
    platform     "github.com/Lomank123/go-web-platform/entrypoint"
    "github.com/Lomank123/go-web-platform/logger"
    platformMiddleware "github.com/Lomank123/go-web-platform/middleware"
)

func main() {
    cfg  := config.Load()
    logg := logger.NewLogger(cfg.App.Env)

    router := gin.New()

    platform.SetupPlatform(
        router,
        logg,
        db.Ping,                                    // readiness check
        platformMiddleware.DefaultLoggingConfig(),
        cfg.App.Env,
    )

    // Register application routes after SetupPlatform
    internalRouter.SetupRouter(router, ...)

    router.Run(":" + cfg.App.Port)
}
```

---

## logger

**Import:** `github.com/Lomank123/go-web-platform/logger`

### What it does

Provides a thin `Logger` interface and a `simpleLogger` implementation that writes timestamped, level-tagged lines to stdout.

- ANSI color per level (red for error, yellow for warn, etc.) when the terminal supports it or `NO_COLOR` is not set.
- `Debug` and `Warn` output is suppressed in `Production` environments.
- All other levels (`Info`, `Error`, `Fatal`) are always emitted.

### Interface

```go
type Logger interface {
    Debug(args ...any)
    Info(args ...any)
    Warn(args ...any)
    Error(args ...any)
    Fatal(args ...any)
}
```

### Usage example

```go
logg := logger.NewLogger(cfg.App.Env)

logg.Info("Starting server on port ", cfg.App.Port)
logg.Debug("Request payload: ", payload)
logg.Error("Failed to connect to DB: ", err)
logg.Fatal("Unrecoverable startup error: ", err)
```

---

## middleware

**Import:** `github.com/Lomank123/go-web-platform/middleware`

### LoggingMiddleware

Logs every HTTP request and its response in a structured format. Attached automatically by `SetupPlatform`.

Key behaviour:
- Extracts `X-Trace-ID` from the incoming request header; generates a UUID if absent and stores it in the Gin context.
- Skips logging for `/health`, `/ready`, and `/metrics` endpoints.
- Detects binary payloads (images, PDFs, videos) and multipart form data — skips body logging for those.
- Masks sensitive fields (`password`, `token`, `secret`, `key`, `authorization`) in both request and response JSON bodies.
- Truncates field values longer than `MaxFieldLength`.
- Supports `PrettyLog` mode (multi-line, indented JSON) or compact single-line mode.

#### Configuration — `types.LoggingConfig`

```go
type LoggingConfig struct {
    SensitiveFields []string  // additional fields to mask beyond the built-in list
    LogRequestBody  bool
    LogResponseBody bool
    MaxFieldLength  int
    PrettyLog       bool
}
```

`middleware.DefaultLoggingConfig()` returns a ready-to-use config with request/response body logging enabled, `PrettyLog: true`, and a sensible `MaxFieldLength`.

---

### ErrorHandlerMiddleware

Runs after all handlers via `c.Next()`. If `c.Errors` is non-empty it picks up the last error and converts it into a JSON response.

Error type priority:

1. **`*BaseError`** — uses the status code and error code the error carries itself.
2. **`validator.ValidationErrors`** — returns 400 with a `ValidationErrorDTO` listing each invalid field with its JSON name.
3. **Anything else** — returns 500.

#### Why this pattern

Handlers never write error responses themselves. They push errors onto the Gin context with `c.Error(err)` and return. The middleware handles rendering. Domain errors carry their own HTTP status via `BaseError.Status`, so no separate handler object or wiring change is ever needed.

#### Usage in handlers

```go
func (h *myHandler) CreateV1(c *gin.Context) {
    var payload dto.CreateInput
    if err := c.ShouldBindJSON(&payload); err != nil {
        _ = c.Error(err)   // ValidationErrors flow through automatically
        return
    }

    result, err := h.service.Create(c.Request.Context(), payload)
    if err != nil {
        _ = c.Error(err)   // any *BaseError is handled by middleware
        return
    }

    platformResponse.Ok(c, result)
}
```

---

## error

**Import:** `github.com/Lomank123/go-web-platform/error`

All error-related definitions live here: the `HTTPError` interface, error structs, constructors, pre-defined sentinel errors, and JSON response shapes.

### `BaseError`

General-purpose domain error. Implements `HTTPError`. `Status` defaults to 400 when zero.

```go
type BaseError struct {
    Code    types.ErrorCode
    Message string
    Status  int   // 0 → HTTP 400
}

func NewError(code types.ErrorCode, message string) error
func NewErrorWithStatus(code types.ErrorCode, message string, status int) error
```

### Pre-defined sentinel errors

```go
var (
    ErrEntityNotFound        // 404 ENTITY_NOT_FOUND
    ErrInvalidRequestPayload // 400 INVALID_REQUEST_PAYLOAD
    ErrUnauthorized          // 401 UNAUTHORIZED_ERROR
    ErrInternalServerError   // 500 INTERNAL_SERVER_ERROR
    ErrForbidden             // 403 FORBIDDEN_ERROR
    ErrValidation            // 400 VALIDATION_ERROR
)
```

### Response DTOs

```go
type ErrorDTO struct {
    ErrorCode string `json:"error_code"`
    Message   string `json:"message"`
}

type ValidationErrorDTO struct {
    ErrorDTO
    Errors []ValidationErrorItem `json:"errors"`
}

type ValidationErrorItem struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

### Usage examples

**Return a pre-defined sentinel:**
```go
if user == nil {
    return nil, platformError.ErrEntityNotFound  // middleware returns 404
}
```

**Create a one-off error:**
```go
return nil, platformError.NewErrorWithStatus("SESSION_EXPIRED", "session has expired", http.StatusUnauthorized)
```

**Define a reusable domain error:**
```go
// internal/error/errors.go
var ErrPlanLimitReached = platformError.NewErrorWithStatus(
    "PLAN_LIMIT_REACHED", "plan limit reached", http.StatusPaymentRequired,
)

// handler
_ = c.Error(internalError.ErrPlanLimitReached)
```

**Inspect an upstream error in gateway middleware:**
```go
var baseErr *platformError.BaseError
if errors.As(err, &baseErr) && baseErr.Code == constants.EntityNotFound {
    _ = c.Error(platformError.ErrEntityNotFound)  // re-map to local 404
    return
}
_ = c.Error(err)  // forward status + code as-is
```

---

## transport/http

**Import:** `github.com/Lomank123/go-web-platform/transport/http`

### `Ok`

The only response helper. Use it to write a successful 200 response.

```go
func Ok(c *gin.Context, data any)
```

All error responses are handled by `ErrorHandlerMiddleware` via `c.Error(err)`. There are no error helper functions — errors carry their own HTTP semantics through the `HTTPError` interface.

### Platform routes (`AddPlatformRoutes`)

Called by `SetupPlatform` — services do not call this directly. It registers:

- `GET /health` — always returns `{"status": "ready"}` (liveness probe).
- `GET /ready` — calls the `readinessFunc` passed to `SetupPlatform`; returns 503 if it errors (readiness probe).
- `GET /swagger/*any` — serves the Swagger UI; only registered in `local`, `dev`, and `stage` environments.

---

## db/session

**Import:** `github.com/Lomank123/go-web-platform/db/session`

### What it does

Abstracts PostgreSQL transaction management so services can run operations inside a transaction without passing `pgx.Tx` explicitly through every function call. The active transaction is stored in the `context.Context` and retrieved by repositories.

Both `Transaction` and `Begin` are safe to nest: if a transaction already exists in the context the inner call participates in the outer one without starting a new transaction. `Commit` and `Rollback` on an inner (nested) session are no-ops — only the outermost owner actually commits or rolls back.

### Interface

```go
type Session interface {
    Begin(ctx context.Context) (Session, error)
    Transaction(ctx context.Context, f func(context.Context) error) error
    Rollback() error
    Commit() error
    Context() context.Context
}
```

### Helper functions

```go
// Returns tx from context; creates one from pool if none exists.
func GetOrCreateTx(ctx context.Context, fallback *pgxpool.Pool) (pgx.Tx, error)

// Returns tx from context; returns nil if none exists (no side effects).
func GetTxFromContext(ctx context.Context) pgx.Tx
```

### Usage examples

**Initialize in `main.go`:**
```go
pgSession := session.NewPostgres(pgPool)
userService := service.NewUserService(pgSession, userRepo, ...)
```

**Automatic transaction — callback style:**
```go
func (s *userService) CreateWithProfile(ctx context.Context, ...) error {
    return s.session.Transaction(ctx, func(ctx context.Context) error {
        user, err := s.userRepo.Create(ctx, ...)
        if err != nil {
            return err   // auto-rollback
        }
        return s.profileRepo.Create(ctx, user.ID, ...)
        // auto-commit on nil return
    })
}
```

**Manual transaction — explicit begin/commit/rollback:**
```go
func (s *userService) CreateWithProfile(ctx context.Context, ...) error {
    sess, err := s.session.Begin(ctx)
    if err != nil {
        return err
    }
    defer sess.Rollback() // no-op if Commit was already called

    user, err := s.userRepo.Create(sess.Context(), ...)
    if err != nil {
        return err
    }
    if err = s.profileRepo.Create(sess.Context(), user.ID, ...); err != nil {
        return err
    }
    return sess.Commit()
}
```

**Nesting — inner call participates in the outer transaction:**
```go
func (s *userService) Outer(ctx context.Context) error {
    return s.session.Transaction(ctx, func(ctx context.Context) error {
        // Both calls share the same transaction.
        // The inner Transaction call does not start a new tx or commit.
        return s.inner(ctx)
    })
}

func (s *userService) Inner(ctx context.Context) error {
    return s.session.Transaction(ctx, func(ctx context.Context) error {
        return s.userRepo.Create(ctx, ...)
    })
}
```

**Repository layer — use transaction from context if present:**
```go
func (r *userRepository) Create(ctx context.Context, ...) (*domain.User, error) {
    tx := session.GetTxFromContext(ctx)
    if tx != nil {
        return scanUser(tx.QueryRow(ctx, query, ...))
    }
    return scanUser(r.db.QueryRow(ctx, query, ...))
}
```

---

## utils

**Import:** `github.com/Lomank123/go-web-platform/utils`

### `GetEnv`

Reads an environment variable and returns a fallback string if the variable is not set.

```go
func GetEnv(key, fallback string) string
```

### `SendRequest`

Makes an internal HTTP call from one service to another. Handles JSON marshalling, trace ID forwarding, and error response decoding.

```go
func SendRequest(
    method    string,
    url       string,
    payload   any,
    outputDTO any,
    context   *gin.Context,   // used only for trace ID propagation; may be nil
) error
```

On a non-2xx response it decodes the body into an `ErrorDTO` and returns a `*BaseError` carrying the upstream status code, error code, and message. On network or decode failure it returns a `*BaseError` with status 500.

```go
func (s *emailService) SendEmail(c *gin.Context, payload dto.EmailPayload) error {
    var out dto.SendEmailOutputDTO
    return platformUtils.SendRequest("POST", s.emailServiceURL+"/api/v1/send", payload, &out, c)
}

err := s.emailClient.SendEmail(c, payload)
if err != nil {
    _ = c.Error(err)   // *BaseError with upstream status + code forwarded by middleware
    return
}
```

---

## constants

**Import:** `github.com/Lomank123/go-web-platform/constants`

### Error codes

```go
const (
    EntityNotFound        types.ErrorCode = "ENTITY_NOT_FOUND"
    InvalidRequestPayload types.ErrorCode = "INVALID_REQUEST_PAYLOAD"
    ValidationError       types.ErrorCode = "VALIDATION_ERROR"
    UnauthorizedError     types.ErrorCode = "UNAUTHORIZED_ERROR"
    InternalServerError   types.ErrorCode = "INTERNAL_SERVER_ERROR"
    ForbiddenError        types.ErrorCode = "FORBIDDEN_ERROR"
)
```

### Environments

```go
const (
    Local       types.Environment = "local"
    Development types.Environment = "dev"
    Staging     types.Environment = "stage"
    Production  types.Environment = "prod"
)
```

### Headers

```go
const TraceIDHeader = "X-Trace-ID"
```

---

## types

**Import:** `github.com/Lomank123/go-web-platform/types`

Shared primitive type definitions used across multiple packages to avoid import cycles.

```go
type ErrorCode   string
type Environment string

type PaginationInputParams struct {
    Page  int64 `json:"page,omitempty"`
    Limit int64 `json:"limit,omitempty"`
}

type PaginationOutputParams struct {
    Page       int64 `json:"page"`
    TotalPages int64 `json:"total_pages"`
    PerPage    int64 `json:"per_page"`
    Total      int64 `json:"total"`
}

type LoggingConfig struct { ... }
```

---

## Error flow — end to end

```
[service repository]
    └─ returns platformError.ErrEntityNotFound  (404, self-describing)

[service handler]
    └─ c.Error(err)  →  Gin error list

[service ErrorHandlerMiddleware]
    └─ *BaseError matched  →  HTTP 404 {"error_code":"ENTITY_NOT_FOUND","message":"entity not found"}

[gateway utils.SendRequest]
    └─ resp.StatusCode == 404  →  NewErrorWithStatus("ENTITY_NOT_FOUND", "entity not found", 404)

[gateway handler]
    └─ c.Error(err)
    └─ *BaseError matched  →  HTTP 404 {"error_code":"ENTITY_NOT_FOUND","message":"entity not found"}
```

## Migration guide (from previous version)

| Before | After |
|---|---|
| `types.ErrorHandler` interface | No longer needed — errors carry their own HTTP status via `BaseError.Status` |
| `types.ErrorHandlerPayload` | Removed — no longer needed |
| `SetupPlatform(..., errorHandler, env)` | `SetupPlatform(..., env)` — no `errorHandler` argument |
| `platformResponse.NotFound(c, err)` | `c.Error(platformError.ErrEntityNotFound)` |
| `platformResponse.Unauthorized(c, err)` | `c.Error(platformError.ErrUnauthorized)` |
| `platformResponse.BadRequest(c, err, code)` | `c.Error(platformError.NewError(code, err.Error()))` |
| `platformResponse.Error(c, err, status, code)` | `c.Error(platformError.NewErrorWithStatus(code, err.Error(), status))` |
| `dto.ErrorDTO` | `platformError.ErrorDTO` |
| `dto.ValidationErrorDTO` | `platformError.ValidationErrorDTO` |
| Validation errors report `"field": "Email"` | Now report `"field": "email"` (json tag name) |
