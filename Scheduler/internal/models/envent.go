package models

import (
	//"github.com/gofrs/uuid"
	"time"
)

type Event struct {
	Uid             string    `json:"uid"`
	Description     string    `json:"description"`
	Localisation    string    `json:"localisation"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	LastModificated time.Time `json:"lastmodificated"`
	Type            string    `json:"type"`
}

