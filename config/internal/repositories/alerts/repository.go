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
		err := rows.Scan(&alert.Id, &alert.Email, &alert.All)
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

	row := db.QueryRow("SELECT id, email, all FROM alerts WHERE id=?", id.String())

	var alert models.Alerts
	err = row.Scan(&alert.Id, &alert.Email, &alert.All)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &alert, nil
}
