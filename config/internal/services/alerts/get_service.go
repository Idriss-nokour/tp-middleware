package alerts

import (
	"github.com/sirupsen/logrus"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	repository "middleware/example/internal/repositories/alerts"
	"net/http"
)

	


func GetAllAlerts() ([]models.Alerts, error) {
	alerts, err := repository.GetAllAlerts()
	if err != nil {
		logrus.Errorf("error retrieving alerts: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Something went wrong while retrieving alerts",
			Code:    http.StatusInternalServerError,
		}
	}

	return alerts, nil
}

func GetAlertByID(id uuid.UUID) (*models.Alerts, error) {
	alert, err := repository.GetAlertByID(id)
	if err != nil {
		logrus.Errorf("error retrieving alert by ID: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Something went wrong while retrieving the alert",
			Code:    http.StatusInternalServerError,
		}
	}
	if alert == nil {
		return nil, &models.CustomError{
			Message: "Alert not found",
			Code:    http.StatusNotFound,
		}
	}

	return alert, nil
}


func CreateAlert(alert models.Alerts) (*models.Alerts, error) {
	createdAlert, err := repository.InsertAlert(alert)
	if err != nil {
		logrus.Errorf("Error creating alert: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Error while creating alert",
			Code:    http.StatusInternalServerError,
		}
	}
	return createdAlert, nil
}


func UpdateAlert(id uuid.UUID, alert models.Alerts) (*models.Alerts, error) {
	updatedAlert, err := repository.UpdateAlert(id, alert)
	if err != nil {
		logrus.Errorf("Error updating alert: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Error while updating alert",
			Code:    http.StatusInternalServerError,
		}
	}
	return updatedAlert, nil
}




func DeleteAlert(id uuid.UUID) error {
	err := repository.DeleteAlert(id)
	if err != nil {
		logrus.Errorf("Error deleting alert: %s", err.Error())
		return &models.CustomError{
			Message: "Error while deleting alert",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}
