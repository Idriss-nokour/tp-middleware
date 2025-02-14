package alerts

import (
	"github.com/gofrs/uuid"
	"middleware/example/internal/helpers"
	"middleware/example/internal/models"
	"database/sql"
)

func GetAllAlerts() ([]models.Alerts, error) {
	// Ouvrir la base de données
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	rows, err := db.Query("SELECT * FROM alerts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []models.Alerts{}

	for rows.Next() {
		var alert models.Alerts
		err := rows.Scan(&alert.Id, &alert.Email, &alert.IsAll)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func GetAlertByID(id uuid.UUID) (*models.Alerts, error) {
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	row := db.QueryRow("SELECT id, email, is_all, ressource_id FROM alerts WHERE id=?", id.String())

	var alert models.Alerts
	err = row.Scan(&alert.Id, &alert.Email, &alert.IsAll)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &alert, nil
}

func InsertAlert(alert models.Alerts) (*models.Alerts, error) {
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	newUUID := uuid.Must(uuid.NewV4())  
	alert.Id = &newUUID 


	_, err = db.Exec("INSERT INTO alerts (id, email, is_all, ressource_id) VALUES (?, ?, ?, ?)", alert.Id.String(), alert.Email, alert.IsAll, alert.RessourceID.String())
	if err != nil {
		return nil, err
	}

	return &alert, nil
}



func UpdateAlert(id uuid.UUID, alert models.Alerts) (*models.Alerts, error) {
	db, err := helpers.OpenDB()
	if err != nil {
		return nil, err
	}
	defer helpers.CloseDB(db)

	_, err = db.Exec("UPDATE alerts SET email = ?, is_all = ?, ressource_id = ? WHERE id = ?", alert.Email, alert.IsAll, id.String(), alert.RessourceID.String())
	if err != nil {
		return nil, err
	}

	alert.Id = &id
	return &alert, nil
}



func DeleteAlert(id uuid.UUID) error {
	db, err := helpers.OpenDB()
	if err != nil {
		return err
	}
	defer helpers.CloseDB(db)
	_, err = db.Exec("DELETE FROM alerts WHERE id = ?", id.String())
	if err != nil {
		return err
	}

	return nil
}
