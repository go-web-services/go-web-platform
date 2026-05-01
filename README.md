# go-web-platform

Shared Go library for the go-web ecosystem. Provides structured logging, request/response middleware, centralized error handling, HTTP helpers, database transaction management, and inter-service communication — so each service only implements its own business logic.

---

## Installation

```bash
go get github.com/go-web-services/go-web-platform
```

Requires Go 1.23+.

---

## Quick start

Call `SetupPlatform` once in `main.go` before registering your application routes:

```go
package main

import (
    platform   "github.com/go-web-services/go-web-platform/entrypoint"
    "github.com/go-web-services/go-web-platform/logger"
    mw         "github.com/go-web-services/go-web-platform/middleware"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg  := config.Load()
    logg := logger.NewLogger(cfg.App.Env)

    router := gin.New()

    platform.SetupPlatform(
        router,
        logg,
        db.Ping,                  // readiness check — nil = always ready
        mw.DefaultLoggingConfig(),
        cfg.App.Env,
    )

    // Register your application routes here
    api.SetupRoutes(router, ...)

    router.Run(":" + cfg.App.Port)
}
```

`SetupPlatform` handles everything else: Gin release mode, panic recovery, request/response logging, error handling middleware, and `/health` + `/ready` + `/swagger` routes.

---

## Error handling

All errors flow through the `ErrorHandlerMiddleware` via `c.Error(err)`. Handlers never write error responses themselves.

### Using pre-defined errors

```go
import platformError "github.com/go-web-services/go-web-platform/error"

// In a service or repository
if user == nil {
    return nil, platformError.ErrEntityNotFound  // renders as HTTP 404
}
```

| Sentinel | Status |
|---|---|
| `ErrEntityNotFound` | 404 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| `ErrInvalidRequestPayload` | 400 |
| `ErrValidation` | 400 |
| `ErrInternalServerError` | 500 |

### Defining domain errors

```go
// internal/error/errors.go
import (
    platformError "github.com/go-web-services/go-web-platform/error"
    "net/http"
)

var (
    ErrPlanLimitReached = platformError.NewErrorWithStatus(
        "PLAN_LIMIT_REACHED", "plan limit reached", http.StatusPaymentRequired,
    )
    ErrSessionExpired = platformError.NewErrorWithStatus(
        "SESSION_EXPIRED", "session has expired", http.StatusUnauthorized,
    )
)
```

### In handlers

```go
import (
    platformError    "github.com/go-web-services/go-web-platform/error"
    platformResponse "github.com/go-web-services/go-web-platform/transport/http"
)

func (h *handler) CreateV1(c *gin.Context) {
    var input dto.CreateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        _ = c.Error(err)  // validation errors are formatted automatically
        return
    }

    result, err := h.service.Create(c.Request.Context(), input)
    if err != nil {
        _ = c.Error(err)  // any *BaseError is handled by middleware
        return
    }

    platformResponse.Ok(c, result)
}
```

### Validation error response shape

Field names in validation errors match the `json` struct tag, not the Go field name:

```json
{
  "error_code": "VALIDATION_ERROR",
  "message": "validation error occurred",
  "errors": [
    {"field": "email", "message": "invalid email"},
    {"field": "password", "message": "minimum length is 8"}
  ]
}
```

---

## Inter-service communication

```go
import (
    platformError "github.com/go-web-services/go-web-platform/error"
    platformUtils  "github.com/go-web-services/go-web-platform/utils"
    "github.com/go-web-services/go-web-platform/constants"
)

func (s *client) GetUser(c *gin.Context, id string) (*dto.UserDTO, error) {
    var out dto.UserDTO
    err := platformUtils.SendRequest("GET", s.baseURL+"/users/"+id, nil, &out, c)
    return &out, err
}

// In the calling handler — inspect upstream error code if needed
user, err := s.userClient.GetUser(c, id)
if err != nil {
    var baseErr *platformError.BaseError
    if errors.As(err, &baseErr) && baseErr.Code == constants.EntityNotFound {
        _ = c.Error(platformError.ErrEntityNotFound)  // re-map to local 404
        return
    }
    _ = c.Error(err)  // forward upstream status as-is
    return
}
```

`SendRequest` automatically forwards the `X-Trace-ID` header for distributed tracing.

---

## Database transactions

```go
import "github.com/go-web-services/go-web-platform/db/session"

// main.go
pgSession := session.NewPostgres(pgPool)
userService := service.NewUserService(pgSession, userRepo, profileRepo)

// service layer — wrap multiple repo calls atomically
func (s *userService) CreateWithProfile(ctx context.Context, ...) error {
    return s.session.Transaction(ctx, func(ctx context.Context) error {
        user, err := s.userRepo.Create(ctx, ...)
        if err != nil {
            return err  // auto-rollback
        }
        return s.profileRepo.Create(ctx, user.ID, ...)
        // auto-commit on nil return
    })
}

// repository layer — picks up the transaction from context automatically
func (r *repo) Create(ctx context.Context, ...) (*domain.User, error) {
    tx := session.GetTxFromContext(ctx)
    if tx != nil {
        return scan(tx.QueryRow(ctx, query, args...))
    }
    return scan(r.pool.QueryRow(ctx, query, args...))
}
```

---

## Logging

```go
import "github.com/go-web-services/go-web-platform/logger"

logg := logger.NewLogger(cfg.App.Env)

logg.Info("server starting on port ", port)
logg.Debug("request payload: ", payload)  // suppressed in production
logg.Error("db connection failed: ", err)
logg.Fatal("unrecoverable error: ", err)
```

Pass `logg` down to components that need it. Do not use the standard `log` package.

---

## Packages

| Package | Import path |
|---|---|
| Bootstrap | `github.com/go-web-services/go-web-platform/entrypoint` |
| Logger | `github.com/go-web-services/go-web-platform/logger` |
| Middleware | `github.com/go-web-services/go-web-platform/middleware` |
| Errors + DTOs | `github.com/go-web-services/go-web-platform/error` |
| HTTP response | `github.com/go-web-services/go-web-platform/transport/http` |
| DB transactions | `github.com/go-web-services/go-web-platform/db/session` |
| Utils | `github.com/go-web-services/go-web-platform/utils` |
| Constants | `github.com/go-web-services/go-web-platform/constants` |
| Types | `github.com/go-web-services/go-web-platform/types` |

For full API reference and advanced usage see [docs/overview.md](docs/overview.md).  
For upgrading from a previous version see [docs/migration.md](docs/migration.md).
