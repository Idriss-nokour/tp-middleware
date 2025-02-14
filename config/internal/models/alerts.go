package models

import (
	"github.com/gofrs/uuid"
)

type Alerts struct {
	Id    *uuid.UUID `json:"id"`
	Email string     `json:"Email"`
	IsAll   bool       `json:"is_all"`
}
