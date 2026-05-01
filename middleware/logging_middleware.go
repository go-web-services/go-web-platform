package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-web-services/go-web-platform/constants"
	"github.com/go-web-services/go-web-platform/logger"
	"github.com/go-web-services/go-web-platform/types"
	"github.com/google/uuid"
)

// DefaultLoggingConfig returns a default configuration for the logging middleware
func DefaultLoggingConfig() types.LoggingConfig {
	return types.LoggingConfig{
		SensitiveFields: []string{
			"password",
			"token",
			"secret",
			"key",
			"authorization",
		},
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxFieldLength:  1000,
		PrettyLog:       false,
	}
}

// LoggingMiddleware creates a middleware that logs all incoming and outgoing requests
// Assumes all requests and responses are JSON payloads without checking content type
func LoggingMiddleware(log logger.Logger, config types.LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Extract or generate trace ID
		traceID := extractOrGenerateTraceID(c)

		// Store trace ID in request and response headers
		c.Request.Header.Set(constants.TraceIDHeader, traceID)
		c.Writer.Header().Set(constants.TraceIDHeader, traceID)

		// Special case: /metrics, /swagger/*, /health, and /ready endpoints
		if (c.Request.Method == "GET" && c.Request.URL.Path == "/metrics") ||
			(c.Request.Method == "GET" && strings.HasPrefix(c.Request.URL.Path, "/swagger/")) ||
			(c.Request.Method == "GET" && c.Request.URL.Path == "/health") ||
			(c.Request.Method == "GET" && c.Request.URL.Path == "/ready") {
			c.Next()
			return
		}

		// Log request
		requestBody := ""
		var bodyBytes []byte
		if config.LogRequestBody && c.Request.Body != nil {
			// Check if this is a multipart/form-data request
			contentType := c.GetHeader("Content-Type")
			isMultipart := strings.Contains(contentType, "multipart/form-data")

			if !isMultipart {
				var err error
				bodyBytes, err = io.ReadAll(c.Request.Body)
				if err != nil {
					log.Warn("Failed to read request body",
						"traceID", traceID,
						"error", err.Error(),
					)
				}

				// Create a new reader for logging
				bodyReader := io.NopCloser(bytes.NewBuffer(bodyBytes))
				requestBody = readAndMaskJSONBody(bodyReader, log, config, traceID, contentType)

				// Restore the original body with the raw bytes
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				// For multipart requests, just log that it's a multipart request
				requestBody = "[multipart/form-data request - body not logged]"
			}
		}

		// Format request headers
		requestHeaders := formatHeaders(c.Request.Header, config)

		// Log incoming request
		query := truncateString(c.Request.URL.RawQuery, config.MaxFieldLength)
		if config.PrettyLog {
			log.Info("Incoming request:",
				"\ntraceID:", traceID,
				"\nmethod:", c.Request.Method,
				"\npath:", c.Request.URL.Path,
				"\nquery:", query,
				"\nheaders:", requestHeaders,
				"\nbody:", requestBody,
			)
		} else {
			log.Info(fmt.Sprintf(
				"Incoming request traceID=%s method=%s path=%s query=%s headers=%s body=%s",
				traceID, c.Request.Method, c.Request.URL.Path, oneLineForLog(query), oneLineForLog(requestHeaders), oneLineForLog(requestBody),
			))
		}

		// Create a custom response writer to capture the response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Format response headers
		responseHeaders := formatHeaders(writer.Header(), config)

		// Get response body if available
		responseBody := ""
		if config.LogResponseBody && writer.body.Len() > 0 {
			// Create a reader from the captured response body
			bodyReader := io.NopCloser(bytes.NewReader(writer.body.Bytes()))
			contentType := writer.Header().Get("Content-Type")
			responseBody = readAndMaskJSONBody(bodyReader, log, config, traceID, contentType)
		}

		// Log outgoing response
		if config.PrettyLog {
			log.Info("Outgoing response:",
				"\ntraceID:", traceID,
				"\nmethod:", c.Request.Method,
				"\npath:", c.Request.URL.Path,
				"\nstatus:", c.Writer.Status(),
				"\nduration:", duration.String(),
				"\nheaders:", responseHeaders,
				"\nbody:", responseBody,
			)
		} else {
			log.Info(fmt.Sprintf(
				"Outgoing response traceID=%s method=%s path=%s status=%d duration=%s headers=%s body=%s",
				traceID, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration.String(),
				oneLineForLog(responseHeaders), oneLineForLog(responseBody),
			))
		}
	}
}

// oneLineForLog replaces newline and carriage-return runes with spaces so one log record
// stays one physical line (e.g. for Docker). It does not use strings.Fields, which would
// break JSON and other values that contain intentional spaces.
func oneLineForLog(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\n', '\r':
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// truncateString truncates a string if it exceeds maxLength
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "...(truncated)"
}

