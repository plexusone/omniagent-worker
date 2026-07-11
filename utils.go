package worker

import (
	"github.com/google/uuid"
)

// generateID generates a unique ID for requests.
func generateID() string {
	return uuid.New().String()
}
