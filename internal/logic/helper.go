package logic

import (
	"context"
	"fmt"
)

// GetUserIDFromContext extracts the authenticated user ID from context.
// go-zero JWT middleware injects claims into context with their original types.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value("userId")
	if v == nil {
		return "", fmt.Errorf("unauthorized: missing user context")
	}

	// Try string first
	if userId, ok := v.(string); ok {
		if userId == "" {
			return "", fmt.Errorf("unauthorized: empty user id")
		}
		return userId, nil
	}

	// If not string, convert to string
	userId := fmt.Sprintf("%v", v)
	if userId == "" {
		return "", fmt.Errorf("unauthorized: empty user id")
	}
	return userId, nil
}