// formatHeaders formats HTTP headers in a readable way
func formatHeaders(headers http.Header, config types.LoggingConfig) string {
	if len(headers) == 0 {
		return "no headers"
	}

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if !config.PrettyLog {
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			value := strings.Join(headers[key], ", ")
			if isSensitiveField(key, config.SensitiveFields) {
				value = "********"
			} else {
				value = truncateString(value, config.MaxFieldLength)
			}
			parts = append(parts, fmt.Sprintf("%s: %s", key, value))
		}
		return strings.Join(parts, "; ")
	}

	var sb strings.Builder
	for _, key := range keys {
		values := headers[key]
		value := strings.Join(values, ", ")
		if isSensitiveField(key, config.SensitiveFields) {
			value = "********"
		} else {
			value = truncateString(value, config.MaxFieldLength)
		}
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	}

	return sb.String()
}

// extractOrGenerateTraceID extracts the trace ID from request headers or generates a new one
func extractOrGenerateTraceID(c *gin.Context) string {
	// Try to get trace ID from the request header
	traceID := c.GetHeader(constants.TraceIDHeader)

	// If no trace ID was found, generate a new one
	if traceID == "" {
		traceID = uuid.New().String()
	}

	return traceID
}

// readAndMaskJSONBody reads a JSON body from an io.ReadCloser, masks sensitive data,
// and returns the formatted JSON string
func readAndMaskJSONBody(body io.ReadCloser, log logger.Logger, config types.LoggingConfig, traceID string, contentType string) string {
	if body == nil {
		return ""
	}

	if isBinaryType(contentType) {
		return "<file_content>"
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		log.Warn("Failed to read body",
			"traceID", traceID,
			"error", err.Error(),
		)
		return ""
	}

	if len(bodyBytes) == 0 {
		return ""
	}

	var jsonData any
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		log.Warn("Failed to parse body as JSON - body may be empty or malformed",
			"traceID", traceID,
			"error", err.Error(),
		)
		raw := truncateString(string(bodyBytes), config.MaxFieldLength)
		if !config.PrettyLog {
			return oneLineForLog(raw)
		}
		return raw
	}

	// Mask sensitive data and truncate long values
	maskedJSON := maskSensitiveData(jsonData, config)

	var formattedJSON []byte
	if config.PrettyLog {
		formattedJSON, err = json.MarshalIndent(maskedJSON, "", "  ")
	} else {
		formattedJSON, err = json.Marshal(maskedJSON)
	}
	if err != nil {
		log.Warn("Failed to format JSON",
			"traceID", traceID,
			"error", err.Error(),
		)
		raw := truncateString(string(bodyBytes), config.MaxFieldLength)
		if !config.PrettyLog {
			return oneLineForLog(raw)
		}
		return raw
	}

	out := string(formattedJSON)
	if !config.PrettyLog {
		return oneLineForLog(out)
	}
	return out
}

// responseWriter is a custom ResponseWriter that captures the response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// maskSensitiveData masks sensitive fields in the given data
func maskSensitiveData(data any, config types.LoggingConfig) any {
	switch v := data.(type) {
	case map[string]any:
		masked := make(map[string]any)
		for key, value := range v {
			if isSensitiveField(key, config.SensitiveFields) {
				masked[key] = "********"
			} else {
				switch val := value.(type) {
				case string:
					masked[key] = truncateString(val, config.MaxFieldLength)
				case map[string]any:
					masked[key] = maskSensitiveData(val, config)
				case []any:
					masked[key] = maskSensitiveData(val, config)
				default:
					masked[key] = value
				}
			}
		}
		return masked

	case []any:
		masked := make([]any, len(v))
		for i, value := range v {
			switch val := value.(type) {
			case string:
				masked[i] = truncateString(val, config.MaxFieldLength)
			case map[string]any:
				masked[i] = maskSensitiveData(val, config)
			case []any:
				masked[i] = maskSensitiveData(val, config)
			default:
				masked[i] = value
			}
		}
		return masked

	case string:
		return truncateString(v, config.MaxFieldLength)

	default:
		return data
	}
}

// isSensitiveField checks if a field name contains any sensitive keywords
func isSensitiveField(field string, sensitiveFields []string) bool {
	field = strings.ToLower(field)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(field, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// isBinaryType checks if the content type indicates a binary file
func isBinaryType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "video/") ||
		strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "application/pdf") ||
		strings.HasPrefix(contentType, "application/octet-stream") ||
		strings.HasPrefix(contentType, "application/zip") ||
		strings.HasPrefix(contentType, "application/x-gzip")
}
