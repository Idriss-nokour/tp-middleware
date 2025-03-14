package models

import (
)

type Alert struct {
    EventID   string `json:"event_id"`
    Message   string `json:"message"`
    Email     string `json:"email"`
    EventType string `json:"event_type"`
}
