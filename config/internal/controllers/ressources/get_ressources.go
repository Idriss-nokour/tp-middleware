package ressources

import (
	"encoding/json"
	"net/http"
	"middleware/example/internal/models"
	"middleware/example/internal/services/ressources"
)

func GetRessources(w http.ResponseWriter, r *http.Request) {
	ressources, err := ressources.GetAllRessources()
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ressources)
}
