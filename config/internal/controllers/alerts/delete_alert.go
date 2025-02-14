package alerts

import (
	"net/http"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	"middleware/example/internal/services/alerts"
)

func DeleteAlert(w http.ResponseWriter, r *http.Request) {
	alertId := r.Context().Value("alertId").(uuid.UUID)

	err := alerts.DeleteAlert(alertId)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusNoContent) 
}
