package alerts

import (
	"github.com/sirupsen/logrus"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	repository "middleware/example/internal/repositories/alerts"
	"net/http"
)

	


// GetAllAlerts récupère toutes les alertes
func GetAllAlerts() ([]models.Alerts, error) {
	// Appeler la fonction du repository pour obtenir toutes les alertes
	alerts, err := repository.GetAllAlerts()
	if err != nil {
		logrus.Errorf("error retrieving alerts: %s", err.Error())
		// Retourner une erreur personnalisée
		return nil, &models.CustomError{
			Message: "Something went wrong while retrieving alerts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Retourner les alertes récupérées
	return alerts, nil
}

// GetAlertByID récupère une alerte spécifique par son ID
func GetAlertByID(id uuid.UUID) (*models.Alerts, error) {
	// Appeler la fonction du repository pour obtenir une alerte spécifique par ID
	alert, err := repository.GetAlertByID(id)
	if err != nil {
		logrus.Errorf("error retrieving alert by ID: %s", err.Error())
		// Retourner une erreur personnalisée
		return nil, &models.CustomError{
			Message: "Something went wrong while retrieving the alert",
			Code:    http.StatusInternalServerError,
		}
	}
	// Vérifier si l'alerte existe
	if alert == nil {
		// Si l'alerte n'est pas trouvée, retourner une erreur
		return nil, &models.CustomError{
			Message: "Alert not found",
			Code:    http.StatusNotFound,
		}
	}

	// Retourner l'alerte
	return alert, nil
}
