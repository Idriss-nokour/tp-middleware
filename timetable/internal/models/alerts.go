package models

import (
    "time"
)

type Alert struct {
    EventID   string    `json:"event_id"`
    Message   string    `json:"message"`
    Email     string    `json:"email"`
    EventType string    `json:"event_type"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

