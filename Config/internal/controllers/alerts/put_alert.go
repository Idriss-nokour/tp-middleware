package alerts

import (
	"encoding/json"
	"net/http"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	"middleware/example/internal/services/alerts"
)

func UpdateAlert(w http.ResponseWriter, r *http.Request) {
	alertId := r.Context().Value("alertId").(uuid.UUID)

	var alert models.Alerts

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	updatedAlert, err := alerts.UpdateAlert(alertId, alert)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedAlert)
}
