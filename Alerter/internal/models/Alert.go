package models

import (
    "github.com/google/uuid"
	"time"
)

type Alert struct {
	EventID        uuid.UUID `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

