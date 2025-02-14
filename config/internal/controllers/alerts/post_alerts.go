package alerts

import (
	"encoding/json"
	"net/http"
	"middleware/example/internal/models"
	"middleware/example/internal/services/alerts"
)

func CreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert models.Alerts

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	createdAlert, err := alerts.CreateAlert(alert)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAlert)
}
