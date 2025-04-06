package models

import (
    "time"
)

type Alert struct {
    EventID   string    `json:"event_id"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

