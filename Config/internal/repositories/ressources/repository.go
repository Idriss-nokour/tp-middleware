package ressources

import (
	"github.com/gofrs/uuid"
	"middleware/example/internal/helpers"
	"middleware/example/internal/models"
	"database/sql"
)

func GetAllRessources() ([]models.Ressources, error) {
	// Ouvrir la base de données
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	rows, err := db.Query("SELECT * FROM ressources")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ressources := []models.Ressources{}

	for rows.Next() {
		var ressource models.Ressources
		err := rows.Scan(&ressource.Id, &ressource.UcaId, &ressource.Name)
		if err != nil {
			return nil, err
		}
		ressources = append(ressources, ressource)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ressources, nil
}

func GetRessourceByID(id uuid.UUID) (*models.Ressources, error) {
	// Ouvrir la base de données
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	row := db.QueryRow("SELECT id, uca_id, name FROM ressources WHERE id=?", id.String())

	var ressource models.Ressources
	err = row.Scan(&ressource.Id, &ressource.UcaId, &ressource.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &ressource, nil
}





func InsertRessource(ressource models.Ressources) (*models.Ressources, error) {
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)
	
	newUUID := uuid.Must(uuid.NewV4())  
	ressource.Id = &newUUID 

	_, err = db.Exec("INSERT INTO ressources (id, uca_id, name) VALUES (?, ?, ?)", 
		ressource.Id.String(), ressource.UcaId, ressource.Name)
	if err != nil {
		return nil, err
	}

	return &ressource, nil
}


func UpdateRessource(id uuid.UUID, resource models.Ressources) (*models.Ressources, error) {
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	_, err = db.Exec("UPDATE ressources SET uca_id = ?, name = ? WHERE id = ?", 
		resource.UcaId, resource.Name, id.String())
	if err != nil {
		return nil, err
	}

	resource.Id = &id
	return &resource, nil
}


func DeleteRessource(id uuid.UUID) error {
	db, err := helpers.OpenDB()
	if err != nil {
		return err
	}
	defer helpers.CloseDB(db)

	_, err = db.Exec("DELETE FROM ressources WHERE id = ?", id.String())
	if err != nil {
		return err
	}

	return nil
}

