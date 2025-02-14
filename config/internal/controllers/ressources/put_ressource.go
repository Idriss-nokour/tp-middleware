package ressources

import (
	"encoding/json"
	"net/http"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	"middleware/example/internal/services/ressources"

)

func UpdateRessource(w http.ResponseWriter, r *http.Request) {
	ressourceId := r.Context().Value("ressourceId").(uuid.UUID)

	var ressource models.Ressources

	if err := json.NewDecoder(r.Body).Decode(&ressource); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	updatedRessource, err := ressources.UpdateRessource(ressourceId, ressource)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedRessource)
}
