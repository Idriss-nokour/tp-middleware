package models

import (
	"github.com/gofrs/uuid"
)

type Ressources struct {
	Id    *uuid.UUID `json:"id"`
	UcaId int        `json:"UcaId"`
	Name  string     `json:"Name"`
}
