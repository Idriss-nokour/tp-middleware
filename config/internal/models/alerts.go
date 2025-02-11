package models

import (
	"github.com/gofrs/uuid"
)

type Alerts struct {
	Id    *uuid.UUID `json:"id"`
	Email string     `json:"Email"`
	All   bool       `json:"All"`
}
