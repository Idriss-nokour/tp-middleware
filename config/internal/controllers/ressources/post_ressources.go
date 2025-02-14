package ressources

import (
	"encoding/json"
	"net/http"
	"middleware/example/internal/models"
	"middleware/example/internal/services/ressources"
)

func CreateRessource(w http.ResponseWriter, r *http.Request) {
	var ressource models.Ressources

	if err := json.NewDecoder(r.Body).Decode(&ressource); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	createdRessource, err := ressources.CreateRessource(ressource)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRessource)
}


