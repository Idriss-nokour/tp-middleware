package alerts

import (
	"encoding/json"
	"net/http"
	//"github.com/sirupsen/logrus"
	"middleware/example/internal/models"
	"middleware/example/internal/services/alerts"
)

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := alerts.GetAllAlerts()
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alerts)
}
