package models

import "github.com/gofrs/uuid"

type Ressources struct {
	Id    *uuid.UUID `json:"id"`    
	UcaId int        `json:"uca_id"` 
	Name  string     `json:"name"`  
}
