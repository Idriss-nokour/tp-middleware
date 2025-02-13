package alerts

import (
	"encoding/json"
	"net/http"
	"github.com/gofrs/uuid"
	//"github.com/sirupsen/logrus"
	"middleware/example/internal/models"
	"middleware/example/internal/services/alerts"


)

func GetAlert(w http.ResponseWriter, r *http.Request) {
	alertId := r.Context().Value("alertId").(uuid.UUID)

	alert, err := alerts.GetAlertByID(alertId)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}
