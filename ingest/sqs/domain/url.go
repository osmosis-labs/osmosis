package domain

import (
	"context"
	"net/url"

	"github.com/labstack/echo"
)

// RequestPathKeyType is a custom type for request path key.
type RequestPathKeyType string

const (
	// RequestPathCtxKey is the key used to store the request path in the request context
	RequestPathCtxKey RequestPathKeyType = "request_path"
)

// ParseURLPath parses the URL path from the echo context
func ParseURLPath(c echo.Context) (string, error) {
	parsedURL, err := url.Parse(c.Request().RequestURI)
	if err != nil {
		return "", err
	}

	return parsedURL.Path, nil
}

// GetURLPathFromContext returns the request path from the context
func GetURLPathFromContext(ctx context.Context) (string, error) {
	// Get request path for metrics
	requestPath, ok := ctx.Value(RequestPathCtxKey).(string)
	if !ok || (ok && len(requestPath) == 0) {
		requestPath = "unknown"
	}
	return requestPath, nil
}
