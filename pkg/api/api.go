// Package api provides shared API helpers for straitKubegateway,
// including common response types, error codes, and request context.
package api

import (
	"context"
	"fmt"
	"time"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// CommandResponse is the standard response for mutation commands.
type CommandResponse struct {
	CommandID string `json:"commandId"`
	Status    string `json:"status"`
	TraceID   string `json:"traceId"`
}

// NewCommandResponse creates a new accepted command response.
func NewCommandResponse(commandID, traceID string) CommandResponse {
	return CommandResponse{
		CommandID: commandID,
		Status:    "accepted",
		TraceID:   traceID,
	}
}

// ListResponse is a generic paginated list response.
type ListResponse[T any] struct {
	Items      []T    `json:"items"`
	TotalCount int64  `json:"totalCount"`
	Cursor     string `json:"cursor,omitempty"`
	PageSize   int    `json:"pageSize"`
}

// ListRequest is a generic paginated list request.
type ListRequest struct {
	Cursor   string `json:"cursor,omitempty"`
	PageSize int    `json:"pageSize"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"`
	Filter   string `json:"filter,omitempty"`
}

// ErrorCode represents a machine-readable error code.
type ErrorCode string

const (
	ErrorCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	ErrorCodeInvalidArgument   ErrorCode = "INVALID_ARGUMENT"
	ErrorCodePermissionDenied  ErrorCode = "PERMISSION_DENIED"
	ErrorCodeUnauthenticated   ErrorCode = "UNAUTHENTICATED"
	ErrorCodeInternal          ErrorCode = "INTERNAL"
	ErrorCodeUnavailable       ErrorCode = "UNAVAILABLE"
	ErrorCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	ErrorCodeConflict          ErrorCode = "CONFLICT"
)

// APIError is a structured error returned by the API.
type APIError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Retryable bool      `json:"retryable"`
	TraceID   string    `json:"traceId,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAPIError creates a new API error.
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// contextKey is an unexported type for context keys.
type contextKey int

const (
	keyTraceID contextKey = iota
	keyRequestID
	keyMetadata
)

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, keyTraceID, traceID)
}

// TraceIDFromContext extracts the trace ID from the context.
func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(keyTraceID).(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, keyRequestID, requestID)
}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(keyRequestID).(string); ok {
		return id
	}
	return ""
}

// WithMetadata adds observability metadata to the context.
func WithMetadata(ctx context.Context, md *types.Metadata) context.Context {
	return context.WithValue(ctx, keyMetadata, md)
}

// MetadataFromContext extracts observability metadata from the context.
func MetadataFromContext(ctx context.Context) *types.Metadata {
	if md, ok := ctx.Value(keyMetadata).(*types.Metadata); ok {
		return md
	}
	return &types.Metadata{Timestamp: time.Now().UTC()}
}
