# Migration Guide

This guide covers the breaking changes introduced in the error handling refactor and provides exact before/after replacements for each one.

---

## Breaking changes at a glance

| Area | What changed |
|---|---|
| `SetupPlatform` | Removed `errorHandler` parameter |
| `types.ErrorHandler` | Interface removed — no longer needed |
| `types.ErrorHandlerPayload` | Struct removed — no longer needed |
| `dto` package | Removed — types moved to the `error` package |
| `RequestError` / `NewRequestError` | Removed — `*BaseError` handles all cases |
| `constants.RequestError` | Constant removed |
| Response helpers | All error helpers removed — only `Ok` remains |
| Sentinel HTTP status codes | Now carry correct status codes (see details below) |
| Validation field names | Now report JSON tag names instead of Go struct field names |

---

## 1. `SetupPlatform` — remove the `errorHandler` argument

```go
// Before
platform.SetupPlatform(router, logg, db.Ping, cfg, &internalError.ErrorHandler{}, env)
platform.SetupPlatform(router, logg, nil, cfg, nil, env)

// After
platform.SetupPlatform(router, logg, db.Ping, cfg, env)
platform.SetupPlatform(router, logg, nil, cfg, env)
```

---

## 2. Replace `ErrorHandler` implementations with self-describing errors

Services that had a custom `ErrorHandler` to map domain errors to HTTP responses should delete it entirely. Instead, define each domain error using `NewErrorWithStatus` so the error carries its own HTTP status.

```go
// Before — internal/error/handler.go
type ErrorHandler struct{}

func (eh *ErrorHandler) Handle(err error) (platformTypes.ErrorHandlerPayload, bool) {
    var domainErr *MyDomainError
    if errors.As(err, &domainErr) {
        return platformTypes.ErrorHandlerPayload{
            StatusCode: http.StatusConflict,
            Payload: platformDTO.ErrorDTO{
                ErrorCode: string(domainErr.Code),
                Message:   domainErr.Message,
            },
        }, true
    }
    return platformTypes.ErrorHandlerPayload{}, false
}
```

```go
// After — delete handler.go, define errors in internal/error/errors.go
var ErrMyDomain = platformError.NewErrorWithStatus(
    "MY_DOMAIN_ERROR", "my domain error message", http.StatusConflict,
)

// Handler usage is unchanged:
_ = c.Error(internalError.ErrMyDomain)
```

---

## 3. Replace `dto` package imports

The `dto` package has been removed. Its types now live in the `error` package.

```go
// Before
import platformDTO "github.com/Lomank123/go-web-platform/dto"

platformDTO.ErrorDTO{...}
platformDTO.ValidationErrorDTO{...}
platformDTO.ValidationErrorItem{...}
```

```go
// After
import platformError "github.com/Lomank123/go-web-platform/error"

platformError.ErrorDTO{...}
platformError.ValidationErrorDTO{...}
platformError.ValidationErrorItem{...}
```

---

## 4. Replace `RequestError` checks

`RequestError` has been removed. `SendRequest` now returns a `*BaseError` carrying the upstream status code and error code. Inspect it the same way as any other `*BaseError`.

```go
// Before
var reqErr *platformError.RequestError
if errors.As(err, &reqErr) {
    if reqErr.StatusCode == http.StatusNotFound {
        platformResponse.NotFound(c, errors.New("resource not found"))
        return
    }
    _ = c.Error(platformError.ErrInternalServerError)
    return
}
```

```go
// After
var baseErr *platformError.BaseError
if errors.As(err, &baseErr) {
    if baseErr.Code == constants.EntityNotFound {
        _ = c.Error(platformError.ErrEntityNotFound)
        return
    }
    _ = c.Error(platformError.ErrInternalServerError)
    return
}
```

In most gateway cases you can skip the inspection entirely and just forward the error — the middleware will use the upstream status code directly:

```go
_ = c.Error(err)  // BaseError.Status carries the upstream status; middleware renders it
```

---

## 5. Replace response helper calls

All error response helpers have been removed. Replace each one with `c.Error` and the equivalent sentinel or constructor.

```go
// Before → After

platformResponse.Unauthorized(c, err)
// → _ = c.Error(platformError.ErrUnauthorized)

platformResponse.Forbidden(c, err)
// → _ = c.Error(platformError.ErrForbidden)

platformResponse.NotFound(c, err)
// → _ = c.Error(platformError.ErrEntityNotFound)

platformResponse.InternalServerError(c, err)
// → _ = c.Error(platformError.ErrInternalServerError)

platformResponse.BadRequest(c, err, myErrorCode)
// → _ = c.Error(platformError.NewError(myErrorCode, err.Error()))

platformResponse.Error(c, err, http.StatusConflict, myErrorCode)
// → _ = c.Error(platformError.NewErrorWithStatus(myErrorCode, err.Error(), http.StatusConflict))
```

`platformResponse.Ok` is unchanged.

---

## 6. Sentinel errors now return correct HTTP status codes

Previously all sentinels mapped to HTTP 400. They now carry their intended status. If your service was compensating by calling a response helper after checking the error type, that compensation is no longer needed — just push the sentinel and let the middleware handle it.

| Sentinel | Old status | New status |
|---|---|---|
| `ErrEntityNotFound` | 400 | **404** |
| `ErrUnauthorized` | 400 | **401** |
| `ErrForbidden` | 400 | **403** |
| `ErrInternalServerError` | 400 | **500** |
| `ErrInvalidRequestPayload` | 400 | 400 (unchanged) |
| `ErrValidation` | 400 | 400 (unchanged) |

Common pattern that can be simplified:

```go
// Before — manually mapping to the right status because sentinels were always 400
if err == platformError.ErrEntityNotFound {
    platformResponse.NotFound(c, err)
    return
}

// After — sentinel already carries 404; just push it
_ = c.Error(err)
```

---

## 7. Validation errors now report JSON field names

The validator is now configured to use `json` tag names in validation error responses. No code changes are needed on the server side, but **API clients** must be updated if they were matching on Go struct field names.

```json
// Before
{"error_code": "VALIDATION_ERROR", "message": "...", "errors": [
    {"field": "Email", "message": "invalid email"},
    {"field": "Password", "message": "minimum length is 8"}
]}

// After
{"error_code": "VALIDATION_ERROR", "message": "...", "errors": [
    {"field": "email", "message": "invalid email"},
    {"field": "password", "message": "minimum length is 8"}
]}
```

---

## Checklist

- [ ] Remove `errorHandler` argument from every `SetupPlatform` call
- [ ] Delete `internal/error/handler.go` (or equivalent) in each service
- [ ] Replace each domain error type with `NewError` / `NewErrorWithStatus` sentinel variables
- [ ] Replace `platformDTO.*` imports with `platformError.*`
- [ ] Replace `RequestError` / `NewRequestError` usages with `*BaseError` checks
- [ ] Replace all `platformResponse.Unauthorized/Forbidden/NotFound/InternalServerError/BadRequest/Error` calls with `c.Error(...)`
- [ ] Remove any status-code compensation logic around `ErrEntityNotFound`, `ErrUnauthorized`, and `ErrForbidden`
- [ ] Update API clients to use lowercase JSON field names in validation error responses
