package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	platformConstants "github.com/go-web-services/go-web-platform/constants"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/go-web-services/go-web-platform/types"
)

const defaultUpstreamRequestTimeout = 30 * time.Second

var upstreamHTTPClient = &http.Client{
	Timeout: defaultUpstreamRequestTimeout,
}

// GetEnv returns the value of the environment variable specified by key.
// If the variable is not set, it prints a message and returns the provided fallback value.
func GetEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		fmt.Println("Environment variable ", key, " not found, using fallback value: ", fallback)
		return fallback
	}
	return val
}

// SendRequest makes an HTTP call to an internal service and decodes the response.
// On a non-2xx response it returns a *BaseError carrying the upstream status code,
// error code, and message so callers can inspect or forward it.
func SendRequest(method, url string, payload any, outputDTO any, ginCtx *gin.Context) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return platformError.ErrInvalidRequestPayload
	}

	reqCtx := context.Background()
	if ginCtx != nil && ginCtx.Request != nil {
		reqCtx = ginCtx.Request.Context()
	}

	req, err := http.NewRequestWithContext(reqCtx, method, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if ginCtx != nil {
		if traceID := ginCtx.GetHeader(platformConstants.TraceIDHeader); traceID != "" {
			req.Header.Set(platformConstants.TraceIDHeader, traceID)
		}
	}

	resp, err := upstreamHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var errPayload platformError.ErrorDTO
		if err = json.NewDecoder(resp.Body).Decode(&errPayload); err != nil {
			return platformError.NewErrorWithStatus(
				platformConstants.InternalServerError,
				fmt.Sprintf("upstream request failed with status %d", resp.StatusCode),
				http.StatusInternalServerError,
			)
		}
		return platformError.NewErrorWithStatus(
			types.ErrorCode(errPayload.ErrorCode),
			errPayload.Message,
			resp.StatusCode,
		)
	}

	if err = json.NewDecoder(resp.Body).Decode(outputDTO); err != nil {
		return platformError.NewErrorWithStatus(
			platformConstants.InternalServerError,
			"upstream request succeeded but response could not be decoded",
			http.StatusInternalServerError,
		)
	}

	return nil
}
