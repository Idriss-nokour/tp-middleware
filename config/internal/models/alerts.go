package models

import (
	"github.com/gofrs/uuid"
	//"time"
)

type Alerts struct {
	Id    *uuid.UUID `json:"id"`
	Email 	string     `json:"Email"`
	IsAll   bool       `json:"is_all"`
	RessourceID uuid.UUID  `json:"ressource_id"`
}
