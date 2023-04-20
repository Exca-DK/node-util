package utils

import "github.com/google/uuid"

func init() {
	uuid.EnableRandPool()
}

type UUID struct {
	uuid.UUID
}

func NewId() UUID {
	return UUID{uuid.New()}
}
