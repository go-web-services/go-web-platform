package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	platformConstants "github.com/Lomank123/go-web-platform/constants"
	platformError "github.com/Lomank123/go-web-platform/error"
)

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

// SendRequest make HTTP request and decode the response body
// Used for internal services communication only.
func SendRequest(method, url string, payload any, outputDTO any, context *gin.Context) error {

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return platformError.ErrInvalidRequestPayload
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	// Set default content type
	req.Header.Set("Content-Type", "application/json")

	// Setting custom headers
	traceID := context.GetHeader(platformConstants.TraceIDHeader)
	headers := map[string]string{
		platformConstants.TraceIDHeader: traceID,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If request was not successful
	if resp.StatusCode >= http.StatusBadRequest {
		var errPayload map[string]any
		err = json.NewDecoder(resp.Body).Decode(&errPayload)
		if err != nil {
			return platformError.NewRequestError(http.StatusInternalServerError, map[string]any{
				"message": fmt.Sprintf("Request failed to internal service with status code %d", resp.StatusCode),
			})
		}

		return platformError.NewRequestError(resp.StatusCode, errPayload)
	}

	err = json.NewDecoder(resp.Body).Decode(outputDTO)
	if err != nil {
		return platformError.NewRequestError(http.StatusInternalServerError, map[string]any{
			"message": "Request was successful but response body could not be decoded",
		})
	}

	return nil
}
